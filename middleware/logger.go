package middleware

import (
	"encoding/json"
	"log"

	tele "gopkg.in/telebot.v3"
)

// Logger возвращает middleware, которое логирует входящие обновления Telegram.
// Функция принимает неограниченное число параметров типа *log.Logger. Если передан хотя бы один логгер,
// используется он, иначе применяется логгер по умолчанию (log.Default()).
// Middleware сериализует обновление (Update) в формат JSON с отступами и выводит результат через выбранный логгер.
func Logger(logger ...*log.Logger) tele.MiddlewareFunc {
	var l *log.Logger
	if len(logger) > 0 {
		l = logger[0]
	} else {
		l = log.Default()
	}
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			// Сериализуем входящее обновление в формат JSON с отступами.
			data, _ := json.MarshalIndent(c.Update(), "", "  ")
			// Выводим сериализованные данные в лог.
			l.Println(string(data))
			// Передаем управление следующему обработчику в цепочке.
			return next(c)
		}
	}
}
