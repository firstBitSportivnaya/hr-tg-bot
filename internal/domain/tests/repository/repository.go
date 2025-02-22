package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IT-Nick/internal/domain/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"time"
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

// StartTest начинает тест для пользователя и возвращает ID назначения теста (user_test_id)
func (r *TestRepository) StartTest(ctx context.Context, userID, testID int) (int, error) {
	// Проверяем, существует ли назначение теста
	var existingUserTestID int
	checkQuery := `
        SELECT id FROM user_tests 
        WHERE user_id = $1 AND test_id = $2 AND status = 'assigned'
    `
	err := r.db.QueryRow(ctx, checkQuery, userID, testID).Scan(&existingUserTestID)
	if err != nil && err != pgx.ErrNoRows {
		return 0, fmt.Errorf("failed to check existing test assignment: %w", err)
	}

	if err == pgx.ErrNoRows {
		// Если записи нет, создаем новую
		log.Printf("No existing test assignment found, creating new for user %d and test %d", userID, testID)
		query := `
            INSERT INTO user_tests (
                user_id, test_id, status, start_time, 
                current_question_index, correct_answers_count, 
                timer_deadline
            )
            VALUES (
                $1, $2, 'in_progress', CURRENT_TIMESTAMP, 
                0, 0, 
                CURRENT_TIMESTAMP + (SELECT duration * INTERVAL '1 minute' FROM tests WHERE id = $2)
            )
            RETURNING id
        `
		var userTestID int
		err = r.db.QueryRow(ctx, query, userID, testID).Scan(&userTestID)
		if err != nil {
			return 0, fmt.Errorf("failed to create new test assignment: %w", err)
		}
		log.Printf("Created new test assignment with ID %d", userTestID)
		return userTestID, nil
	}

	// Если запись существует, обновляем ее
	log.Printf("Found existing test assignment with ID %d, updating", existingUserTestID)
	query := `
        UPDATE user_tests 
        SET status = 'in_progress',
            start_time = CURRENT_TIMESTAMP,
            current_question_index = 0,
            correct_answers_count = 0,
            timer_deadline = CURRENT_TIMESTAMP + (SELECT duration * INTERVAL '1 minute' FROM tests WHERE id = $2),
            end_time = NULL
        WHERE id = $1
        RETURNING id
    `
	var userTestID int
	err = r.db.QueryRow(ctx, query, existingUserTestID, testID).Scan(&userTestID)
	if err != nil {
		return 0, fmt.Errorf("failed to start test: %w", err)
	}

	log.Printf("Started test with ID %d for user %d", userTestID, userID)
	return userTestID, nil
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

// UpdateTimerMessageID обновляет ID сообщения таймера в таблице user_tests
func (r *TestRepository) UpdateTimerMessageID(ctx context.Context, userTestID int, messageID int) error {
	query := `
        UPDATE user_tests 
        SET message_id = $1,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $2
    `
	commandTag, err := r.db.Exec(ctx, query, messageID, userTestID)
	if err != nil {
		return fmt.Errorf("failed to update timer message ID: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected when updating timer message ID for user_test_id %d", userTestID)
	}
	return nil
}

// GetQuestionsByTestID получает все вопросы для конкретного теста
func (r *TestRepository) GetQuestionsByTestID(ctx context.Context, testID int) ([]model.Question, error) {
	query := `
        SELECT id, test_id, question_text, answer_type, correct_answer, test_options
        FROM questions
        WHERE test_id = $1
        ORDER BY id
    `
	rows, err := r.db.Query(ctx, query, testID)
	if err != nil {
		return nil, fmt.Errorf("failed to query questions: %w", err)
	}
	defer rows.Close()

	var questions []model.Question
	for rows.Next() {
		var q model.Question
		var testOptions []byte
		err := rows.Scan(
			&q.ID,
			&q.TestID,
			&q.QuestionText,
			&q.AnswerType,
			&q.CorrectAnswer,
			&testOptions,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan question: %w", err)
		}

		// Десериализация JSON опций
		if len(testOptions) > 0 {
			if err := json.Unmarshal(testOptions, &q.TestOptions); err != nil {
				return nil, fmt.Errorf("failed to unmarshal test options: %w", err)
			}
		}
		questions = append(questions, q)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over rows: %w", err)
	}

	return questions, nil
}

// UpdateCurrentQuestionIndex обновляет текущий индекс вопроса для пользователя
func (r *TestRepository) UpdateCurrentQuestionIndex(ctx context.Context, userID int64, testID int, index int) error {
	query := `
        UPDATE user_tests
        SET current_question_index = $1,
            updated_at = CURRENT_TIMESTAMP
        WHERE test_id = $2
        AND user_id = (SELECT id FROM users WHERE telegram_id = $3)
        AND status = 'in_progress'
    `
	commandTag, err := r.db.Exec(ctx, query, index, testID, userID)
	if err != nil {
		return fmt.Errorf("failed to update current question index: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected when updating question index for user %d and test %d", userID, testID)
	}
	return nil
}

// UpdateUserTestState обновляет состояние теста в таблице user_tests
func (r *TestRepository) UpdateUserTestState(ctx context.Context, userTestID int, currentQuestionIndex int, correctAnswersCount int) error {
	_, err := r.db.Exec(ctx,
		"UPDATE user_tests SET current_question_index = $1, correct_answers_count = $2 WHERE id = $3",
		currentQuestionIndex, correctAnswersCount, userTestID)
	if err != nil {
		return fmt.Errorf("failed to update user test state: %w", err)
	}
	return nil
}

// SaveAnswer сохраняет ответ пользователя в таблицу answers
func (r *TestRepository) SaveAnswer(ctx context.Context, userTestID int, questionID int, userAnswer string, isCorrect bool) error {
	_, err := r.db.Exec(ctx,
		"INSERT INTO answers (user_test_id, question_id, user_answer, is_correct) VALUES ($1, $2, $3, $4)",
		userTestID, questionID, userAnswer, isCorrect)
	if err != nil {
		return fmt.Errorf("failed to save answer: %w", err)
	}
	return nil
}

// UpdateUserTestStatus обновляет статус теста в таблице user_tests
func (r *TestRepository) UpdateUserTestStatus(ctx context.Context, userTestID int, status string) error {
	_, err := r.db.Exec(ctx,
		"UPDATE user_tests SET status = $1 WHERE id = $2",
		status, userTestID)
	if err != nil {
		return fmt.Errorf("failed to update user test status: %w", err)
	}
	return nil
}

// UpdateUserTestEndTime устанавливает время завершения теста теста в таблице user_tests
func (r *TestRepository) UpdateUserTestEndTime(ctx context.Context, userTestID int, endTime time.Time) error {
	_, err := r.db.Exec(ctx,
		"UPDATE user_tests SET end_time = $1 WHERE id = $2",
		endTime, userTestID)
	if err != nil {
		return fmt.Errorf("failed to update user test status: %w", err)
	}
	return nil
}

// GetUserTestState получает текущее состояние теста из таблицы user_tests
func (r *TestRepository) GetUserTestState(ctx context.Context, userTestID int) (int, int, string, error) {
	var currentQuestionIndex, correctAnswersCount int
	var status string
	err := r.db.QueryRow(ctx,
		"SELECT current_question_index, correct_answers_count, status FROM user_tests WHERE id = $1",
		userTestID).Scan(&currentQuestionIndex, &correctAnswersCount, &status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, 0, "", fmt.Errorf("user test not found")
		}
		return 0, 0, "", fmt.Errorf("failed to get user test state: %w", err)
	}
	return currentQuestionIndex, correctAnswersCount, status, nil
}

