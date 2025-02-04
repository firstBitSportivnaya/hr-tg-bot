package helpers

import (
	"gopkg.in/telebot.v3"
)

// RequireState проверяет, соответствует ли текущее состояние пользователя одному из разрешённых.
// Если состояние не установлено или не совпадает с ни одним из разрешённых, выполнение запрещается.
func RequireState(c telebot.Context, allowed ...string) bool {
	st, ok := c.Get("state").(string)
	if !ok || st == "" {
		return false
	}
	for _, a := range allowed {
		if st == a {
			return true
		}
	}
	_ = c.Send("Вы не можете выполнить это действие на данном этапе. Ваше текущее состояние: " + st)
	return false
}
