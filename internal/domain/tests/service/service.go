package service

import (
	"context"
	"fmt"
	"github.com/IT-Nick/internal/domain/dto"
	"github.com/IT-Nick/internal/domain/model"
	"github.com/IT-Nick/internal/domain/tests/repository"
	usersRepo "github.com/IT-Nick/internal/domain/users/repository"
	"log"
	"time"
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

// StartTestForUser начинает тест для пользователя и возвращает ID назначения теста (user_test_id)
func (s *TestService) StartTestForUser(ctx context.Context, username string, testID int) (int, error) {
	// Получаем пользователя по username
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return 0, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return 0, fmt.Errorf("user %s not found", username)
	}

	// Проверяем, назначен ли тест пользователю
	isAssigned, err := s.testRepo.CheckTestAssignment(ctx, user.ID, testID)
	if err != nil {
		return 0, fmt.Errorf("failed to check test assignment: %w", err)
	}
	if !isAssigned {
		return 0, fmt.Errorf("test %d is not assigned to user %d", testID, user.ID)
	}

	// Получаем тест для проверки duration
	test, err := s.testRepo.GetTestByID(ctx, testID)
	if err != nil {
		return 0, fmt.Errorf("failed to get test: %w", err)
	}

	// Проверяем duration теста
	if test.Duration <= 0 {
		return 0, fmt.Errorf("invalid test duration: %d minutes", test.Duration)
	}
	log.Printf("Starting test %d for user %d with duration %d minutes", testID, user.ID, test.Duration)

	// Начинаем тест и получаем user_test_id
	userTestID, err := s.testRepo.StartTest(ctx, user.ID, testID)
	if err != nil {
		return 0, fmt.Errorf("failed to start test: %w", err)
	}

	return userTestID, nil
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

func (s *TestService) SaveTimerMessageID(ctx context.Context, userTestID int, messageID int) error {
	return s.testRepo.UpdateTimerMessageID(ctx, userTestID, messageID)
}

func (s *TestService) GetQuestionsByTestID(ctx context.Context, testID int) ([]model.Question, error) {
	return s.testRepo.GetQuestionsByTestID(ctx, testID)
}

func (s *TestService) UpdateCurrentQuestionIndex(ctx context.Context, userID int64, testID int, index int) error {
	return s.testRepo.UpdateCurrentQuestionIndex(ctx, userID, testID, index)
}

// UpdateUserTestState обновляет состояние теста в таблице user_tests
func (s *TestService) UpdateUserTestState(ctx context.Context, userTestID int, currentQuestionIndex int, correctAnswersCount int) error {
	err := s.testRepo.UpdateUserTestState(ctx, userTestID, currentQuestionIndex, correctAnswersCount)
	if err != nil {
		return fmt.Errorf("failed to update user test state: %w", err)
	}
	return nil
}

// SaveAnswer сохраняет ответ пользователя в таблицу answers
func (s *TestService) SaveAnswer(ctx context.Context, userTestID int, questionID int, userAnswer string, isCorrect bool) error {
	err := s.testRepo.SaveAnswer(ctx, userTestID, questionID, userAnswer, isCorrect)
	if err != nil {
		return fmt.Errorf("failed to save answer: %w", err)
	}
	return nil
}

// UpdateUserTestStatus обновляет статус теста в таблице user_tests
func (s *TestService) UpdateUserTestStatus(ctx context.Context, userTestID int, status string) error {
	err := s.testRepo.UpdateUserTestStatus(ctx, userTestID, status)
	if err != nil {
		return fmt.Errorf("failed to update user test status: %w", err)
	}
	return nil
}

func (s *TestService) UpdateUserTestEndTime(ctx context.Context, userTestID int, endTime time.Time) error {
	err := s.testRepo.UpdateUserTestEndTime(ctx, userTestID, endTime)
	if err != nil {
		return fmt.Errorf("failed to update user test end time: %w", err)
	}
	return nil
}

// GetUserTestState получает текущее состояние теста из таблицы user_tests
func (s *TestService) GetUserTestState(ctx context.Context, userTestID int) (int, int, string, error) {
	currentQuestionIndex, correctAnswersCount, status, err := s.testRepo.GetUserTestState(ctx, userTestID)
	if err != nil {
		return 0, 0, "", fmt.Errorf("failed to get user test state: %w", err)
	}
	return currentQuestionIndex, correctAnswersCount, status, nil
}

// GetUserTestIDByUserID получает ID текущего теста пользователя из таблицы user_tests
func (s *TestService) GetUserTestIDByUserID(ctx context.Context, userID int64) (int, error) {
	userTestID, err := s.testRepo.GetUserTestIDByUserID(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get user test ID by user ID: %w", err)
	}
	return userTestID, nil
}

// GetTestIDByUserTestID получает test_id по userTestID из таблицы user_tests
func (s *TestService) GetTestIDByUserTestID(ctx context.Context, userTestID int) (int, error) {
	testID, err := s.testRepo.GetTestIDByUserTestID(ctx, userTestID)
	if err != nil {
		return 0, fmt.Errorf("failed to get test ID by user test ID: %w", err)
	}
	return testID, nil
}

// GetUserTestReport получает полный отчет по тестам пользователя
func (s *TestService) GetUserTestReport(ctx context.Context, userID int) ([]dto.TestHistory, error) {
	// Получаем все тесты пользователя
	userTests, err := s.testRepo.GetUserTestsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tests: %w", err)
	}

	var testHistory []dto.TestHistory
	for _, userTest := range userTests {
		// Получаем информацию о тесте
		test, err := s.testRepo.GetTestByID(ctx, userTest.TestID)
		if err != nil {
			return nil, fmt.Errorf("failed to get test %d: %w", userTest.TestID, err)
		}

		// Получаем вопросы теста
		questions, err := s.testRepo.GetQuestionsByTestID(ctx, userTest.TestID)
		if err != nil {
			return nil, fmt.Errorf("failed to get questions for test %d: %w", userTest.TestID, err)
		}

		// Получаем ответы пользователя
		answers, err := s.testRepo.GetAnswersByUserTestID(ctx, userTest.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get answers for user test %d: %w", userTest.ID, err)
		}

		// Получаем информацию о назначившем пользователе
		assignedByUser, err := s.userRepo.GetUserByID(ctx, userTest.AssignedBy)
		if err != nil {
			return nil, fmt.Errorf("failed to get assigned by user %d: %w", userTest.AssignedBy, err)
		}
		assignedByUsername := ""
		if assignedByUser != nil {
			assignedByUsername = assignedByUser.TelegramUsername
		}

		// Формируем список вопросов с ответами
		var questionInfos []dto.QuestionInfo
		for _, q := range questions {
			var userAnswer string
			var isCorrect bool
			var answeredAt string

			// Проверяем, есть ли ответ для этого вопроса
			for _, a := range answers {
				if a.QuestionID == q.ID {
					userAnswer = a.UserAnswer
					isCorrect = a.IsCorrect
					answeredAt = a.CreatedAt.String()
					break
				}
			}

			// Проверяем TestOptions, чтобы избежать nil
			testOptions := []string{}
			if q.TestOptions != nil {
				testOptions = q.TestOptions
			}

			questionInfos = append(questionInfos, dto.QuestionInfo{
				QuestionID:    q.ID,
				QuestionText:  q.QuestionText,
				AnswerType:    q.AnswerType,
				CorrectAnswer: q.CorrectAnswer,
				TestOptions:   testOptions,
				UserAnswer:    userAnswer,
				IsCorrect:     isCorrect,
				AnsweredAt:    answeredAt,
			})
		}

		// Проверяем указатели в модели UserTest
		status := ""
		if userTest.Status != nil {
			status = *userTest.Status
		}

		startTime := userTest.StartTime.String()
		endTime := ""
		if userTest.EndTime != nil {
			endTime = userTest.EndTime.String()
		}

		timerDeadline := userTest.TimerDeadline.String()

		testHistory = append(testHistory, dto.TestHistory{
			UserTestID:     userTest.ID,
			TestID:         test.ID,
			TestName:       test.TestName,
			TestType:       test.TestType,
			Duration:       test.Duration,
			QuestionCount:  test.QuestionCount,
			Status:         status,
			StartTime:      startTime,
			EndTime:        endTime,
			CorrectAnswers: userTest.CorrectAnswersCount,
			TotalQuestions: len(questions),
			TimerDeadline:  timerDeadline,
			AssignedBy:     assignedByUsername,
			Questions:      questionInfos,
		})
	}

	return testHistory, nil
}

