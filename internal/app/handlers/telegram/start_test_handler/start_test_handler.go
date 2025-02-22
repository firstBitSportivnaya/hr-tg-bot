package start_test_handler

import (
	"context"
	"fmt"
	messageService "github.com/IT-Nick/internal/domain/messages/service"
	"github.com/IT-Nick/internal/domain/model"
	testService "github.com/IT-Nick/internal/domain/tests/service"
	usersService "github.com/IT-Nick/internal/domain/users/service"
	"github.com/IT-Nick/internal/infra/timer"
	"gopkg.in/telebot.v4"
	"log"
	"math/rand/v2"
	"strings"
	"time"
)

// StartTestHandler структура для обработки нажатия кнопки "Начать тест"
type StartTestHandler struct {
	bot            *telebot.Bot
	testService    *testService.TestService
	messageService *messageService.MessageService
	userService    *usersService.UserService
	timerUpdater   *timer.Updater
}

// NewStartTestHandler возвращает новый экземпляр обработчика
func NewStartTestHandler(
	bot *telebot.Bot,
	testService *testService.TestService,
	messageService *messageService.MessageService,
	userService *usersService.UserService,
	timerUpdater *timer.Updater,
) *StartTestHandler {
	return &StartTestHandler{
		bot:            bot,
		testService:    testService,
		messageService: messageService,
		userService:    userService,
		timerUpdater:   timerUpdater,
	}
}

// sendQuestion отправляет вопрос пользователю с порядковым номером и общим количеством вопросов
func (h *StartTestHandler) sendQuestion(recipient *telebot.User, question model.Question, questionNumber int) error {
	// Формируем текст вопроса с порядковым номером и общим количеством вопросов
	var messageBuilder strings.Builder
	messageBuilder.WriteString(fmt.Sprintf("❓ *Вопрос %d:*\n%s\n\n", questionNumber, question.QuestionText))

	var buttons []telebot.Btn
	var markup *telebot.ReplyMarkup
	markup = h.bot.NewMarkup()

	if question.TestOptions != nil && len(question.TestOptions) > 0 {
		for i, option := range question.TestOptions {
			btnText := fmt.Sprintf("%d. %s", i+1, option)
			callbackData := fmt.Sprintf("answer_%d_%d_%s", question.ID, i, option)
			btn := markup.Data(btnText, callbackData)
			buttons = append(buttons, btn)
		}

		rows := make([]telebot.Row, 0)
		for _, btn := range buttons {
			rows = append(rows, markup.Row(btn))
		}
		markup.Inline(rows...)
	}

	_, err := h.bot.Send(recipient, messageBuilder.String(), &telebot.SendOptions{
		ParseMode:   telebot.ModeMarkdown,
		ReplyMarkup: markup,
	})
	if err != nil {
		return fmt.Errorf("failed to send question: %w", err)
	}

	return nil
}

