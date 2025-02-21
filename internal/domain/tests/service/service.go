package service

import (
	"context"
	"fmt"
	"github.com/IT-Nick/internal/domain/model"
	"github.com/IT-Nick/internal/domain/tests/repository"
	usersRepo "github.com/IT-Nick/internal/domain/users/repository"
)

// TestService для работы с тестами
type TestService struct {
	testRepo *repository.TestRepository
	userRepo *usersRepo.UserRepository
}

// NewTestService создает новый экземпляр TestService
func NewTestService(testRepo *repository.TestRepository, userRepo *usersRepo.UserRepository) *TestService {
	return &TestService{
		testRepo: testRepo,
		userRepo: userRepo,
	}
}

// GetTestsWithPagination получает тесты с пагинацией
func (s *TestService) GetTestsWithPagination(ctx context.Context, page int, pageSize int) ([]model.Test, error) {
	tests, err := s.testRepo.GetTestsWithPagination(ctx, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get tests: %w", err)
	}
	return tests, nil
}

// GetTotalTestsCount получает общее количество тестов
func (s *TestService) GetTotalTestsCount(ctx context.Context) (int, error) {
	count, err := s.testRepo.GetTotalTestsCount(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get total test count: %w", err)
	}
	return count, nil
}

// AssignTestToUser назначает тест существующему пользователю
func (s *TestService) AssignTestToUser(ctx context.Context, userID int, testID int, assignedByUsername string) (int, error) {
	assignedBy, err := s.userRepo.GetUserByUsername(ctx, assignedByUsername)
	if err != nil {
		return 0, fmt.Errorf("failed to get assigning user: %w", err)
	}
	if assignedBy == nil {
		return 0, fmt.Errorf("assigning user %s not found", assignedByUsername)
	}

	userTestID, err := s.testRepo.AssignTestToUser(ctx, userID, testID, assignedBy.ID)
	if err != nil {
		return 0, fmt.Errorf("failed to assign test: %w", err)
	}
	return userTestID, nil
}

// AssignPendingTest создает отложенное назначение теста
func (s *TestService) AssignPendingTest(ctx context.Context, telegramUsername string, testID int, assignedByUsername string) (int, error) {
	assignedBy, err := s.userRepo.GetUserByUsername(ctx, assignedByUsername)
	if err != nil {
		return 0, fmt.Errorf("failed to get assigning user: %w", err)
	}
	if assignedBy == nil {
		return 0, fmt.Errorf("assigning user %s not found", assignedByUsername)
	}

	userTestID, err := s.testRepo.AssignPendingTest(ctx, telegramUsername, testID, assignedBy.ID)
	if err != nil {
		return 0, fmt.Errorf("failed to assign pending test: %w", err)
	}
	return userTestID, nil
}

// GetAvailableTestsForUser получает список доступных тестов для пользователя
func (s *TestService) GetAvailableTestsForUser(ctx context.Context, username string) ([]model.Test, error) {
	// Получаем пользователя по username
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user %s not found", username)
	}

	// Получаем доступные тесты для пользователя
	tests, err := s.testRepo.GetAvailableTestsForUser(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get available tests: %w", err)
	}

	return tests, nil
}

// StartTestForUser начинает тест для пользователя
func (s *TestService) StartTestForUser(ctx context.Context, username string, testID int) error {
	// Получаем пользователя по username
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user %s not found", username)
	}

	// Проверяем, назначен ли тест пользователю
	isAssigned, err := s.testRepo.CheckTestAssignment(ctx, user.ID, testID)
	if err != nil {
		return fmt.Errorf("failed to check test assignment: %w", err)
	}

	if !isAssigned {
		return fmt.Errorf("test %d is not assigned to user %d", testID, user.ID)
	}

	// Начинаем тест
	err = s.testRepo.StartTest(ctx, user.ID, testID)
	if err != nil {
		return fmt.Errorf("failed to start test: %w", err)
	}

	return nil
}

// ProcessPendingTests обрабатывает отложенные тесты для нового пользователя
func (s *TestService) ProcessPendingTests(ctx context.Context, userID int, telegramUsername string) error {
	// Получаем отложенные тесты
	pendingTests, err := s.testRepo.GetPendingTests(ctx, telegramUsername)
	if err != nil {
		return fmt.Errorf("failed to get pending tests: %w", err)
	}

	if len(pendingTests) == 0 {
		return nil
	}

	// Активируем отложенные тесты
	err = s.testRepo.ActivatePendingTests(ctx, userID, telegramUsername)
	if err != nil {
		return fmt.Errorf("failed to activate pending tests: %w", err)
	}

	return nil
}