// GetUserTestIDByUserID получает ID текущего теста пользователя из таблицы user_tests
func (r *TestRepository) GetUserTestIDByUserID(ctx context.Context, telegramID int64) (int, error) {
	// Сначала находим user_id по telegram_id
	var userID int
	err := r.db.QueryRow(ctx,
		"SELECT id FROM users WHERE telegram_id = $1",
		telegramID).Scan(&userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, fmt.Errorf("user with telegram ID %d not found", telegramID)
		}
		return 0, fmt.Errorf("failed to query user ID: %w", err)
	}

	// Ищем активный тест для user_id
	var userTestID int
	err = r.db.QueryRow(ctx,
		"SELECT id FROM user_tests WHERE user_id = $1 AND status = 'in_progress' ORDER BY created_at DESC LIMIT 1",
		userID).Scan(&userTestID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, fmt.Errorf("no active test found for user ID %d (telegram ID %d)", userID, telegramID)
		}
		return 0, fmt.Errorf("failed to query user test ID: %w", err)
	}
	return userTestID, nil
}

// GetTestIDByUserTestID получает test_id по userTestID из таблицы user_tests
func (r *TestRepository) GetTestIDByUserTestID(ctx context.Context, userTestID int) (int, error) {
	var testID int
	err := r.db.QueryRow(ctx,
		"SELECT test_id FROM user_tests WHERE id = $1",
		userTestID).Scan(&testID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, fmt.Errorf("user test ID %d not found", userTestID)
		}
		return 0, fmt.Errorf("failed to query test ID: %w", err)
	}
	return testID, nil
}

