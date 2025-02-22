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
	err := r.db.QueryRow(ctx, "SELECT id, role_id, telegram_id, telegram_username FROM users WHERE telegram_username=$1", username).
		Scan(&user.ID, &user.RoleID, &user.TelegramID, &user.TelegramUsername)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Если пользователя нет, возвращаем nil
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return &user, nil
}

// CreateUser создает нового пользователя в базе данных
func (r *UserRepository) CreateUser(ctx context.Context, username string, telegramId int64, telegramFirstName string, roleID int) (int, error) {
	var userID int
	err := r.db.QueryRow(ctx, "INSERT INTO users (telegram_id, telegram_username, telegram_first_name, role_id) VALUES ($1, $2, $3, $4) RETURNING id", telegramId, username, telegramFirstName, roleID).
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

// GetUserByTelegramID получает пользователя по ID telegram
func (r *UserRepository) GetUserByTelegramID(ctx context.Context, telegramID int64) (*model.User, error) {
	query := `
        SELECT id, role_id, telegram_id, telegram_username, telegram_first_name, real_first_name, 
               real_second_name, real_surname, current_state, created_at, updated_at
        FROM users
        WHERE telegram_id = $1
    `
	var user model.User
	err := r.db.QueryRow(ctx, query, telegramID).Scan(
		&user.ID, &user.RoleID, &user.TelegramID, &user.TelegramUsername, &user.TelegramFirstName, &user.RealFirstName,
		&user.RealSecondName, &user.RealSurname, &user.CurrentState, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return &user, nil
}

// GetUserByID получает пользователя по ID
func (r *UserRepository) GetUserByID(ctx context.Context, userID int) (*model.User, error) {
	query := `
        SELECT id, role_id, telegram_id, telegram_username, telegram_first_name, real_first_name, 
               real_second_name, real_surname, current_state, created_at, updated_at
        FROM users
        WHERE id = $1
    `
	var user model.User
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&user.ID, &user.RoleID, &user.TelegramID, &user.TelegramUsername, &user.TelegramFirstName, &user.RealFirstName,
		&user.RealSecondName, &user.RealSurname, &user.CurrentState, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return &user, nil
}

// GetUserTestByID получает назначение теста по ID
func (r *UserRepository) GetUserTestByID(ctx context.Context, userTestID int) (*model.UserTest, error) {
	query := `
        SELECT id, user_id, test_id, assigned_by, pending_username, current_question_index, 
               correct_answers_count, message_id, timer_deadline, start_time, end_time, status,
               created_at, updated_at
        FROM user_tests
        WHERE id = $1
    `
	var userTest model.UserTest
	err := r.db.QueryRow(ctx, query, userTestID).Scan(
		&userTest.ID, &userTest.UserID, &userTest.TestID, &userTest.AssignedBy, &userTest.PendingUsername,
		&userTest.CurrentQuestionIndex, &userTest.CorrectAnswersCount, &userTest.MessageID, &userTest.TimerDeadline,
		&userTest.StartTime, &userTest.EndTime, &userTest.Status, &userTest.CreatedAt, &userTest.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user test by ID: %w", err)
	}
	return &userTest, nil
}
