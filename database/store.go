package database

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/IT-Nick/tasks"
)

// GlobalStore представляет глобальное хранилище состояния пользователей.
// Оно может быть установлено в main для обеспечения доступа к состояниям из middleware и других модулей.
var GlobalStore Store

// UserState описывает состояние пользователя в системе.
// Содержит информацию о роли, состоянии тестирования, текущем прогрессе, а также данные, необходимые для работы таймера.
type UserState struct {
	Role              string       `json:"role"`                // Роль пользователя (например, "user", "hr", "admin")
	State             string       `json:"state"`               // Текущее состояние (например, "welcome", "testing", "finished")
	CurrentQuestion   int          `json:"current_question"`    // Индекс текущего вопроса теста
	Score             int          `json:"score"`               // Количество правильных ответов
	TestTasks         []tasks.Task `json:"test_tasks"`          // Список тестовых вопросов, назначенных пользователю
	Answers           map[int]int  `json:"answers"`             // Карта, где ключ – индекс вопроса, значение – выбранный вариант ответа
	TelegramUsername  string       `json:"telegram_username"`   // Telegram-username пользователя
	TelegramFirstName string       `json:"telegram_first_name"` // Имя пользователя в Telegram
	AssignedBy        string       `json:"assigned_by"`         // Имя пользователя, назначившего тест
	AssignedByID      int64        `json:"assigned_by_id"`      // ID пользователя, назначившего тест
	TestType          string       `json:"test_type"`           // Вид теста
	// Поля для работы с таймером теста.
	TimerMessageID    int       `json:"timer_message_id"`    // ID сообщения, содержащего таймер
	TimerDeadline     time.Time `json:"timer_deadline"`      // Время, до которого пользователь должен завершить тест
	QuestionMessageID int       `json:"question_message_id"` // (Опционально) ID сообщения с текущим вопросом
}

// Store определяет интерфейс для работы с состояниями пользователей.
// Позволяет получать, сохранять и удалять состояния по ID пользователя.
type Store interface {
	Get(userID int64) (UserState, bool)
	Set(userID int64, state UserState) error
	Delete(userID int64) error
}

// MemoryStore представляет реализацию Store с хранением данных в оперативной памяти.
type MemoryStore struct {
	data map[int64]UserState // Карта состояний пользователей, индексированная по userID.
	mu   sync.RWMutex        // RWMutex обеспечивает потокобезопасный доступ к данным.
}

// NewMemoryStore создает и возвращает новый экземпляр MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{data: make(map[int64]UserState)}
}

// Get возвращает состояние пользователя по его ID.
// Если состояние найдено, возвращается состояние и true, иначе – false.
func (m *MemoryStore) Get(userID int64) (UserState, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	state, ok := m.data[userID]
	return state, ok
}

// Set сохраняет или обновляет состояние пользователя по его ID.
func (m *MemoryStore) Set(userID int64, state UserState) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[userID] = state
	return nil
}

// Delete удаляет состояние пользователя по его ID.
func (m *MemoryStore) Delete(userID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, userID)
	return nil
}

// JSONStore представляет реализацию Store с сохранением данных в JSON-файл.
type JSONStore struct {
	filename string     // Путь к JSON-файлу, где хранятся состояния.
	mu       sync.Mutex // Mutex обеспечивает эксклюзивный доступ при чтении/записи файла.
}

// NewJSONStore создает новый экземпляр JSONStore.
// Если указанный файл не существует, он создается с пустой начальной структурой.
func NewJSONStore(filename string) *JSONStore {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		initial := make(map[int64]UserState)
		data, _ := json.Marshal(initial)
		_ = os.WriteFile(filename, data, 0644)
	}
	return &JSONStore{filename: filename}
}

// load считывает и десериализует данные из JSON-файла, содержащего состояния пользователей.
func (j *JSONStore) load() (map[int64]UserState, error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	data, err := os.ReadFile(j.filename)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать файл %s: %v", j.filename, err)
	}
	if len(data) == 0 {
		return make(map[int64]UserState), nil
	}
	var m map[int64]UserState
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("не удалось разобрать JSON: %v", err)
	}
	return m, nil
}

// save сериализует данные и сохраняет их в JSON-файл.
// Функция возвращает ошибку, если не удалось выполнить сериализацию или запись в файл.
func (j *JSONStore) save(m map[int64]UserState) error {
	j.mu.Lock()
	defer j.mu.Unlock()
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("не удалось сериализовать данные: %v", err)
	}
	if err := os.WriteFile(j.filename, data, 0644); err != nil {
		return fmt.Errorf("не удалось записать файл %s: %v", j.filename, err)
	}
	return nil
}

// Get возвращает состояние пользователя по его ID из JSON-файла.
// Если произошла ошибка при загрузке или состояние не найдено, возвращается false.
func (j *JSONStore) Get(userID int64) (UserState, bool) {
	m, err := j.load()
	if err != nil {
		return UserState{}, false
	}
	state, ok := m[userID]
	return state, ok
}

// Set сохраняет или обновляет состояние пользователя в JSON-файле.
func (j *JSONStore) Set(userID int64, state UserState) error {
	m, err := j.load()
	if err != nil {
		return err
	}
	m[userID] = state
	return j.save(m)
}

// Delete удаляет состояние пользователя из JSON-файла по его ID.
func (j *JSONStore) Delete(userID int64) error {
	m, err := j.load()
	if err != nil {
		return err
	}
	delete(m, userID)
	return j.save(m)
}

// LoadAllStates загружает и возвращает все состояния пользователей из JSON-файла.
func (j *JSONStore) LoadAllStates() (map[int64]UserState, error) {
	return j.load()
}

// NewStore возвращает реализацию интерфейса Store в зависимости от указанного типа хранения.
// Если storageType равен "json", возвращается JSONStore, иначе используется in‑memory реализация (MemoryStore).
func NewStore(storageType, filename string) Store {
	if storageType == "json" {
		return NewJSONStore(filename)
	}
	return NewMemoryStore()
}
