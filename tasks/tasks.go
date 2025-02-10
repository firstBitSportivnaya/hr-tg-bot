package tasks

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"
)

// Task описывает один тестовый вопрос, включающий текст вопроса, варианты ответов,
// индекс правильного ответа и информацию о резервировании вопроса для конкретного кандидата.
// Поле ReservedBy используется для того, чтобы "зарезервировать" вопрос для кандидата и не допустить его повторного использования.
type Task struct {
	ID         int      `json:"id"`      // Уникальный идентификатор вопроса.
	Text       string   `json:"text"`    // Текст вопроса.
	Options    []string `json:"options"` // Список вариантов ответов.
	Answer     int      `json:"answer"`  // Индекс правильного ответа в срезе Options.
	Type       string   `json:"type"`    // Тип вопроса (например, "logic", "math" и т.д.)
	ReservedBy string   `json:"-"`       // Идентификатор кандидата, зарезервировавшего вопрос (не сериализуется в JSON).
}

// Manager управляет списком тестовых вопросов и их резервированием для кандидатов.
// Обеспечивает потокобезопасный доступ к списку вопросов с использованием мьютекса.
type Manager struct {
	tasks []Task     // Список всех тестовых вопросов.
	mu    sync.Mutex // Мьютекс для синхронизации доступа к списку вопросов.
}

// NewManager загружает список тестовых вопросов из указанного JSON-файла и возвращает нового менеджера.
// Если чтение файла или десериализация JSON не удались, возвращается ошибка.
func NewManager(filename string) (*Manager, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать файл с вопросами: %v", err)
	}
	var ts []Task
	if err := json.Unmarshal(data, &ts); err != nil {
		return nil, fmt.Errorf("не удалось разобрать JSON: %v", err)
	}
	return &Manager{tasks: ts}, nil
}

// GetRandomTasks выбирает случайный набор из n вопросов для кандидата с заданным candidateID.
// Функция учитывает резервирование вопросов:
//   - Если вопрос не зарезервирован, он добавляется в список свободных вопросов.
//   - Если вопрос зарезервирован для другого кандидата, он добавляется в список зарезервированных.
//
// В зависимости от количества свободных вопросов, выбирается случайный набор вопросов, и если вопрос не зарезервирован,
// он резервируется за кандидатом.
func (m *Manager) GetRandomTasks(n int, candidateID, testType string) ([]Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var free []int     // Индексы свободных вопросов.
	var reserved []int // Индексы зарезервированных вопросов.

	// Проходим по всем вопросам и отбираем только те, у которых совпадает тип (если testType задан).
	for i, t := range m.tasks {
		// Если задан тип и он не совпадает – пропускаем вопрос.
		if testType != "" && t.Type != testType {
			continue
		}
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

// ReleaseCandidateTasks освобождает все вопросы, ранее зарезервированные для кандидата с заданным candidateID.
// Эта функция вызывается, после завершения теста, чтобы вопросы могли быть использованы повторно.
func (m *Manager) ReleaseCandidateTasks(candidateID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.tasks {
		if m.tasks[i].ReservedBy == candidateID {
			m.tasks[i].ReservedBy = ""
		}
	}
}

// randomIndices выбирает случайные индексы из переданного среза индексов.
// Функция возвращает новый срез, содержащий count случайных элементов из исходного среза.
func randomIndices(indices []int, count int) []int {
	cpy := make([]int, len(indices))
	copy(cpy, indices)
	// Перемешиваем копию среза случайным образом.
	rand.Shuffle(len(cpy), func(i, j int) {
		cpy[i], cpy[j] = cpy[j], cpy[i]
	})
	if count > len(cpy) {
		count = len(cpy)
	}
	return cpy[:count]
}
