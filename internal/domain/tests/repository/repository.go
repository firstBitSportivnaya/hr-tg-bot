package repository

import (
	"context"
	"fmt"
	"github.com/IT-Nick/internal/domain/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TestRepository репозиторий для работы с тестами
type TestRepository struct {
	db *pgxpool.Pool
}

// NewTestRepository создает новый экземпляр TestRepository
func NewTestRepository(db *pgxpool.Pool) *TestRepository {
	return &TestRepository{db: db}
}

// GetTestsWithPagination получает тесты с пагинацией
func (r *TestRepository) GetTestsWithPagination(ctx context.Context, page int, pageSize int) ([]model.Test, error) {
	offset := (page - 1) * pageSize
	rows, err := r.db.Query(ctx, "SELECT id, test_name, test_type, duration, question_count FROM tests LIMIT $1 OFFSET $2", pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query tests: %w", err)
	}
	defer rows.Close()

	var tests []model.Test
	for rows.Next() {
		var test model.Test
		if err := rows.Scan(&test.ID, &test.TestName, &test.TestType, &test.Duration, &test.QuestionCount); err != nil {
			return nil, fmt.Errorf("failed to scan test: %w", err)
		}
		tests = append(tests, test)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over rows: %w", err)
	}

	return tests, nil
}

// GetTotalTestsCount возвращает общее количество тестов
func (r *TestRepository) GetTotalTestsCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM tests").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get total test count: %w", err)
	}
	return count, nil
}

// AssignTestToUser назначает тест существующему пользователю
func (r *TestRepository) AssignTestToUser(ctx context.Context, userID int, testID int, assignedByID int) (int, error) {
	var userTestID int
	err := r.db.QueryRow(ctx, `
                INSERT INTO user_tests (user_id, test_id, assigned_by, status, created_at, updated_at) 
                VALUES ($1, $2, $3, 'assigned', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) 
                RETURNING id
        `, userID, testID, assignedByID).Scan(&userTestID)

	if err != nil {
		return 0, fmt.Errorf("failed to assign test to user: %w", err)
	}
	return userTestID, nil
}

// AssignPendingTest создает отложенное назначение теста
func (r *TestRepository) AssignPendingTest(ctx context.Context, telegramUsername string, testID int, assignedByID int) (int, error) {
	var userTestID int
	err := r.db.QueryRow(ctx, `
                INSERT INTO user_tests (pending_username, test_id, assigned_by, status, created_at, updated_at) 
                VALUES ($1, $2, $3, 'pending', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) 
                RETURNING id
        `, telegramUsername, testID, assignedByID).Scan(&userTestID)

	if err != nil {
		return 0, fmt.Errorf("failed to assign pending test: %w", err)
	}
	return userTestID, nil
}

// GetAvailableTestsForUser получает список доступных тестов для пользователя
func (r *TestRepository) GetAvailableTestsForUser(ctx context.Context, userID int) ([]model.Test, error) {
	query := `
                SELECT t.id, t.test_name, t.test_type, t.duration, t.question_count
                FROM tests t
                JOIN user_tests ut ON t.id = ut.test_id
                WHERE ut.user_id = $1 AND ut.status = 'assigned'
        `

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query available tests: %w", err)
	}
	defer rows.Close()

	var tests []model.Test
	for rows.Next() {
		var test model.Test
		err := rows.Scan(
			&test.ID,
			&test.TestName,
			&test.TestType,
			&test.Duration,
			&test.QuestionCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan test: %w", err)
		}
		tests = append(tests, test)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error in rows: %w", err)
	}

	return tests, nil
}

// CheckTestAssignment проверяет, назначен ли тест пользователю
func (r *TestRepository) CheckTestAssignment(ctx context.Context, userID, testID int) (bool, error) {
	query := `
                SELECT EXISTS (
                        SELECT 1 
                        FROM user_tests 
                        WHERE user_id = $1 
                        AND test_id = $2 
                        AND status = 'assigned'
                )
        `

	var exists bool
	err := r.db.QueryRow(ctx, query, userID, testID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check test assignment: %w", err)
	}

	return exists, nil
}

// StartTest начинает тест для пользователя
func (r *TestRepository) StartTest(ctx context.Context, userID, testID int) error {
	query := `
                UPDATE user_tests 
                SET status = 'in_progress',
                        start_time = CURRENT_TIMESTAMP,
                        current_question_index = 0,
                        correct_answers_count = 0,
                        timer_deadline = CURRENT_TIMESTAMP + (SELECT duration * INTERVAL '1 minute' FROM tests WHERE id = $2)
                WHERE user_id = $1 
                AND test_id = $2 
                AND status = 'assigned'
        `

	result, err := r.db.Exec(ctx, query, userID, testID)
	if err != nil {
		return fmt.Errorf("failed to start test: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("no test assignment found or test already started")
	}

	return nil
}

// GetPendingTests получает отложенные тесты для пользователя
func (r *TestRepository) GetPendingTests(ctx context.Context, telegramUsername string) ([]struct {
	TestID       int
	AssignedByID int
}, error) {
	rows, err := r.db.Query(ctx, `
                SELECT test_id, assigned_by 
                FROM user_tests 
                WHERE pending_username = $1 AND status = 'pending'
        `, telegramUsername)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending tests: %w", err)
	}
	defer rows.Close()

	var tests []struct {
		TestID       int
		AssignedByID int
	}
	for rows.Next() {
		var test struct {
			TestID       int
			AssignedByID int
		}
		if err := rows.Scan(&test.TestID, &test.AssignedByID); err != nil {
			return nil, fmt.Errorf("failed to scan pending test: %w", err)
		}
		tests = append(tests, test)
	}

	return tests, nil
}

// ActivatePendingTests активирует отложенные тесты для нового пользователя
func (r *TestRepository) ActivatePendingTests(ctx context.Context, userID int, telegramUsername string) error {
	_, err := r.db.Exec(ctx, `
                UPDATE user_tests 
                SET user_id = $1, 
                        pending_username = NULL,
                        status = 'assigned',
                        updated_at = CURRENT_TIMESTAMP
                WHERE pending_username = $2 AND status = 'pending'
        `, userID, telegramUsername)
	if err != nil {
		return fmt.Errorf("failed to activate pending tests: %w", err)
	}
	return nil
}
