package assign_next_page_handler

import (
	"context"
	"fmt"
	testsService "github.com/IT-Nick/internal/domain/tests/service"
	"github.com/IT-Nick/internal/domain/users/service"
	"gopkg.in/telebot.v4"
	"sync"
)

type AssignNextPageHandler struct {
	userService *service.UserService
	testService *testsService.TestService
	pageState   map[int64]int
	mutex       sync.Mutex
}

func NewAssignNextPageHandler(userService *service.UserService, testService *testsService.TestService, pageState map[int64]int) *AssignNextPageHandler {
	return &AssignNextPageHandler{
		userService: userService,
		testService: testService,
		pageState:   pageState,
	}
}

func (h *AssignNextPageHandler) Handle(c telebot.Context) error {
	userID := c.Sender().ID

	h.mutex.Lock() // Захватываем мьютекс
	page := h.pageState[userID]
	if page == 0 {
		page = 1 // Если страница не установлена, начинаем с первой страницы
	}
	h.mutex.Unlock() // Освобождаем мьютекс

	pageSize := 3

	// Запрашиваем тесты с пагинацией
	tests, err := h.testService.GetTestsWithPagination(context.Background(), page+1, pageSize)
	if err != nil {
		return c.Send(fmt.Sprintf("Failed to get tests: %v", err))
	}

	// Получаем общее количество тестов для вычисления последней страницы
	totalTests, err := h.testService.GetTotalTestsCount(context.Background())
	if err != nil {
		return c.Send(fmt.Sprintf("Failed to get total test count: %v", err))
	}

	totalPages := (totalTests + pageSize - 1) / pageSize // вычисляем количество страниц

	// Обновляем текущую страницу в мапе
	h.mutex.Lock()
	h.pageState[userID] = page + 1
	h.mutex.Unlock()

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

	if page > 0 { // Кнопка "prev_page" только если не на первой странице
		paginationButtons = append(paginationButtons, telebot.InlineButton{
			Text:   "<",
			Unique: "prev_page",
		})
	}

	// Если это не последняя страница, показываем кнопку "next_page"
	if page+1 < totalPages {
		paginationButtons = append(paginationButtons, telebot.InlineButton{
			Text:   ">",
			Unique: "next_page",
		})
	} else { // Если на последней странице, показываем "Конец"
		paginationButtons = append(paginationButtons, telebot.InlineButton{
			Text:   "Конец",
			Unique: "end_page",
		})
	}

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

func (h *AssignNextPageHandler) GetHandlerFunc() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		return h.Handle(c)
	}
}
