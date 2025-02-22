package answer_handler

import (
	"context"
	"fmt"
	"github.com/IT-Nick/internal/domain/model"
	testsService "github.com/IT-Nick/internal/domain/tests/service"
	"gopkg.in/telebot.v4"
	"strconv"
	"strings"
	"time"
)

type AnswerHandler struct {
	bot         *telebot.Bot
	testService *testsService.TestService
}

func NewAnswerHandler(
	bot *telebot.Bot,
	testService *testsService.TestService,
) *AnswerHandler {
	return &AnswerHandler{
		bot:         bot,
		testService: testService,
	}
}

// sendQuestion отправляет вопрос пользователю с порядковым номером и общим количеством вопросов
func (h *AnswerHandler) sendQuestion(recipient *telebot.User, question model.Question, questionNumber int) error {
	// Формируем текст вопроса с порядковым номером и общим количеством вопросов
	var messageBuilder strings.Builder
	messageBuilder.WriteString(fmt.Sprintf("❓ *Вопрос %d:*\n%s\n\n", questionNumber, question.QuestionText))

	// Формируем клавиатуру с вариантами ответа
	var buttons []telebot.Btn
	var markup *telebot.ReplyMarkup
	markup = h.bot.NewMarkup()

	if question.TestOptions != nil && len(question.TestOptions) > 0 {
		for i, option := range question.TestOptions {
			btnText := fmt.Sprintf("%d. %s", i+1, option)
			// Используем question.ID для callbackData, чтобы сохранить уникальность
			callbackData := fmt.Sprintf("answer_%d_%d_%s", question.ID, i, option)
			btn := markup.Data(btnText, callbackData)
			buttons = append(buttons, btn)
		}

		// Создаем инлайн-клавиатуру
		rows := make([]telebot.Row, 0)
		for _, btn := range buttons {
			rows = append(rows, markup.Row(btn))
		}
		markup.Inline(rows...)
	}

	// Отправляем сообщение с вопросом
	_, err := h.bot.Send(recipient, messageBuilder.String(), &telebot.SendOptions{
		ParseMode:   telebot.ModeMarkdown,
		ReplyMarkup: markup,
	})
	if err != nil {
		return fmt.Errorf("failed to send question: %w", err)
	}

	return nil
}

func (h *AnswerHandler) Handle(c telebot.Context) error {
	telegramID := c.Sender().ID
	callbackData := c.Callback().Data

	// Очищаем callbackData от нестандартных символов
	cleanedData := strings.TrimSpace(callbackData)
	cleanedData = strings.ReplaceAll(cleanedData, "\f", "")
	cleanedData = strings.ReplaceAll(cleanedData, "\\f", "")

	// Проверяем, что callback начинается с "answer_"
	if !strings.HasPrefix(cleanedData, "answer_") {
		return nil
	}

	// Парсим callback данные (answer_questionID_optionIndex_answerText)
	parts := strings.Split(cleanedData, "_")
	if len(parts) < 4 {
		return fmt.Errorf("invalid callback data: %s", callbackData)
	}

	questionIDStr := parts[1]                  // questionID
	optionIndexStr := parts[2]                 // optionIndex
	answerText := strings.Join(parts[3:], "_") // answerText

	questionID, err := strconv.Atoi(questionIDStr)
	if err != nil {
		return fmt.Errorf("invalid question ID: %w", err)
	}

	optionIndex, err := strconv.Atoi(optionIndexStr)
	if err != nil {
		return fmt.Errorf("invalid option index: %w", err)
	}

	// Получаем текущее состояние теста из базы
	ctx := context.Background()
	userTestID, err := h.testService.GetUserTestIDByUserID(ctx, telegramID)
	if err != nil {
		return c.Send("Тест не найден. Пожалуйста, начните тест заново.")
	}

	currentQuestionIndex, correctAnswersCount, status, err := h.testService.GetUserTestState(ctx, userTestID)
	if err != nil {
		return fmt.Errorf("failed to get user test state: %w", err)
	}

	if status == "finished" {
		return c.Send("Тест уже завершен.")
	}

	// Получаем выбранные вопросы теста
	selectedQuestions, err := h.testService.GetSelectedQuestions(ctx, userTestID)
	if err != nil {
		return fmt.Errorf("failed to get selected questions: %w", err)
	}

	// Проверяем, есть ли вопросы
	if len(selectedQuestions) == 0 {
		return fmt.Errorf("no selected questions found for user test ID %d", userTestID)
	}

	// Проверяем, что currentQuestionIndex валиден
	if currentQuestionIndex < 0 || currentQuestionIndex >= len(selectedQuestions) {
		return fmt.Errorf("invalid current question index: %d, total questions: %d", currentQuestionIndex, len(selectedQuestions))
	}

	currentQuestion := selectedQuestions[currentQuestionIndex]

	// Проверяем, что текущий вопрос соответствует callback
	if currentQuestion.ID != questionID {
		return fmt.Errorf("mismatch between current question ID %d and callback question ID %d", currentQuestion.ID, questionID)
	}

	// Проверяем правильность ответа
	isCorrect := false
	if len(currentQuestion.TestOptions) > optionIndex {
		isCorrect = currentQuestion.TestOptions[optionIndex] == currentQuestion.CorrectAnswer
	}

	// Сохраняем ответ в таблицу answers
	err = h.testService.SaveAnswer(ctx, userTestID, questionID, answerText, isCorrect)
	if err != nil {
		return fmt.Errorf("failed to save answer: %w", err)
	}

	// Обновляем correct_answers_count
	if isCorrect {
		correctAnswersCount++
	}

	// Увеличиваем current_question_index
	currentQuestionIndex++

	// Обновляем состояние теста в базе
	err = h.testService.UpdateUserTestState(ctx, userTestID, currentQuestionIndex, correctAnswersCount)
	if err != nil {
		return fmt.Errorf("failed to update user test state: %w", err)
	}

	// Удаляем предыдущее сообщение с вопросом
	err = h.bot.Delete(c.Message())
	if err != nil {
		return fmt.Errorf("failed to delete previous question: %w", err)
	}

	// Проверяем, есть ли следующий вопрос
	if currentQuestionIndex >= len(selectedQuestions) {
		// Тест завершен
		err = h.testService.UpdateUserTestStatus(ctx, userTestID, "finished")
		err = h.testService.UpdateUserTestEndTime(ctx, userTestID, time.Now())
		if err != nil {
			return fmt.Errorf("failed to update test status: %w", err)
		}
		return c.Send("Тест завершен! Ваши ответы сохранены.")
	}

	// Отправляем следующий вопрос с порядковым номером и общим количеством вопросов
	nextQuestion := selectedQuestions[currentQuestionIndex]
	err = h.sendQuestion(c.Sender(), nextQuestion, currentQuestionIndex+1)
	if err != nil {
		return fmt.Errorf("failed to send next question: %w", err)
	}

	return nil
}
