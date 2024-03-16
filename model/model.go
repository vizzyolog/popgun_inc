package model

// Структура для хранения информации о задаче
type Task struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Assignee    string `json:"assignee"`
	Status      string `json:"status"`
}

// Структура для хранения информации о пользователе
type User struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

// Структура ItemData для представления данных события
type ItemData struct {
	PublicID    string  `json:"public_id"`
	Title       string  `json:"title"`
	Description *string `json:"description"` // Указатель на строку, чтобы поддерживать null
}

// Структура Item для представления события создания элемента
type Item struct {
	EventID      string   `json:"event_id"`
	EventVersion int      `json:"event_version"`
	EventName    string   `json:"event_name"`
	EventTime    string   `json:"event_time"`
	Producer     string   `json:"producer"`
	Data         ItemData `json:"data"`
}

// Структура Account для представления события создания аккаунта
type Account struct {
	EventID      string      `json:"event_id"`
	EventVersion int         `json:"event_version"`
	EventName    string      `json:"event_name"`
	EventTime    string      `json:"event_time"`
	Producer     string      `json:"producer"`
	Data         AccountData `json:"data"`
}

type AccountData struct {
	PublicID  string  `json:"public_id"`
	Email     string  `json:"email"`
	FirstName *string `json:"first_name"` // Указатель на строку, чтобы поддерживать null
	LastName  *string `json:"last_name"`  // Указатель на строку, чтобы поддерживать null
	Position  *string `json:"position"`   // Указатель на строку, чтобы поддерживать null
}
