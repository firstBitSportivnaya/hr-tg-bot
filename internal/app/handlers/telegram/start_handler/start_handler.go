package start_handler

import (
	"context"
	"fmt"
	messageService "github.com/IT-Nick/internal/domain/messages/service"
	rolesService "github.com/IT-Nick/internal/domain/roles/service"
	"github.com/IT-Nick/internal/domain/users/service"
	"gopkg.in/telebot.v4"
)

// StartHandler структура для обработки команды /start
type StartHandler struct {
	userService    *service.UserService
	messageService *messageService.MessageService
	roleService    *rolesService.RoleService
}

// NewStartHandler возвращает структуру обработчика
func NewStartHandler(userService *service.UserService, messageService *messageService.MessageService, roleService *rolesService.RoleService) *StartHandler {
	return &StartHandler{
		userService:    userService,
		messageService: messageService,
		roleService:    roleService,
	}
}

// Handle метод, который будет использоваться для обработки команды /start
func (h *StartHandler) Handle(c telebot.Context) error {
	username := c.Sender().Username
	if username == "" {
		return c.Send("Username is required")
	}

	// Используем дефолтный контекст
	ctx := context.Background()

	// Попытка получить или создать пользователя
	userID, err := h.userService.GetOrCreateUser(ctx, username, "user")
	if err != nil {
		return c.Send(fmt.Sprintf("Failed to process user: %v", err))
	}

	// Получаем сообщение для пользователя
	welcomeMessage, err := h.messageService.GetMessageByKey(ctx, "welcome_message_user")
	if err != nil {
		return c.Send(fmt.Sprintf("Failed to retrieve welcome message: %v", err))
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

	// Отправляем сообщение с клавиатурой
	return c.Send(fmt.Sprintf(welcomeMessage, userID, username), &telebot.SendOptions{
		ParseMode: telebot.ModeMarkdown,
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
