package middlewares

import (
	"fmt"
	"net/http"
	"time"
)

type TimerMiddleware struct {
	h http.Handler
}

func NewTimerMiddleware(h http.Handler) http.Handler {
	return &TimerMiddleware{h: h}
}

func (m *TimerMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func(now time.Time) {
		fmt.Printf("%s spent %s", r.URL.String(), time.Since(now))
	}(time.Now())

	m.h.ServeHTTP(w, r)
}
