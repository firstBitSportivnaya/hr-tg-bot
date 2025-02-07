/*
MIT License

Copyright (c) 2025 Первый Бит

Данная лицензия разрешает использование, копирование, изменение, слияние, публикацию, распространение,
лицензирование и/или продажу копий программного обеспечения при соблюдении следующих условий:

В вышеуказанном уведомлении об авторских правах и данном уведомлении о разрешении должны быть включены все копии
или значимые части программного обеспечения.

ПРОГРАММНОЕ ОБЕСПЕЧЕНИЕ ПРЕДОСТАВЛЯЕТСЯ "КАК ЕСТЬ", БЕЗ ГАРАНТИЙ ЛЮБОГО РОДА, ЯВНЫХ ИЛИ ПОДРАЗУМЕВАЕМЫХ,
ВКЛЮЧАЯ, НО НЕ ОГРАНИЧИВАЯСЬ, ГАРАНТИЯМИ КОММЕРЧЕСКОЙ ПРИГОДНОСТИ, СООТВЕТСТВИЯ ДЛЯ ОПРЕДЕЛЕННОЙ ЦЕЛИ И
НЕНАРУШЕНИЯ ПРАВ. НИ В КОЕМ СЛУЧАЕ АВТОРЫ ИЛИ ПРАВООБЛАДАТЕЛИ НЕ НЕСУТ ОТВЕТСТВЕННОСТИ ПО ИСКАМ,
УСЛОВИЯМ, ДАМГЕ или другим обязательствам, возникающим из, или в связи с использованием, или иным образом
связанным с данным программным обеспечением.
*/

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
func (m *Manager) GetRandomTasks(n int, candidateID string) ([]Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var free []int     // Индексы вопросов, которые не зарезервированы.
	var reserved []int // Индексы вопросов, зарезервированных для других кандидатов.

	// Разделяем вопросы на свободные и зарезервированные.
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
		// Если свободных вопросов достаточно, выбираем случайные n индексов.
		selected = randomIndices(free, n)
	} else {
		// Если свободных вопросов недостаточно, выбираем все свободные и добавляем случайные из зарезервированных.
		selected = append(selected, free...)
		remaining := n - len(free)
		if len(reserved) >= remaining {
			additional := randomIndices(reserved, remaining)
			selected = append(selected, additional...)
		} else {
			// Если и зарезервированных недостаточно, добавляем их все.
			selected = append(selected, reserved...)
		}
	}

	var result []Task
	// Резервируем выбранные вопросы за кандидатом, если они ещё не были зарезервированы.
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
