package tasks

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"sync"
	"time"
)

// Task описывает один тестовый вопрос.
type Task struct {
	ID         int      `json:"id"`
	Text       string   `json:"text"`
	Options    []string `json:"options"`
	Answer     int      `json:"answer"`
	ReservedBy string   `json:"-"`
}

// Manager управляет списком задач и их резервированием.
type Manager struct {
	tasks []Task
	mu    sync.Mutex
}

// NewManager загружает задачи из указанного файла.
func NewManager(filename string) (*Manager, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать файл с вопросами: %v", err)
	}
	var ts []Task
	if err := json.Unmarshal(data, &ts); err != nil {
		return nil, fmt.Errorf("не удалось разобрать JSON: %v", err)
	}
	return &Manager{tasks: ts}, nil
}

// GetRandomTasks выбирает случайный набор из n задач для кандидата с candidateID.
func (m *Manager) GetRandomTasks(n int, candidateID string) ([]Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var free []int
	var reserved []int
	for i, t := range m.tasks {
		if t.ReservedBy == "" {
			free = append(free, i)
		} else if t.ReservedBy != candidateID {
			reserved = append(reserved, i)
		}
	}

	var selected []int
	rand.Seed(time.Now().UnixNano())
	if len(free) >= n {
		selected = randomIndices(free, n)
	} else {
		selected = append(selected, free...)
		remaining := n - len(free)
		if len(reserved) >= remaining {
			additional := randomIndices(reserved, remaining)
			selected = append(selected, additional...)
		} else {
			selected = append(selected, reserved...)
		}
	}

	var result []Task
	for _, idx := range selected {
		task := m.tasks[idx]
		if task.ReservedBy == "" {
			m.tasks[idx].ReservedBy = candidateID
		}
		result = append(result, m.tasks[idx])
	}
	return result, nil
}

// ReleaseCandidateTasks освобождает задачи, зарезервированные для кандидата.
func (m *Manager) ReleaseCandidateTasks(candidateID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.tasks {
		if m.tasks[i].ReservedBy == candidateID {
			m.tasks[i].ReservedBy = ""
		}
	}
}

// randomIndices выбирает случайные индексы из среза.
func randomIndices(indices []int, count int) []int {
	cpy := make([]int, len(indices))
	copy(cpy, indices)
	rand.Shuffle(len(cpy), func(i, j int) {
		cpy[i], cpy[j] = cpy[j], cpy[i]
	})
	if count > len(cpy) {
		count = len(cpy)
	}
	return cpy[:count]
}
