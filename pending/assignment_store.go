package pending

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// TestAssignment хранит данные о назначении теста кандидату.
type TestAssignment struct {
	CandidateID       int64     `json:"candidate_id"`
	CandidateUsername string    `json:"candidate_username"`
	AssignedBy        string    `json:"assigned_by"`
	AssignedAt        time.Time `json:"assigned_at"`
}

// TestAssignmentStore – хранилище для отложенных назначений теста, ключом является строка (CandidateID).
type TestAssignmentStore struct {
	filename string
	mu       sync.Mutex
}

// NewTestAssignmentStore создаёт новое хранилище для тестовых назначений.
func NewTestAssignmentStore(filename string) *TestAssignmentStore {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		initial := make(map[string]TestAssignment)
		data, _ := json.Marshal(initial)
		_ = os.WriteFile(filename, data, 0644)
	}
	return &TestAssignmentStore{filename: filename}
}

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

func (s *TestAssignmentStore) Get(id string) (TestAssignment, bool, error) {
	m, err := s.load()
	if err != nil {
		return TestAssignment{}, false, err
	}
	assignment, ok := m[id]
	return assignment, ok, nil
}

func (s *TestAssignmentStore) Set(id string, assignment TestAssignment) error {
	m, err := s.load()
	if err != nil {
		return err
	}
	m[id] = assignment
	return s.save(m)
}

func (s *TestAssignmentStore) Delete(id string) error {
	m, err := s.load()
	if err != nil {
		return err
	}
	delete(m, id)
	return s.save(m)
}

// RoleAssignment и RoleAssignmentStore оставляем без изменений (они работают по username)
type RoleAssignment struct {
	CandidateUsername string    `json:"candidate_username"`
	NewRole           string    `json:"new_role"`
	AssignedBy        string    `json:"assigned_by"`
	AssignedAt        time.Time `json:"assigned_at"`
}

type RoleAssignmentStore struct {
	filename string
	mu       sync.Mutex
}

func NewRoleAssignmentStore(filename string) *RoleAssignmentStore {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		initial := make(map[string]RoleAssignment)
		data, _ := json.Marshal(initial)
		_ = os.WriteFile(filename, data, 0644)
	}
	return &RoleAssignmentStore{filename: filename}
}

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

func (s *RoleAssignmentStore) Get(username string) (RoleAssignment, bool, error) {
	m, err := s.load()
	if err != nil {
		return RoleAssignment{}, false, err
	}
	assignment, ok := m[username]
	return assignment, ok, nil
}

func (s *RoleAssignmentStore) Set(username string, assignment RoleAssignment) error {
	m, err := s.load()
	if err != nil {
		return err
	}
	m[username] = assignment
	return s.save(m)
}

func (s *RoleAssignmentStore) Delete(username string) error {
	m, err := s.load()
	if err != nil {
		return err
	}
	delete(m, username)
	return s.save(m)
}
