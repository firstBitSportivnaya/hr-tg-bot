package poller

import (
	"log"
	"net/http"

	"github.com/IT-Nick/config"
	"gopkg.in/telebot.v3"
)

// NewPoller создаёт Poller в зависимости от режима.
func NewPoller(cfg *config.Config) telebot.Poller {
	if cfg.Mode == "webhook" {
		if cfg.WebhookURL == "" {
			log.Fatalf("В режиме webhook переменная WEBHOOK_URL должна быть задана")
		}
		return &telebot.Webhook{
			Listen: cfg.ListenAddr,
			Endpoint: &telebot.WebhookEndpoint{
				PublicURL: cfg.WebhookURL,
			},
		}
	}
	return &telebot.LongPoller{Timeout: cfg.PollInterval}
}

// StartHTTPServer запускает HTTP-сервер для вебхуков.
func StartHTTPServer(cfg *config.Config, handler http.Handler) error {
	server := &http.Server{
		Addr:    cfg.ListenAddr,
		Handler: handler,
	}
	return server.ListenAndServe()
}
