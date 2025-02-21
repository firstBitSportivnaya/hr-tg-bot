package assign_test_handler

import (
	"context"
	"fmt"
	testsService "github.com/IT-Nick/internal/domain/tests/service"
	"github.com/IT-Nick/internal/domain/users/service"
	"gopkg.in/telebot.v4"
	"strings"
)

// AssignTestHandler обрабатывает назначение теста пользователю
type AssignTestHandler struct {
	userService *service.UserService
	testService *testsService.TestService
	testState   map[int64]int
}

// NewAssignTestHandler возвращает структуру обработчика для назначения теста
func NewAssignTestHandler(
	userService *service.UserService,
	testService *testsService.TestService,
	testState map[int64]int,
) *AssignTestHandler {
	return &AssignTestHandler{
		userService: userService,
		testService: testService,
		testState:   testState,
	}
}

func (h *AssignTestHandler) Handle(c telebot.Context) error {
	userID := c.Sender().ID
	messageText := c.Message().Text

	if !strings.HasPrefix(messageText, "@") {
		return c.Send("Пожалуйста, укажите имя пользователя в формате @username.")
	}

	username := strings.TrimPrefix(messageText, "@")
	if username == "" {
		return c.Send("Имя пользователя не может быть пустым.")
	}

	testID, exists := h.testState[userID]
	if !exists {
		return c.Send("Ошибка: тест не выбран. Пожалуйста, выберите тест заново.")
	}

	assignedBy := c.Sender().Username
	if assignedBy == "" {
		return c.Send("Ошибка: не удалось определить пользователя, назначающего тест. Убедитесь, что у вас установлен username в Telegram.")
	}

	ctx := context.Background()
	user, err := h.userService.GetUserByUsername(ctx, username)
	if err != nil {
		return c.Send(fmt.Sprintf("Ошибка при поиске пользователя @%s: %v", username, err))
	}

	if user == nil {
		_, err = h.testService.AssignPendingTest(ctx, username, testID, assignedBy)
		if err != nil {
			return c.Send(fmt.Sprintf("Ошибка при создании отложенного назначения теста: %v", err))
		}

		// Очищаем состояние теста
		delete(h.testState, userID)

		return c.Send(fmt.Sprintf("Пользователь @%s не найден в системе. Ему был добавлен отложенный тест #%d (когда он напишет /start, он появится в системе уже с назначенным тестом).", username, testID))
	}

	_, err = h.testService.AssignTestToUser(ctx, user.ID, testID, assignedBy)
	if err != nil {
		return c.Send(fmt.Sprintf("Ошибка при назначении теста: %v", err))
	}

	delete(h.testState, userID)

	return c.Send(fmt.Sprintf("Тест #%d успешно назначен пользователю @%s.", testID, username))
}

func (h *AssignTestHandler) GetHandlerFunc() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		return h.Handle(c)
	}
}