// GetUserTestsByUserID получает все тесты пользователя
func (r *TestRepository) GetUserTestsByUserID(ctx context.Context, userID int) ([]model.UserTest, error) {
	query := `
        SELECT id, user_id, test_id, assigned_by, status, start_time, end_time, current_question_index, 
               correct_answers_count, timer_deadline, created_at, updated_at
        FROM user_tests
        WHERE user_id = $1
        ORDER BY created_at DESC
    `
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user tests: %w", err)
	}
	defer rows.Close()

	var userTests []model.UserTest
	for rows.Next() {
		var ut model.UserTest
		err := rows.Scan(
			&ut.ID,
			&ut.UserID,
			&ut.TestID,
			&ut.AssignedBy,
			&ut.Status,
			&ut.StartTime,
			&ut.EndTime,
			&ut.CurrentQuestionIndex,
			&ut.CorrectAnswersCount,
			&ut.TimerDeadline,
			&ut.CreatedAt,
			&ut.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user test: %w", err)
		}
		userTests = append(userTests, ut)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over rows: %w", err)
	}

	return userTests, nil
}

// GetTestByID получает информацию о тесте по его ID
func (r *TestRepository) GetTestByID(ctx context.Context, testID int) (*model.Test, error) {
	query := `
        SELECT id, test_name, test_type, duration, question_count, created_at, updated_at
        FROM tests
        WHERE id = $1
    `
	var test model.Test
	err := r.db.QueryRow(ctx, query, testID).Scan(
		&test.ID,
		&test.TestName,
		&test.TestType,
		&test.Duration,
		&test.QuestionCount,
		&test.CreatedAt,
		&test.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("test ID %d not found", testID)
		}
		return nil, fmt.Errorf("failed to query test: %w", err)
	}
	return &test, nil
}

// GetAnswersByUserTestID получает все ответы пользователя для конкретного теста
func (r *TestRepository) GetAnswersByUserTestID(ctx context.Context, userTestID int) ([]model.Answer, error) {
	query := `
        SELECT id, user_test_id, question_id, user_answer, is_correct, created_at, updated_at
        FROM answers
        WHERE user_test_id = $1
        ORDER BY created_at
    `
	rows, err := r.db.Query(ctx, query, userTestID)
	if err != nil {
		return nil, fmt.Errorf("failed to query answers: %w", err)
	}
	defer rows.Close()

	var answers []model.Answer
	for rows.Next() {
		var a model.Answer
		err := rows.Scan(
			&a.ID,
			&a.UserTestID,
			&a.QuestionID,
			&a.UserAnswer,
			&a.IsCorrect,
			&a.CreatedAt,
			&a.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan answer: %w", err)
		}
		answers = append(answers, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over rows: %w", err)
	}

	return answers, nil
}

// GetActiveUserTests получает все активные тесты (status = 'in_progress')
func (r *TestRepository) GetActiveUserTests(ctx context.Context) ([]model.UserTest, error) {
	query := `
        SELECT id, user_id, test_id, assigned_by, status, start_time, end_time, current_question_index, 
               correct_answers_count, timer_deadline, created_at, updated_at
        FROM user_tests
        WHERE status = 'in_progress'
        ORDER BY start_time
    `
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active user tests: %w", err)
	}
	defer rows.Close()

	var userTests []model.UserTest
	for rows.Next() {
		var ut model.UserTest
		err := rows.Scan(
			&ut.ID,
			&ut.UserID,
			&ut.TestID,
			&ut.AssignedBy,
			&ut.Status,
			&ut.StartTime,
			&ut.EndTime,
			&ut.CurrentQuestionIndex,
			&ut.CorrectAnswersCount,
			&ut.TimerDeadline,
			&ut.CreatedAt,
			&ut.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user test: %w", err)
		}
		userTests = append(userTests, ut)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over rows: %w", err)
	}

	return userTests, nil
}

