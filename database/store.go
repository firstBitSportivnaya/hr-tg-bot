package database

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/IT-Nick/tasks"
)

// GlobalStore может быть установлен в main для доступа из middleware.
var GlobalStore Store

// UserState представляет данные, связанные с пользователем.
type UserState struct {
	Role              string       `json:"role"`                // Роль: "user", "hr", "admin"
	State             string       `json:"state"`               // Текущее состояние: "welcome", "assigned", "testing", "finished"
	CurrentQuestion   int          `json:"current_question"`    // Индекс текущего вопроса
	Score             int          `json:"score"`               // Количество правильных ответов
	TestTasks         []tasks.Task `json:"test_tasks"`          // Набор вопросов для теста
	Answers           map[int]int  `json:"answers"`             // Выбранные ответы: ключ – индекс вопроса, значение – индекс варианта
	TelegramUsername  string       `json:"telegram_username"`   // Ник пользователя
	TelegramFirstName string       `json:"telegram_first_name"` // Имя пользователя
	// AssignedBy – ID HR или admin, который назначил тест кандидату (в виде строки).
	AssignedBy string `json:"assigned_by"`
}

// Store определяет интерфейс для работы с состоянием.
type Store interface {
	Get(userID int64) (UserState, bool)
	Set(userID int64, state UserState) error
	Delete(userID int64) error
}

// MemoryStore — in‑memory реализация.
type MemoryStore struct {
	data map[int64]UserState
	mu   sync.RWMutex
}

// NewMemoryStore создаёт новый MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{data: make(map[int64]UserState)}
}

func (m *MemoryStore) Get(userID int64) (UserState, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	state, ok := m.data[userID]
	return state, ok
}

func (m *MemoryStore) Set(userID int64, state UserState) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[userID] = state
	return nil
}

func (m *MemoryStore) Delete(userID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, userID)
	return nil
}

// JSONStore — реализация, сохраняющая данные в JSON-файл.
type JSONStore struct {
	filename string
	mu       sync.Mutex
}

// NewJSONStore создаёт новый JSONStore с указанным файлом.
func NewJSONStore(filename string) *JSONStore {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		initial := make(map[int64]UserState)
		data, _ := json.Marshal(initial)
		_ = os.WriteFile(filename, data, 0644)
	}
	return &JSONStore{filename: filename}
}

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

func (j *JSONStore) Get(userID int64) (UserState, bool) {
	m, err := j.load()
	if err != nil {
		return UserState{}, false
	}
	state, ok := m[userID]
	return state, ok
}

func (j *JSONStore) Set(userID int64, state UserState) error {
	m, err := j.load()
	if err != nil {
		return err
	}
	m[userID] = state
	return j.save(m)
}

func (j *JSONStore) Delete(userID int64) error {
	m, err := j.load()
	if err != nil {
		return err
	}
	delete(m, userID)
	return j.save(m)
}

// NewStore возвращает реализацию Store в зависимости от типа хранения.
func NewStore(storageType, filename string) Store {
	if storageType == "json" {
		return NewJSONStore(filename)
	}
	return NewMemoryStore()
}
