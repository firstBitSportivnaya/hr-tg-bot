package middleware

import (
	"errors"
	"log"

	tele "gopkg.in/telebot.v3"
)

// Recover возвращает middleware-функцию, которая перехватывает панику, возникшую в обработчике, и вызывает
// заданный обработчик ошибки. Если пользователь не передал свою функцию обработки ошибки, используется
// функция по умолчанию, которая логирует сообщение о панике.
//
// Параметры:
//   - onError: опциональный параметр, представляющий функцию, принимающую ошибку и контекст телеграм-бота.
//     Она вызывается, если в обработчике произошла паника.
//
// Принцип работы:
//  1. Оборачивается вызов следующего обработчика в defer-функцию.
//  2. Если происходит panic, defer-функция перехватывает значение panic (r) и преобразует его в объект error.
//  3. Вызывается функция обработки ошибки (handleError) с полученной ошибкой и текущим контекстом.
//  4. Возвращаемая ошибка устанавливается равной преобразованной ошибке, чтобы middleware корректно завершился.
func Recover(onError ...func(error, tele.Context)) tele.MiddlewareFunc {
	var handleError func(error, tele.Context)
	// Если пользователь передал свою функцию обработки ошибки, используем её.
	if len(onError) > 0 {
		handleError = onError[0]
	} else {
		// Иначе, используем функцию по умолчанию, которая логирует информацию о панике.
		handleError = func(err error, c tele.Context) {
			log.Printf("Recovered from panic: %v", err)
		}
	}

	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) (err error) {
			// Данный defer перехватывает паники, возникающие в следующем обработчике.
			defer func() {
				if r := recover(); r != nil {
					var e error
					// Преобразуем значение panic в объект error.
					switch x := r.(type) {
					case error:
						e = x
					case string:
						e = errors.New(x)
					default:
						e = errors.New("unknown panic")
					}
					// Вызываем обработчик ошибки с полученной ошибкой и контекстом.
					handleError(e, c)
					// Передаем ошибку дальше, чтобы цепочка обработчиков могла корректно завершиться.
					err = e
				}
			}()
			// Вызываем следующий обработчик в цепочке.
			return next(c)
		}
	}
}