// GetActiveTests получает список активных тестов (пользователей, решающих тесты)
func (s *TestService) GetActiveTests(ctx context.Context) ([]dto.ActiveTestInfo, error) {
	// Получаем все активные тесты (status = 'in_progress')
	activeTests, err := s.testRepo.GetActiveUserTests(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active user tests: %w", err)
	}

	var activeTestInfos []dto.ActiveTestInfo
	for _, userTest := range activeTests {
		// Получаем информацию о пользователе
		user, err := s.userRepo.GetUserByID(ctx, userTest.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user %d: %w", userTest.UserID, err)
		}
		if user == nil {
			continue // Пропускаем, если пользователь не найден
		}

		// Получаем информацию о тесте
		test, err := s.testRepo.GetTestByID(ctx, userTest.TestID)
		if err != nil {
			return nil, fmt.Errorf("failed to get test %d: %w", userTest.TestID, err)
		}

		// Получаем вопросы теста
		questions, err := s.testRepo.GetQuestionsByTestID(ctx, userTest.TestID)
		if err != nil {
			return nil, fmt.Errorf("failed to get questions for test %d: %w", userTest.TestID, err)
		}

		// Фильтруем вопросы типа "single"
		var singleQuestions []model.Question
		for _, q := range questions {
			if q.AnswerType == "single" {
				singleQuestions = append(singleQuestions, q)
			}
		}

		// Получаем текущий вопрос
		var currentQuestion dto.QuestionInfoActive
		if userTest.CurrentQuestionIndex >= 0 && userTest.CurrentQuestionIndex < len(singleQuestions) {
			q := singleQuestions[userTest.CurrentQuestionIndex]
			currentQuestion = dto.QuestionInfoActive{
				QuestionID:   q.ID,
				QuestionText: q.QuestionText,
				AnswerType:   q.AnswerType,
				TestOptions:  q.TestOptions,
			}
		}

		// Получаем предыдущие ответы
		answers, err := s.testRepo.GetAnswersByUserTestID(ctx, userTest.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get answers for user test %d: %w", userTest.ID, err)
		}

		var previousAnswers []dto.AnswerInfo
		for _, a := range answers {
			var questionText string
			for _, q := range questions {
				if q.ID == a.QuestionID {
					questionText = q.QuestionText
					break
				}
			}
			previousAnswers = append(previousAnswers, dto.AnswerInfo{
				QuestionID:   a.QuestionID,
				QuestionText: questionText,
				UserAnswer:   a.UserAnswer,
				IsCorrect:    a.IsCorrect,
				AnsweredAt:   a.CreatedAt.String(),
			})
		}

		// Вычисляем оставшееся время
		remainingTime := "0"
		if userTest.TimerDeadline.After(time.Now()) {
			timeLeft := userTest.TimerDeadline.Sub(time.Now())
			minutes := int(timeLeft.Minutes())
			seconds := int(timeLeft.Seconds()) % 60
			remainingTime = fmt.Sprintf("%02d:%02d", minutes, seconds)
		}

		fullName := fmt.Sprintf("%s %s %s", user.RealFirstName, user.RealSecondName, user.RealSurname)
		activeTestInfos = append(activeTestInfos, dto.ActiveTestInfo{
			TelegramUsername: user.TelegramUsername,
			FullName:         fullName,
			TestID:           test.ID,
			TestName:         test.TestName,
			TestType:         test.TestType,
			Duration:         test.Duration,
			CurrentQuestion:  currentQuestion,
			PreviousAnswers:  previousAnswers,
			CorrectAnswers:   userTest.CorrectAnswersCount,
			TotalQuestions:   len(singleQuestions),
			RemainingTime:    remainingTime,
			Status:           userTest.Status,
		})
	}

	return activeTestInfos, nil
}