// GetUserTestByTestIDAndUserID получает запись user_test по testID и userID
func (r *TestRepository) GetUserTestByTestIDAndUserID(ctx context.Context, testID int, userID int) (*model.UserTest, error) {
	query := `
        SELECT id, user_id, test_id, assigned_by, status
        FROM user_tests
        WHERE test_id = $1 AND user_id = $2 AND status = 'assigned'
    `
	row := r.db.QueryRow(ctx, query, testID, userID)

	var userTest model.UserTest
	err := row.Scan(
		&userTest.ID,
		&userTest.UserID,
		&userTest.TestID,
		&userTest.AssignedBy,
		&userTest.Status,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan user test: %w", err)
	}

	return &userTest, nil
}

// GetQuestionByID получает вопрос по его ID
func (r *TestRepository) GetQuestionByID(ctx context.Context, questionID int) (*model.Question, error) {
	query := `
        SELECT id, test_id, question_text, answer_type, correct_answer, test_options, created_at, updated_at
        FROM questions
        WHERE id = $1
    `
	row := r.db.QueryRow(ctx, query, questionID)

	var question model.Question
	var testOptionsJSON []byte
	err := row.Scan(
		&question.ID,
		&question.TestID,
		&question.QuestionText,
		&question.AnswerType,
		&question.CorrectAnswer,
		&testOptionsJSON,
		&question.CreatedAt,
		&question.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil // Вопрос не найден
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan question: %w", err)
	}

	// Десериализуем test_options из JSONB в []string
	if testOptionsJSON != nil {
		err = json.Unmarshal(testOptionsJSON, &question.TestOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal test options: %w", err)
		}
	}

	return &question, nil
}

// SaveSelectedQuestions сохраняет ID выбранных вопросов в user_tests
func (r *TestRepository) SaveSelectedQuestions(ctx context.Context, userTestID int, questionIDs []int) error {
	query := `
        UPDATE user_tests
        SET selected_question_ids = $2
        WHERE id = $1
    `
	// Преобразуем []int в []int32 для pgtype.Array[int32]
	int32IDs := make([]int32, len(questionIDs))
	for i, id := range questionIDs {
		int32IDs[i] = int32(id)
	}

	// Создаем pgtype.Array[int32]
	array := pgtype.Array[int32]{
		Elements: int32IDs,
		Dims:     []pgtype.ArrayDimension{{Length: int32(len(int32IDs)), LowerBound: 1}},
		Valid:    len(int32IDs) > 0,
	}

	_, err := r.db.Exec(ctx, query, userTestID, array)
	if err != nil {
		return fmt.Errorf("failed to save selected question IDs: %w", err)
	}
	return nil
}

// GetSelectedQuestionIDs получает ID выбранных вопросов из user_tests
func (r *TestRepository) GetSelectedQuestionIDs(ctx context.Context, userTestID int) ([]int, error) {
	query := `
        SELECT selected_question_ids
        FROM user_tests
        WHERE id = $1
    `
	var array pgtype.Array[int32]
	err := r.db.QueryRow(ctx, query, userTestID).Scan(&array)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("user test %d not found", userTestID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get selected question IDs: %w", err)
	}

	// Проверяем, что массив валиден
	if !array.Valid {
		return []int{}, nil
	}

	// Преобразуем []int32 в []int
	questionIDs := make([]int, len(array.Elements))
	for i, elem := range array.Elements {
		questionIDs[i] = int(elem)
	}
	return questionIDs, nil
}
