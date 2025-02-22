package start_handler

import (
	"context"
	"fmt"
	messageService "github.com/IT-Nick/internal/domain/messages/service"
	"github.com/IT-Nick/internal/domain/model"
	rolesService "github.com/IT-Nick/internal/domain/roles/service"
	testService "github.com/IT-Nick/internal/domain/tests/service"
	usersService "github.com/IT-Nick/internal/domain/users/service"
	"gopkg.in/telebot.v4"
	"strconv"
	"strings"
)

// StartHandler структура для обработки команды /start
type StartHandler struct {
	userService    *usersService.UserService
	messageService *messageService.MessageService
	roleService    *rolesService.RoleService
	testService    *testService.TestService
}

// NewStartHandler возвращает структуру обработчика
func NewStartHandler(
	userService *usersService.UserService,
	messageService *messageService.MessageService,
	roleService *rolesService.RoleService,
	testService *testService.TestService,
) *StartHandler {
	return &StartHandler{
		userService:    userService,
		messageService: messageService,
		roleService:    roleService,
		testService:    testService,
	}
}

// Handle метод, который будет использоваться для обработки команды /start
func (h *StartHandler) Handle(c telebot.Context) error {
	username := c.Sender().Username
	telegramID := c.Sender().ID
	telegramFirstName := c.Sender().FirstName

	if username == "" {
		return c.Send("Username is required")
	}

	// Используем дефолтный контекст
	ctx := context.Background()

	// Попытка получить или создать пользователя
	userID, err := h.userService.GetOrCreateUser(ctx, username, telegramID, telegramFirstName, "user")
	if err != nil {
		return c.Send(fmt.Sprintf("Failed to process user: %v", err))
	}

	user, err := h.userService.GetUserByID(ctx, userID)
	if err != nil {
		return c.Send(fmt.Sprintf("Failed to process user: %v", err))
	}

	// Проверяем параметры start (например, из ссылки/QR-кода)
	startParam := c.Data()
	var testID int
	var assignedBy string
	var assignedTests []model.Test
	if startParam != "" && strings.HasPrefix(startParam, "test_") {
		parts := strings.Split(startParam, "_")
		if len(parts) == 4 && parts[0] == "test" {
			testID, err = strconv.Atoi(parts[1])
			if err != nil {
				return c.Send("Неверный формат ID теста в ссылке.")
			}
			//token := parts[2]
			assignedBy = parts[3]

			//// Проверяем токен
			//valid, err := h.testService.ValidateTestLink(ctx, testID, token)
			//if err != nil || !valid {
			//	return c.Send("Недействительная или истекшая ссылка на тест.")
			//}

			// Проверяем, есть ли назначенные тесты
			assignedTests, err = h.testService.GetAvailableTestsForUser(ctx, username)
			if err != nil {
				return c.Send(fmt.Sprintf("Ошибка при получении тестов: %v", err))
			}

			if len(assignedTests) == 0 {
				// Назначаем тест пользователю
				_, err = h.testService.AssignTestToUser(ctx, userID, testID, assignedBy)
				if err != nil {
					// Если пользователь не найден, используем отложенное назначение
					_, err = h.testService.AssignPendingTest(ctx, username, testID, assignedBy)
					if err != nil {
						return c.Send(fmt.Sprintf("Ошибка при создании отложенного назначения теста: %v", err))
					}
				}

				// Обновляем список назначенных тестов
				assignedTests, err = h.testService.GetAvailableTestsForUser(ctx, username)
				if err != nil {
					return c.Send(fmt.Sprintf("Ошибка при получении тестов: %v", err))
				}
			}
		}
	} else {
		// Если нет параметров start, просто проверяем назначенные тесты
		assignedTests, err = h.testService.GetAvailableTestsForUser(ctx, username)
		if err != nil {
			return c.Send(fmt.Sprintf("Failed to retrieve assigned tests: %v", err))
		}
	}

	// Получаем мапу с кнопками для пользователя
	buttonsMessages, err := h.messageService.GetButtons(ctx)
	if err != nil {
		return c.Send(fmt.Sprintf("Failed to retrieve buttons: %v", err))
	}

	// Генерация клавиатуры в зависимости от прав
	keyboard, err := h.roleService.GetRoleBasedKeyboard(ctx, username, buttonsMessages)
	if err != nil {
		return c.Send(fmt.Sprintf("Failed to generate keyboard: %v", err))
	}

	var welcomeMessage string
	if len(assignedTests) > 0 {
		// Если тест назначен, используем welcome_message_user
		test := assignedTests[0]
		welcomeMessageKey := "welcome_message_user"
		welcomeMessage, err = h.messageService.GetMessageByKey(ctx, welcomeMessageKey)
		if err != nil {
			return c.Send(fmt.Sprintf("Failed to retrieve welcome message: %v", err))
		}

		// Получаем информацию о тесте
		testName := test.TestName
		duration := test.Duration
		questionCount := test.QuestionCount

		// Получаем информацию о HR-менеджере
		userTest, err := h.testService.GetUserTestByTestIDAndUsername(ctx, test.ID, username)
		if err != nil {
			return c.Send(fmt.Sprintf("Failed to retrieve user test: %v", err))
		}
		hrManager, err := h.userService.GetUserByID(ctx, userTest.AssignedBy)
		if err != nil {
			return c.Send(fmt.Sprintf("Failed to retrieve HR manager: %v", err))
		}
		hrManagerName := fmt.Sprintf("%s, @%s", *hrManager.TelegramFirstName, hrManager.TelegramUsername)

		// Форматируем сообщение с параметрами
		welcomeMessage = fmt.Sprintf(welcomeMessage,
			*user.TelegramFirstName,
			testName,
			hrManagerName,
			duration,
			questionCount,
		)
	} else {
		// Если теста нет, используем welcome_message_without_assign
		welcomeMessageKey := "welcome_message_without_assign"
		welcomeMessage, err = h.messageService.GetMessageByKey(ctx, welcomeMessageKey)
		if err != nil {
			return c.Send(fmt.Sprintf("Failed to retrieve welcome message: %v", err))
		}

		// Форматируем сообщение с параметрами
		welcomeMessage = fmt.Sprintf(welcomeMessage,
			*user.TelegramFirstName,
		)
	}

	// Отправляем сообщение с клавиатурой
	return c.Send(welcomeMessage, &telebot.SendOptions{
		ParseMode: telebot.ModeHTML,
		ReplyMarkup: &telebot.ReplyMarkup{
			InlineKeyboard: keyboard,
		},
	})
}

// GetHandlerFunc возвращает обработчик в формате telebot.HandlerFunc
func (h *StartHandler) GetHandlerFunc() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		return h.Handle(c)
	}
}
