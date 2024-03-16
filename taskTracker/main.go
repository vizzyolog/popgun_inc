package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	kafkaapp "popug_inc/kafkaApp"
	"popug_inc/model"
	"popug_inc/schemaReg"
)

// Функция для обработки запросов на получение информации о задачах
func tasksHandler(w http.ResponseWriter, r *http.Request) {
	// Здесь должна быть логика определения роли пользователя и выборка задач в соответствии с этой ролью
	fmt.Fprintf(w, "Здесь будет информация задачах")
}

func eventHandler(eventsFromKafka chan string) {
	for {
		msg := <-eventsFromKafka
		fmt.Printf("Получено сообщение:     	%v\n", msg)
	}
}

func main() {
	http.HandleFunc("/tasks", tasksHandler)

	kafka := kafkaapp.NewKafka()
	schemaReg := schemaReg.NewSchemaRegistry()

	go func() {
		for {
			item, err := generateRandomItem()
			if err != nil {
				fmt.Printf("Ошибка генерации или валидации события: %v\n", err)
				continue
			}

			err = schemaReg.Validate("items", "created", "2", item)
			if err != nil {
				panic(err)
			}
			itemString, err := json.Marshal(item)
			if err != nil {
				fmt.Printf("Ошибка преобразования msg в строку: %v\n", err)
				continue
			}

			err = kafka.ProduceKafkaEvent(string(itemString))
			if err != nil {
				fmt.Printf("ошибка отправки  %v \n", err)
				continue
			}
			fmt.Printf("сообщение отправленно: 		%v \n", string(itemString))
			time.Sleep(time.Second * 1)
		}
	}()

	eventsFromKafka := make(chan string, 10)
	go kafka.ConsumeKafkaEvent(eventsFromKafka)

	go eventHandler(eventsFromKafka)

	fmt.Println("Сервер задач запущен на порту 8080")
	http.ListenAndServe(":8080", nil)

	var wg sync.WaitGroup

	wg.Wait() // Ожидаем завершения всех горутин
}

// Функция для генерации случайного идентификатора
func generateRandomID() string {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("%x", seededRand.Int63())
}

// Функция для генерации случайного описания (с шансом вернуть null)
func generateRandomDescription() *string {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	if seededRand.Intn(2) == 0 { // 50% шанс вернуть null
		return nil
	}
	description := "Описание " + generateRandomID()
	return &description
}

// Функция для генерации случайного события создания задачи, соответствующего схеме
func generateRandomItem() (model.Item, error) {
	item := model.Item{
		EventID:      generateRandomID(),
		EventVersion: int(2),
		EventName:    "ItemsCreated",
		EventTime:    time.Now().Format(time.RFC3339),
		Producer:     "taskTracker",
		Data: model.ItemData{
			PublicID:    generateRandomID(),
			Title:       "Задача " + generateRandomID(),
			Description: generateRandomDescription(),
		},
	}

	return item, nil
}
