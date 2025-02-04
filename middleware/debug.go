package middleware

import (
	"fmt"
	"github.com/IT-Nick/database"
	"gopkg.in/telebot.v3"
)

// DebugUserActions возвращает middleware, которое при включённом режиме отладки
// отправляет пользователю отладочное сообщение с информацией: имя, ID, роль, состояние и описание действия.
func DebugUserActions(enabled bool) telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			err := next(c)
			if enabled {
				user := c.Sender()
				role, stateStr := "", ""
				if database.GlobalStore != nil {
					if us, ok := database.GlobalStore.Get(user.ID); ok {
						role = us.Role
						stateStr = us.State
					}
				}
				var action string
				if msg := c.Message(); msg != nil {
					action = "Message: " + msg.Text
				} else if cb := c.Callback(); cb != nil {
					action = "Callback: " + cb.Data
				} else {
					action = "Unknown action"
				}
				debugMsg := fmt.Sprintf("DEBUG: User: %s (ID: %d), Role: %s, State: %s, Action: %s",
					user.FirstName, user.ID, role, stateStr, action)
				go c.Bot().Send(user, debugMsg)
			}
			return err
		}
	}
}
