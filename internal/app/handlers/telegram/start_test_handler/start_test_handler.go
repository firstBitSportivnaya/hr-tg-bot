package start_test_handler

import (
	"context"
	"fmt"
	messageService "github.com/IT-Nick/internal/domain/messages/service"
	testService "github.com/IT-Nick/internal/domain/tests/service"
	"gopkg.in/telebot.v4"
)

// StartTestHandler структура для обработки нажатия кнопки "Начать тест"
type StartTestHandler struct {
	testService    *testService.TestService
	messageService *messageService.MessageService
	userState      map[int64]int
}

// NewStartTestHandler возвращает новый экземпляр обработчика
func NewStartTestHandler(testService *testService.TestService, messageService *messageService.MessageService, userState map[int64]int) *StartTestHandler {
	return &StartTestHandler{
		testService:    testService,
		messageService: messageService,
		userState:      userState,
	}
}

// Handle обрабатывает callback от кнопки "Начать тест"
func (h *StartTestHandler) Handle(c telebot.Context) error {
	ctx := context.Background()
	username := c.Sender().Username
	userID := c.Sender().ID

	startTestMessage, err := h.messageService.GetMessageByKey(ctx, "start_test_message")
	if err != nil {
		return c.Respond(&telebot.CallbackResponse{
			Text: fmt.Sprintf("Ошибка при получении сообщения: %v", err),
		})
	}

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
	err = h.testService.StartTestForUser(ctx, username, test.ID)
	if err != nil {
		return c.Respond(&telebot.CallbackResponse{
			Text: fmt.Sprintf("Ошибка при начале теста: %v", err),
		})
	}

	h.userState[userID] = test.ID

	err = c.Send(fmt.Sprintf(startTestMessage, test.TestName), &telebot.SendOptions{
		ParseMode: telebot.ModeMarkdown,
	})
	if err != nil {
		return err
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
