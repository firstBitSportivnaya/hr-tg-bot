package middleware

import tele "gopkg.in/telebot.v3"

// AutoRespond возвращает middleware-функцию, которая автоматически отвечает на callback-запросы.
// Это необходимо для того, чтобы Telegram не показывал пользователю сообщение о том, что запрос не был обработан,
// что помогает избежать "зависания" интерфейса в Telegram-клиентах.
//
// Функция работает следующим образом:
//  1. Проверяется, является ли входящее обновление callback-запросом.
//  2. Если да, то после выполнения следующего обработчика автоматически вызывается метод Respond()
//     для отправки ответа на callback-запрос.
//  3. Если обновление не является callback-запросом, middleware ничего не делает.
func AutoRespond() tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			// Если обновление является callback-запросом, откладываем вызов метода Respond(),
			// чтобы автоматически отправить ответ после выполнения следующего обработчика.
			if c.Callback() != nil {
				defer c.Respond()
			}
			// Передаем управление следующему обработчику в цепочке.
			return next(c)
		}
	}
}
