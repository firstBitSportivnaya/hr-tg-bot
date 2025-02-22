package assign_handler

import (
	"context"
	"fmt"
	testsService "github.com/IT-Nick/internal/domain/tests/service"
	"github.com/IT-Nick/internal/domain/users/service"
	"gopkg.in/telebot.v4"
	"sync"
)

type AssignStartPageHandler struct {
	userService *service.UserService
	testService *testsService.TestService
	pageState   map[int64]int
	mutex       sync.Mutex
}

func NewAssignStartPageHandler(userService *service.UserService, testService *testsService.TestService, pageState map[int64]int) *AssignStartPageHandler {
	return &AssignStartPageHandler{
		userService: userService,
		testService: testService,
		pageState:   pageState,
	}
}

func (h *AssignStartPageHandler) Handle(c telebot.Context) error {
	userID := c.Sender().ID

	h.mutex.Lock()
	page := h.pageState[userID]
	if page == 0 {
		page = 1
	}
	h.mutex.Unlock()

	pageSize := 3

	// Запрашиваем тесты с пагинацией
	tests, err := h.testService.GetTestsWithPagination(context.Background(), page, pageSize)
	if err != nil {
		return c.Send(fmt.Sprintf("Failed to get tests: %v", err))
	}

	// Удаляем старое сообщение
	if err := c.Delete(); err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	// Формируем список кнопок с тестами
	var buttons []telebot.InlineButton
	for _, test := range tests {
		buttons = append(buttons, telebot.InlineButton{
			Text:   test.TestName,
			Unique: fmt.Sprintf("test_%d", test.ID),
		})
	}

	// Кнопки пагинации
	var paginationButtons []telebot.InlineButton
	paginationButtons = append(paginationButtons, telebot.InlineButton{
		Text:   "Начало",
		Unique: "start_page",
	})
	paginationButtons = append(paginationButtons, telebot.InlineButton{
		Text:   ">",
		Unique: "next_page",
	})

	// Конечная клавиатура
	var keyboard [][]telebot.InlineButton
	for _, button := range buttons {
		keyboard = append(keyboard, []telebot.InlineButton{button})
	}
	keyboard = append(keyboard, paginationButtons)

	// Отправляем новое сообщение с клавиатурой
	return c.Send("Какой тест назначить кандидату?", &telebot.SendOptions{
		ReplyMarkup: &telebot.ReplyMarkup{
			InlineKeyboard: keyboard,
		},
	})
}

func (h *AssignStartPageHandler) GetHandlerFunc() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		return h.Handle(c)
	}
}
