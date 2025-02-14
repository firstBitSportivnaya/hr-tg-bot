package middleware

import (
	"fmt"

	"github.com/IT-Nick/database"
	"gopkg.in/telebot.v3"
)

// DebugUserActions возвращает middleware, которое при включённом режиме отладки отправляет пользователю
// отладочное сообщение, содержащее информацию о пользователе: имя, ID, роль, текущее состояние и описание действия.
// Это может быть полезно для отслеживания и диагностики поведения пользователей во время разработки или отладки.
func DebugUserActions(enabled bool) telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			// Вызываем следующий обработчик и сохраняем его ошибку.
			err := next(c)
			// Если режим отладки включен, формируем отладочное сообщение.
			if enabled {
				user := c.Sender()
				// Инициализируем переменные для хранения роли и состояния пользователя.
				role, stateStr := "", ""
				if database.GlobalStore != nil {
					if us, ok := database.GlobalStore.Get(user.ID); ok {
						role = us.Role
						stateStr = us.State
					}
				}
				// Определяем тип действия пользователя (текстовое сообщение, callback или неизвестное действие).
				var action string
				if msg := c.Message(); msg != nil {
					action = "Message: " + msg.Text
				} else if cb := c.Callback(); cb != nil {
					action = "Callback: " + cb.Data
				} else {
					action = "Unknown action"
				}
				// Формируем строку отладочного сообщения.
				debugMsg := fmt.Sprintf("DEBUG: User: %s (ID: %d), Role: %s, State: %s, Action: %s",
					user.FirstName, user.ID, role, stateStr, action)
				// Отправляем отладочное сообщение в отдельной горутине, чтобы не блокировать основное выполнение.
				go c.Bot().Send(user, debugMsg)
			}
			// Возвращаем результат выполнения следующего обработчика.
			return err
		}
	}
}
