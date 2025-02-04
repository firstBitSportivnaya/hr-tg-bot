package middleware

import (
	"errors"
	"log"

	tele "gopkg.in/telebot.v3"
)

// Recover перехватывает панику в обработчике и вызывает обработчик ошибки.
func Recover(onError ...func(error, tele.Context)) tele.MiddlewareFunc {
	var handleError func(error, tele.Context)
	if len(onError) > 0 {
		handleError = onError[0]
	} else {
		handleError = func(err error, c tele.Context) {
			log.Printf("Recovered from panic: %v", err)
		}
	}
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					var e error
					switch x := r.(type) {
					case error:
						e = x
					case string:
						e = errors.New(x)
					default:
						e = errors.New("unknown panic")
					}
					handleError(e, c)
					err = e
				}
			}()
			return next(c)
		}
	}
}
