package middleware

import tele "gopkg.in/telebot.v3"

// AutoRespond автоматически отвечает на callback-запросы.
func AutoRespond() tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			if c.Callback() != nil {
				defer c.Respond()
			}
			return next(c)
		}
	}
}
