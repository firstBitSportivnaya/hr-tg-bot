package service

import (
	"context"
	"fmt"
	"github.com/IT-Nick/internal/domain/model"
	rolesRepo "github.com/IT-Nick/internal/domain/roles/repository"
	"github.com/IT-Nick/internal/domain/users/repository"
	"gopkg.in/telebot.v4"
)

// UserService содержит логику бизнес-операций для пользователей
type UserService struct {
	userRepo           *repository.UserRepository
	rolePermissionRepo *rolesRepo.RolePermissionRepository
}

// NewUserService создает новый экземпляр UserService
func NewUserService(userRepo *repository.UserRepository, rolePermissionRepo *rolesRepo.RolePermissionRepository) *UserService {
	return &UserService{userRepo: userRepo, rolePermissionRepo: rolePermissionRepo}
}

// GetOrCreateUser возвращает ID пользователя, если он существует, или создает нового
func (s *UserService) GetOrCreateUser(ctx context.Context, username string, telegramId int64, TelegramFirstName string, roleName string) (int, error) {
	// Получаем ID роли по имени
	role, err := s.rolePermissionRepo.GetRoleByRoleName(ctx, roleName)
	if err != nil {
		return 0, fmt.Errorf("failed to get role by name: %w", err)
	}

	// Проверяем, существует ли пользователь
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return 0, fmt.Errorf("failed to get user: %w", err)
	}

	// Если пользователь существует, возвращаем его ID
	if user != nil {
		return user.ID, nil
	}

	// Если пользователь не существует, создаем нового с RoleID
	userID, err := s.userRepo.CreateUser(ctx, username, telegramId, TelegramFirstName, role.ID)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	return userID, nil
}

// GetUserByUsername возвращает пользователя по имени телеграмм
func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetPermissionsForUser получает все права для пользователя
func (s *UserService) GetPermissionsForUser(ctx context.Context, username string) ([]string, error) {
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	permissions, err := s.rolePermissionRepo.GetPermissionsByRoleId(ctx, user.RoleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}

	var permissionNames []string
	for _, perm := range permissions {
		permissionNames = append(permissionNames, perm.Name)
	}
	return permissionNames, nil
}

// GetRoleBasedKeyboard генерирует клавиатуру на основе прав пользователя
func (s *UserService) GetRoleBasedKeyboard(ctx context.Context, username string, buttonsMessages map[string]string) ([][]telebot.InlineButton, error) {
	permissions, err := s.GetPermissionsForUser(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}

	var keyboard []telebot.InlineButton
	for _, permission := range permissions {
		switch permission {
		case "start_test":
			keyboard = append(keyboard, telebot.InlineButton{
				Text:   buttonsMessages["start_test"],
				Unique: "start_test",
				Data:   "start",
			})
		case "assign_test":
			keyboard = append(keyboard, telebot.InlineButton{
				Text:   buttonsMessages["assign_test"],
				Unique: "assign_test",
				Data:   "assign_test",
			})
		case "assign_hr":
			keyboard = append(keyboard, telebot.InlineButton{
				Text:   buttonsMessages["assign_hr"],
				Unique: "assign_hr",
				Data:   "assign_hr",
			})
		case "assign_admin":
			keyboard = append(keyboard, telebot.InlineButton{
				Text:   buttonsMessages["assign_admin"],
				Unique: "assign_admin",
				Data:   "assign_admin",
			})
		}
	}

	var keyboardInColumns [][]telebot.InlineButton
	for _, button := range keyboard {
		keyboardInColumns = append(keyboardInColumns, []telebot.InlineButton{button})
	}

	return keyboardInColumns, nil
}

// UpdateUserRole обновляет роль пользователя в базе данных
func (s *UserService) UpdateUserRole(ctx context.Context, username string, roleName string) (int, error) {
	// Получаем ID роли по имени
	role, err := s.rolePermissionRepo.GetRoleByRoleName(ctx, roleName)
	if err != nil {
		return 0, fmt.Errorf("failed to get role by name: %w", err)
	}

	// Обновляем роль пользователя
	userID, err := s.userRepo.UpdateUserRole(ctx, username, role.ID)
	if err != nil {
		return 0, fmt.Errorf("failed to update user role: %w", err)
	}

	return userID, nil
}

// GetUserByID получает пользователя по ID
func (s *UserService) GetUserByID(ctx context.Context, userID int) (*model.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return user, nil
}

// GetUserByTelegramID получает пользователя по ID telegram
func (s *UserService) GetUserByTelegramID(ctx context.Context, telegramID int64) (*model.User, error) {
	user, err := s.userRepo.GetUserByTelegramID(ctx, telegramID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by telegram ID: %w", err)
	}
	return user, nil
}

// GetUserTestByID получает назначение теста по ID
func (s *UserService) GetUserTestByID(ctx context.Context, userTestID int) (*model.UserTest, error) {
	userTest, err := s.userRepo.GetUserTestByID(ctx, userTestID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user test by ID: %w", err)
	}
	return userTest, nil
}
