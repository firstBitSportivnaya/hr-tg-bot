package repository

import (
	"context"
	"fmt"
	"github.com/IT-Nick/internal/domain/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RolePermissionRepository репозиторий для работы с ролями, правами и их связями
type RolePermissionRepository struct {
	db *pgxpool.Pool
}

// NewRolePermissionRepository создает новый экземпляр RolePermissionRepository
func NewRolePermissionRepository(db *pgxpool.Pool) *RolePermissionRepository {
	return &RolePermissionRepository{db: db}
}

// CreateRole создает роль
func (r *RolePermissionRepository) CreateRole(ctx context.Context, roleName string) (int, error) {
	var roleID int
	err := r.db.QueryRow(ctx, "INSERT INTO roles (role_name) VALUES ($1) RETURNING id", roleName).
		Scan(&roleID)
	if err != nil {
		return 0, fmt.Errorf("failed to create role: %w", err)
	}
	return roleID, nil
}

// GetRoleByTelegramName получает роль по имени телеграм
func (r *RolePermissionRepository) GetRoleByTelegramName(ctx context.Context, telegramName string) (*model.Role, error) {
	var role model.Role
	err := r.db.QueryRow(ctx, `
		SELECT r.id, r.role_name 
		FROM roles r
		JOIN users u ON u.role_id = r.id
		WHERE u.telegram_username = $1`, telegramName).
		Scan(&role.ID, &role.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get role by username: %w", err)
	}
	return &role, nil
}

// GetRoleByRoleName получает роль по имени роли
func (r *RolePermissionRepository) GetRoleByRoleName(ctx context.Context, roleName string) (*model.Role, error) {
	var role model.Role
	err := r.db.QueryRow(ctx, "SELECT id, role_name FROM roles WHERE role_name=$1", roleName).
		Scan(&role.ID, &role.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get role by name: %w", err)
	}
	return &role, nil
}

// GetPermissionByName получает право по имени
func (r *RolePermissionRepository) GetPermissionByName(ctx context.Context, permissionName string) (*model.Permission, error) {
	var permission model.Permission
	err := r.db.QueryRow(ctx, "SELECT id, permission_name FROM permissions WHERE permission_name=$1", permissionName).
		Scan(&permission.ID, &permission.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get permission by name: %w", err)
	}
	return &permission, nil
}

// AssignPermissionToRole связывает право с ролью
func (r *RolePermissionRepository) AssignPermissionToRole(ctx context.Context, roleID, permissionID int) error {
	_, err := r.db.Exec(ctx, "INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)", roleID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to assign permission to role: %w", err)
	}
	return nil
}

// GetPermissionsByRoleId получает все права для роли
func (r *RolePermissionRepository) GetPermissionsByRoleId(ctx context.Context, roleId int) ([]model.Permission, error) {
	rows, err := r.db.Query(ctx, `
		SELECT p.id, p.permission_name
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1`, roleId)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions by role id: %w", err)
	}
	defer rows.Close()

	var permissions []model.Permission
	for rows.Next() {
		var permission model.Permission
		if err := rows.Scan(&permission.ID, &permission.Name); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, permission)
	}
	return permissions, nil
}