// Handle обрабатывает callback от кнопки "Начать тест"
func (h *StartTestHandler) Handle(c telebot.Context) error {
	ctx, cancel := context.WithCancel(context.Background())

	username := c.Sender().Username
	userID := c.Sender().ID

	// Получаем доступные тесты для пользователя
	availableTests, err := h.testService.GetAvailableTestsForUser(ctx, username)
	if err != nil {
		return c.Respond(&telebot.CallbackResponse{
			Text: fmt.Sprintf("Ошибка при получении тестов: %v", err),
		})
	}

	if len(availableTests) == 0 {
		noTestsMessage, err := h.messageService.GetMessageByKey(ctx, "no_available_tests")
		if err != nil {
			return c.Respond(&telebot.CallbackResponse{
				Text: fmt.Sprintf("Ошибка при получении сообщения: %v", err),
			})
		}
		return c.Respond(&telebot.CallbackResponse{
			Text: noTestsMessage,
		})
	}

	test := availableTests[0]
	// Начинаем тест для пользователя
	userTestID, err := h.testService.StartTestForUser(ctx, username, test.ID)
	if err != nil {
		return c.Respond(&telebot.CallbackResponse{
			Text: fmt.Sprintf("Ошибка при начале теста: %v", err),
		})
	}

	// Получаем информацию о назначении теста
	userTest, err := h.userService.GetUserTestByID(ctx, userTestID)
	if err != nil {
		return c.Respond(&telebot.CallbackResponse{
			Text: fmt.Sprintf("Ошибка при получении информации о тесте: %v", err),
		})
	}

	// Получаем сообщение о начале теста
	startTestMessage, err := h.messageService.GetMessageByKey(ctx, "start_test_message")
	if err != nil {
		return c.Respond(&telebot.CallbackResponse{
			Text: fmt.Sprintf("Ошибка при получении сообщения: %v", err),
		})
	}

	// Получаем пользователя, который назначил тест
	assignedByUser, err := h.userService.GetUserByID(ctx, userTest.AssignedBy)
	if err != nil {
		return c.Respond(&telebot.CallbackResponse{
			Text: fmt.Sprintf("Ошибка при получении информации о назначившем пользователе: %v", err),
		})
	}

	// Отправляем сообщение о начале теста пользователю assigned_by
	_, err = h.bot.Send(&telebot.User{ID: *assignedByUser.TelegramID}, fmt.Sprintf(startTestMessage, username, test.TestName), &telebot.SendOptions{
		ParseMode: telebot.ModeMarkdown,
	})
	if err != nil {
		log.Printf("Failed to send message to assigned_by user: %v", err)
		// Не прерываем выполнение, так как это не критично для начала теста
	}

	// Отправляем сообщение с таймером пользователю, который начал тест
	timerMessage, err := h.bot.Send(c.Sender(), "Тест формируется...", &telebot.SendOptions{
		ParseMode: telebot.ModeMarkdown,
	})
	if err != nil {
		return c.Respond(&telebot.CallbackResponse{
			Text: fmt.Sprintf("Ошибка при отправке таймера: %v", err),
		})
	}

	// Сохраняем ID сообщения таймера в таблицу user_tests
	err = h.testService.SaveTimerMessageID(ctx, userTestID, timerMessage.ID)
	if err != nil {
		return c.Respond(&telebot.CallbackResponse{
			Text: fmt.Sprintf("Ошибка при сохранении ID таймера: %v", err),
		})
	}

	// Получаем вопросы теста
	questions, err := h.testService.GetQuestionsByTestID(ctx, test.ID)
	if err != nil {
		return c.Respond(&telebot.CallbackResponse{
			Text: fmt.Sprintf("Ошибка при получении вопросов: %v", err),
		})
	}

	// Фильтруем только вопросы типа "single"
	var singleQuestions []model.Question
	for _, q := range questions {
		if q.AnswerType == "single" {
			singleQuestions = append(singleQuestions, q)
		}
	}

	if len(singleQuestions) == 0 {
		return c.Respond(&telebot.CallbackResponse{
			Text: "В тесте нет вопросов с одним правильным ответом.",
		})
	}

	// Проверяем, что вопросов достаточно для question_count
	if len(singleQuestions) < test.QuestionCount {
		return c.Respond(&telebot.CallbackResponse{
			Text: fmt.Sprintf("Недостаточно вопросов в тесте: доступно %d, требуется %d", len(singleQuestions), test.QuestionCount),
		})
	}

	// Выбираем случайные question_count вопросов
	rand.Shuffle(len(singleQuestions), func(i, j int) {
		singleQuestions[i], singleQuestions[j] = singleQuestions[j], singleQuestions[i]
	})
	selectedQuestions := singleQuestions[:test.QuestionCount]

	// Сохраняем ID выбранных вопросов
	var questionIDs []int
	for _, q := range selectedQuestions {
		questionIDs = append(questionIDs, q.ID)
	}
	err = h.testService.SaveSelectedQuestions(ctx, userTestID, questionIDs)
	if err != nil {
		return c.Respond(&telebot.CallbackResponse{
			Text: fmt.Sprintf("Ошибка при сохранении выбранных вопросов: %v", err),
		})
	}

	// Сохраняем начальное состояние теста в базе
	err = h.testService.UpdateUserTestState(ctx, userTestID, 0, 0) // current_question_index = 0, correct_answers_count = 0
	if err != nil {
		return c.Respond(&telebot.CallbackResponse{
			Text: fmt.Sprintf("Ошибка при обновлении состояния теста: %v", err),
		})
	}

	// Обновляем сообщение таймера перед отправкой первого вопроса
	currentQuestionIndex := 0
	totalQuestions := test.QuestionCount
	timeLeft := time.Until(userTest.TimerDeadline)
	minutes := int(timeLeft.Minutes())
	seconds := int(timeLeft.Seconds()) % 60
	timerText := fmt.Sprintf(
		"⏰ Тест начался! Оставшееся время: %02d:%02d, Вопрос %d/%d",
		minutes, seconds, currentQuestionIndex+1, totalQuestions,
	)

	_, err = h.bot.Edit(timerMessage, timerText, &telebot.SendOptions{
		ParseMode: telebot.ModeMarkdown,
	})
	if err != nil {
		log.Printf("Failed to update timer message: %w", err)
	}

	// Запускаем горутину для обновления таймера с контекстом
	go func() {
		defer cancel()
		h.timerUpdater.UpdateTimer(ctx, userID, timerMessage.ID, userTest.TimerDeadline, userTestID, totalQuestions)
	}()

	// Отправляем первый вопрос с порядковым номером
	currentQuestion := selectedQuestions[0]
	err = h.sendQuestion(c.Sender(), currentQuestion, currentQuestionIndex+1)
	if err != nil {
		return c.Respond(&telebot.CallbackResponse{
			Text: fmt.Sprintf("Ошибка при отправке вопроса: %v", err),
		})
	}

	return c.Respond(&telebot.CallbackResponse{
		Text: "Тест успешно начат!",
	})
}

// GetHandlerFunc возвращает обработчик в формате telebot.HandlerFunc
func (h *StartTestHandler) GetHandlerFunc() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		return h.Handle(c)
	}
}
