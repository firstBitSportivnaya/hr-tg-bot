package select_test_handler

import (
	"fmt"
	testsService "github.com/IT-Nick/internal/domain/tests/service"
	usersService "github.com/IT-Nick/internal/domain/users/service"
	tgbotapi "gopkg.in/telebot.v4"
	"strconv"
	"strings"
)

type SelectTestHandler struct {
	userService *usersService.UserService
	testService *testsService.TestService
	testState   map[int64]int
}

func NewSelectTestHandler(
	userService *usersService.UserService,
	testService *testsService.TestService,
	testState map[int64]int,
) *SelectTestHandler {
	return &SelectTestHandler{
		userService: userService,
		testService: testService,
		testState:   testState,
	}
}

func (h *SelectTestHandler) Handle(c tgbotapi.Context) error {
	// Извлекаем ID теста из Unique кнопки
	data := c.Callback().Data

	cleanedData := strings.TrimSpace(data)
	cleanedData = strings.ReplaceAll(cleanedData, "\f", "")
	cleanedData = strings.ReplaceAll(cleanedData, "\\f", "")

	testIDStr := strings.TrimPrefix(cleanedData, "test_")
	testID, err := strconv.Atoi(testIDStr)
	if err != nil {
		return c.Send("Ошибка при выборе теста.")
	}

	// Сохраняем выбранный тест в состояние
	userID := c.Sender().ID
	h.testState[userID] = testID

	// Дополнительные действия (например, запрос кандидата)
	return c.Send(fmt.Sprintf("Тест #%d выбран. Введите имя кандидата (например, @username).", testID))
}

func (h *SelectTestHandler) GetHandlerFunc() tgbotapi.HandlerFunc {
	return func(c tgbotapi.Context) error {
		return h.Handle(c)
	}
}