// GetUserTestByTestIDAndUsername получает запись user_test по testID и username
func (s *TestService) GetUserTestByTestIDAndUsername(ctx context.Context, testID int, username string) (*model.UserTest, error) {
	// Получаем пользователя по username
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user %s not found", username)
	}

	// Получаем запись user_test
	userTest, err := s.testRepo.GetUserTestByTestIDAndUserID(ctx, testID, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user test: %w", err)
	}
	if userTest == nil {
		return nil, fmt.Errorf("user test not found for test %d and user %s", testID, username)
	}

	return userTest, nil
}

// SaveSelectedQuestions сохраняет ID выбранных вопросов в user_tests
func (s *TestService) SaveSelectedQuestions(ctx context.Context, userTestID int, questionIDs []int) error {
	return s.testRepo.SaveSelectedQuestions(ctx, userTestID, questionIDs)
}

// GetSelectedQuestions получает выбранные вопросы для теста
func (s *TestService) GetSelectedQuestions(ctx context.Context, userTestID int) ([]model.Question, error) {
	questionIDs, err := s.testRepo.GetSelectedQuestionIDs(ctx, userTestID)
	if err != nil {
		return nil, fmt.Errorf("failed to get selected question IDs: %w", err)
	}
	var questions []model.Question
	for _, id := range questionIDs {
		question, err := s.testRepo.GetQuestionByID(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("failed to get question %d: %w", id, err)
		}
		questions = append(questions, *question)
	}
	return questions, nil
}
