package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"popug_tasktracker/kafka"
	"popug_tasktracker/model"
	"popug_tasktracker/schemaReg"
)

const postgresDNS = "host=localhost user=base password=secret dbname=task port=5443 sslmode=disable"

type app struct {
	Db        *gorm.DB
	Kafka     *kafka.Kafkaapp
	SchemaReg *schemaReg.SchemaRegistry
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", "http://localhost:9096/", nil)
		if err != nil {
			fmt.Println("Ошибка создания запроса:", err)
			return
		}

		tokenCookie, err := r.Cookie("token")
		if err != nil {
			fmt.Println("Ошибка получения токена из cookie:", err)
			http.Redirect(w, r, "http://localhost:9096/", http.StatusSeeOther)
			return
		}
		req.Header.Add("Authorization", "Bearer "+tokenCookie.Value)

		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Ошибка выполнения запроса:", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			fmt.Println("Необходима авторизация. Перенаправление на страницу входа...")
			http.Redirect(w, r, "http://localhost:9096/", http.StatusSeeOther)
		}

		next.ServeHTTP(w, r)
	})
}
func setupRoutes(a *app) *mux.Router {
	r := mux.NewRouter()

	// Создаем подроутер для роутов, требующих аутентификации
	authRoutes := r.PathPrefix("/").Subrouter()
	authRoutes.Use(authMiddleware)
	authRoutes.HandleFunc("/", a.tasksHandler).Methods("GET")
	authRoutes.HandleFunc("/newtask", a.newTaskHandler).Methods("POST")

	// Регистрируем роуты, требующие аутентификации, используя authRoutes
	authRoutes.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tasks := []model.Task{}
		result := a.Db.Find(&tasks)
		if result.Error != nil {
			http.Error(w, "Ошибка при получении списка задач из базы данных", http.StatusInternalServerError)
			return
		}
		tmpl, err := template.ParseFiles("static/taskList.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, tasks); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		buf.WriteTo(w)
	}).Methods("GET")

	// Здесь можно добавить другие роуты, которые не требуют аутентификации

	return r
}

func main() {
	port := os.Getenv("TASK_PORT")
	if port == "" {
		port = "9097"
	}

	db, err := gorm.Open(postgres.Open(postgresDNS), &gorm.Config{})
	if err != nil {
		log.Fatalf("err to open DB %v \n", err)
	}
	err = db.AutoMigrate(&model.Task{}, &model.User{})
	if err != nil {
		log.Fatalf("err DB.AutoMigrate %v\n", err)
	}

	kafkaInstance := kafka.NewKafka()
	schemaRegistryInstance := schemaReg.NewSchemaRegistry()
	app := &app{
		Db:        db,
		Kafka:     kafkaInstance,          // Предполагается, что вы создали экземпляр Kafka
		SchemaReg: schemaRegistryInstance, //
	}

	r := setupRoutes(app)

	srv := &http.Server{
		Addr: "localhost:" + port,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}

	//eventsFromKafka := make(chan string, 10)
	//go app.Kafka.ConsumeKafkaEvent(eventsFromKafka)
	//	go app.kafkaEventProducer()
	//go eventHandler(eventsFromKafka)

	fmt.Println("Сервер задач запущен на порту ", port)
	srv.ListenAndServe()
	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c
}

// Функция для обработки запросов на получение информации о задачах
func (a *app) tasksHandler(w http.ResponseWriter, r *http.Request) {
	tasks := []model.Task{}
	result := a.Db.Find(&tasks)
	if result.Error != nil {
		http.Error(w, "Ошибка при получении списка задач из базы данных", http.StatusInternalServerError)
		return
	}
	tmpl, err := template.ParseFiles("static/taskList.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, tasks); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	buf.WriteTo(w)
}

// Функция для роутинга /newTask и создания новой задачи с использованием данных из формы
func (a *app) newTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка при разборе данных формы", http.StatusBadRequest)
		return
	}

	var newTask model.Task
	newTask.Description = r.FormValue("description")
	newTask.Assignee = a.getRandomDeveloperHandler()
	newTask.ID = uuid.New().String()

	result := a.Db.Create(&newTask)
	if result.Error != nil {
		http.Error(w, "Ошибка при сохранении задачи в базу данных", http.StatusInternalServerError)
		return
	}
}

func (a *app) getRandomDeveloperHandler() (string, error) {
	var developer model.User
	result := a.Db.Where("role = ?", "developer").Order("RANDOM()").First(&developer)
	if result.Error != nil {
		return "", result.Error
	}

	if developer.PublicId == "" {
		return "", fmt.Errorf("fail to find random developer")
	}
	return developer.PublicId, nil
}

// func (a *app) kafkaEventProducer() {
// 	for {
// 		item, err := generateRandomItem()
// 		if err != nil {
// 			fmt.Printf("Ошибка генерации или валидации события: %v\n", err)
// 			continue
// 		}

// 		err = a.SchemaReg.Validate("items", "created", item.EventVersion, item)
// 		if err != nil {
// 			panic(err)
// 		}
// 		itemString, err := json.Marshal(item)
// 		if err != nil {
// 			fmt.Printf("Ошибка преобразования msg в строку: %v\n", err)
// 			continue
// 		}

// 		err = a.Kafka.ProduceKafkaEvent(string(itemString))
// 		if err != nil {
// 			fmt.Printf("ошибка отправки  %v \n", err)
// 			continue
// 		}
// 		fmt.Printf("сообщение отправленно: 		%v \n", string(itemString))
// 		time.Sleep(time.Second * 1)
// 	}
// }

func eventHandler(eventsFromKafka chan string) {
	for {
		msg := <-eventsFromKafka
		var user model.User
		fmt.Printf("Получили сообщение %v\n", msg)
		err := json.Unmarshal([]byte(msg), &user)
		if err != nil {
			fmt.Printf("Ошибка десериализации сообщения: %v\n", err)
			continue
		}
		fmt.Printf("Получили новго пользователя: %v\n", user)
	}
}

// // Функция для генерации случайного события создания задачи, соответствующего схеме
// func generateRandomItem() (model.Item, error) {
// 	item := model.Item{
// 		EventID:      generateRandomID(),
// 		EventVersion: int(1),
// 		EventName:    "ItemsCreated",
// 		EventTime:    time.Now().Format(time.RFC3339),
// 		Producer:     "taskTracker",
// 		Data: model.ItemData{
// 			PublicID:    generateRandomID(),
// 			Title:       "Задача " + generateRandomID(),
// 			Description: generateRandomDescription(),
// 		},
// 	}

// 	return item, nil
// }
