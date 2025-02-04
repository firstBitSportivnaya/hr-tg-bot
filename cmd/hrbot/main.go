package main

import (
	"log"
	"os"

	"github.com/IT-Nick/config"
	"github.com/IT-Nick/database"
	"github.com/IT-Nick/handlers"
	"github.com/IT-Nick/middleware"
	"github.com/IT-Nick/poller"
	"gopkg.in/telebot.v3"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	settings := telebot.Settings{
		Token:  cfg.Token,
		Poller: poller.NewPoller(cfg),
	}

	bot, err := telebot.NewBot(settings)
	if err != nil {
		log.Fatalf("Не удалось создать бота: %v", err)
	}

	// Инициализируем кастомный логгер.
	customLogger := log.New(os.Stdout, "[bot] ", log.LstdFlags)

	// Регистрируем цепочку middleware:
	// Если режим отладки включён, добавляем логгер и отладочное логирование действий пользователя.
	if cfg.Debug {
		bot.Use(middleware.Logger(customLogger))
		bot.Use(middleware.DebugUserActions(true))
	}
	bot.Use(
		middleware.AutoRespond(),
		middleware.Recover(),
	)

	// Создаем хранилище через модуль database.
	// Если STORAGE_TYPE == "json", используется файл "data/states.json", иначе in‑memory.
	store := database.NewStore(cfg.StorageType, "data/states.json")
	// Устанавливаем глобальную переменную для доступа из middleware (опционально).
	database.GlobalStore = store

	handlers.RegisterHandlers(bot, store)

	log.Printf("Запуск бота в режиме %s...", cfg.Mode)
	bot.Start()
}
