package repository

import (
	"context"
	"fmt"
	"github.com/IT-Nick/internal/domain/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository реализация интерфейса с использованием базы данных PostgreSQL
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository создает новый экземпляр UserRepositoryPg
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// GetUserByUsername ищет пользователя по его Telegram-username
func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.QueryRow(ctx, "SELECT id, role_id, telegram_username FROM users WHERE telegram_username=$1", username).
		Scan(&user.ID, &user.RoleID, &user.TelegramUsername)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Если пользователя нет, возвращаем nil
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return &user, nil
}

// CreateUser создает нового пользователя в базе данных
func (r *UserRepository) CreateUser(ctx context.Context, username string, roleID int) (int, error) {
	var userID int
	err := r.db.QueryRow(ctx, "INSERT INTO users (telegram_username, role_id) VALUES ($1, $2) RETURNING id", username, roleID).
		Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}
	return userID, nil
}

// UpdateUserRole обновляет роль пользователя в базе данных
func (r *UserRepository) UpdateUserRole(ctx context.Context, username string, roleID int) (int, error) {
	// Обновляем роль пользователя по username
	var userID int
	err := r.db.QueryRow(ctx, "UPDATE users SET role_id=$1 WHERE telegram_username=$2 RETURNING id", roleID, username).
		Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("failed to update user role: %w", err)
	}
	return userID, nil
}
