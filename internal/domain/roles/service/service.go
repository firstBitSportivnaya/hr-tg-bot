package service

import (
	"context"
	"fmt"
	"github.com/IT-Nick/internal/domain/model"
	"github.com/IT-Nick/internal/domain/roles/repository"
	"gopkg.in/telebot.v4"
)

// RoleService для работы с ролями и правами
type RoleService struct {
	rolePermissionRepo *repository.RolePermissionRepository
}

// NewRoleService создает новый экземпляр RoleService
func NewRoleService(rolePermissionRepo *repository.RolePermissionRepository) *RoleService {
	return &RoleService{rolePermissionRepo: rolePermissionRepo}
}

// CreateRoleWithPermissions создает роль с правами
func (s *RoleService) CreateRoleWithPermissions(ctx context.Context, roleName string, permissions []string) (int, error) {
	// Создаем роль
	roleID, err := s.rolePermissionRepo.CreateRole(ctx, roleName)
	if err != nil {
		return 0, fmt.Errorf("failed to create role: %w", err)
	}

	// Привязываем права к роли
	for _, permissionName := range permissions {
		permission, err := s.rolePermissionRepo.GetPermissionByName(ctx, permissionName)
		if err != nil {
			return 0, fmt.Errorf("failed to get permission: %w", err)
		}

		// Связываем роль с правом
		err = s.rolePermissionRepo.AssignPermissionToRole(ctx, roleID, permission.ID)
		if err != nil {
			return 0, fmt.Errorf("failed to assign permission to role: %w", err)
		}
	}

	return roleID, nil
}

// GetRoleByTelegramName получает роль по имени телеграм
func (s *RoleService) GetRoleByTelegramName(ctx context.Context, telegramName string) (*model.Role, error) {
	return s.rolePermissionRepo.GetRoleByTelegramName(ctx, telegramName)
}

// GetRoleByRoleName получает роль по имени роли
func (s *RoleService) GetRoleByRoleName(ctx context.Context, roleName string) (*model.Role, error) {
	return s.rolePermissionRepo.GetRoleByRoleName(ctx, roleName)
}

// GetPermissionsForUser получает все права для пользователя
func (s *RoleService) GetPermissionsForUser(ctx context.Context, roleID int) ([]string, error) {
	permissions, err := s.rolePermissionRepo.GetPermissionsByRoleId(ctx, roleID)
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
func (s *RoleService) GetRoleBasedKeyboard(ctx context.Context, username string, buttonsMessages map[string]string) ([][]telebot.InlineButton, error) {
	// Получаем роль пользователя
	role, err := s.GetRoleByTelegramName(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get role for user: %w", err)
	}

	// Получаем права для роли
	permissions, err := s.GetPermissionsForUser(ctx, role.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions for role: %w", err)
	}

	// Формируем клавиатуру
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
