package json

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// TestAssignment хранит данные о назначении теста кандидату.
// Структура содержит идентификаторы кандидата, информацию о том, кто назначил тест,
// а также дату и время назначения.
type TestAssignment struct {
	CandidateID       int64     `json:"candidate_id"`       // ID кандидата; изначально может быть 0, если тест назначается по username.
	CandidateUsername string    `json:"candidate_username"` // Имя пользователя кандидата (без символа "@").
	AssignedByID      int64     `json:"assigned_by_id"`     // ID пользователя, назначившего тест.
	AssignedBy        string    `json:"assigned_by"`        // Имя пользователя, назначившего тест.
	AssignedAt        time.Time `json:"assigned_at"`        // Время, когда тест был назначен.
	TestType          string    `json:"test_type"`          // Тип теста

}

// TestAssignmentStore представляет хранилище для отложенных назначений теста.
// Ключом в данном хранилище является строка, как правило, username кандидата.
type TestAssignmentStore struct {
	filename string     // Путь к JSON-файлу, где сохраняются назначения тестов.
	mu       sync.Mutex // Mutex для обеспечения потокобезопасного доступа к файлу.
}

// NewTestAssignmentStore создаёт новое хранилище для тестовых назначений.
// Если указанный файл не существует, он будет создан с пустой структурой.
func NewTestAssignmentStore(filename string) *TestAssignmentStore {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// Инициализируем пустое хранилище.
		initial := make(map[string]TestAssignment)
		data, _ := json.Marshal(initial)
		_ = os.WriteFile(filename, data, 0644)
	}
	return &TestAssignmentStore{filename: filename}
}

// load считывает данные из JSON-файла и десериализует их в map[string]TestAssignment.
// Возвращает полученную карту или ошибку, если чтение или десериализация не удались.
func (s *TestAssignmentStore) load() (map[string]TestAssignment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", s.filename, err)
	}
	if len(data) == 0 {
		return make(map[string]TestAssignment), nil
	}
	var m map[string]TestAssignment
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}
	return m, nil
}

// save сериализует переданную карту назначений в JSON и записывает её в файл.
// Возвращает ошибку, если сериализация или запись не удались.
func (s *TestAssignmentStore) save(m map[string]TestAssignment) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}
	if err := os.WriteFile(s.filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %v", s.filename, err)
	}
	return nil
}

// Get возвращает назначение теста по заданному ключу (обычно, username кандидата).
// Если назначение найдено, возвращается значение, true и nil; иначе - zero значение, false и nil.
func (s *TestAssignmentStore) Get(id string) (TestAssignment, bool, error) {
	m, err := s.load()
	if err != nil {
		return TestAssignment{}, false, err
	}
	assignment, ok := m[id]
	return assignment, ok, nil
}

// Set сохраняет или обновляет назначение теста в хранилище для заданного ключа (username кандидата).
// Возвращает ошибку, если операция не удалась.
func (s *TestAssignmentStore) Set(id string, assignment TestAssignment) error {
	m, err := s.load()
	if err != nil {
		return err
	}
	m[id] = assignment
	return s.save(m)
}

// Delete удаляет назначение теста по заданному ключу (username кандидата) из хранилища.
// Возвращает ошибку, если операция удаления не удалась.
func (s *TestAssignmentStore) Delete(id string) error {
	m, err := s.load()
	if err != nil {
		return err
	}
	delete(m, id)
	return s.save(m)
}

// RoleAssignment хранит данные о назначении новой роли кандидату.
// Структура используется для отложенных назначений ролей (например, назначение HR).
type RoleAssignment struct {
	CandidateUsername string    `json:"candidate_username"` // Username кандидата (без "@").
	NewRole           string    `json:"new_role"`           // Новая роль, которую необходимо назначить (например, "hr").
	AssignedBy        string    `json:"assigned_by"`        // Имя пользователя, осуществившего назначение.
	AssignedAt        time.Time `json:"assigned_at"`        // Время, когда роль была назначена.
}

// RoleAssignmentStore представляет хранилище для отложенных назначений ролей.
// Ключом в данном хранилище является username кандидата.
type RoleAssignmentStore struct {
	filename string     // Путь к JSON-файлу, где сохраняются назначения ролей.
	mu       sync.Mutex // Mutex для обеспечения потокобезопасного доступа к файлу.
}

// NewRoleAssignmentStore создаёт новое хранилище для назначений ролей.
// Если указанный файл не существует, он будет создан с пустой структурой.
func NewRoleAssignmentStore(filename string) *RoleAssignmentStore {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		initial := make(map[string]RoleAssignment)
		data, _ := json.Marshal(initial)
		_ = os.WriteFile(filename, data, 0644)
	}
	return &RoleAssignmentStore{filename: filename}
}

// load считывает данные из JSON-файла и десериализует их в map[string]RoleAssignment.
// Возвращает полученную карту или ошибку, если операция не удалась.
func (s *RoleAssignmentStore) load() (map[string]RoleAssignment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", s.filename, err)
	}
	if len(data) == 0 {
		return make(map[string]RoleAssignment), nil
	}
	var m map[string]RoleAssignment
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}
	return m, nil
}

// save сериализует карту назначений ролей в JSON и записывает её в файл.
// Возвращает ошибку, если сериализация или запись не удались.
func (s *RoleAssignmentStore) save(m map[string]RoleAssignment) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}
	if err := os.WriteFile(s.filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %v", s.filename, err)
	}
	return nil
}

// Get возвращает назначение роли для заданного username.
// Если назначение найдено, возвращает значение, true и nil; иначе – zero значение, false и nil.
func (s *RoleAssignmentStore) Get(username string) (RoleAssignment, bool, error) {
	m, err := s.load()
	if err != nil {
		return RoleAssignment{}, false, err
	}
	assignment, ok := m[username]
	return assignment, ok, nil
}

// Set сохраняет или обновляет назначение роли для заданного username.
// Возвращает ошибку, если операция не удалась.
func (s *RoleAssignmentStore) Set(username string, assignment RoleAssignment) error {
	m, err := s.load()
	if err != nil {
		return err
	}
	m[username] = assignment
	return s.save(m)
}

// Delete удаляет назначение роли для заданного username из хранилища.
// Возвращает ошибку, если операция не удалась.
func (s *RoleAssignmentStore) Delete(username string) error {
	m, err := s.load()
	if err != nil {
		return err
	}
	delete(m, username)
	return s.save(m)
}
