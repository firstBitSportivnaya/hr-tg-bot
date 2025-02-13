package main

import (
	"log"
	"os"

	"github.com/IT-Nick/config"
	"github.com/IT-Nick/database"
	"github.com/IT-Nick/handlers"
	"github.com/IT-Nick/helpers"
	"github.com/IT-Nick/middleware"
	"github.com/IT-Nick/poller"
	"gopkg.in/telebot.v3"
)

// main является точкой входа в приложение Telegram-бота.
// Программа настраивает бота, регистрирует middleware, хранилище состояний и обработчики команд.
// Затем бот запускается в выбранном режиме (polling или webhook).
func main() {
	// Загружаем конфигурацию из переменных окружения или .env файла.
	// Если загрузка конфигурации завершается ошибкой, приложение завершает работу.
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Формируем настройки для бота, включая токен и механизм получения обновлений (poller).
	settings := telebot.Settings{
		Token:  cfg.Token,
		Poller: poller.NewPoller(cfg),
	}

	// Создаем экземпляр бота с заданными настройками.
	bot, err := telebot.NewBot(settings)
	if err != nil {
		log.Fatalf("Не удалось создать бота: %v", err)
	}

	// Инициализируем кастомный логгер для вывода логов, связанных с ботом.
	// Логгер выводит сообщения в стандартный вывод (stdout) с префиксом "[bot]".
	customLogger := log.New(os.Stdout, "[bot] ", log.LstdFlags)

	// Регистрируем цепочку middleware для обработки входящих обновлений.
	// Если включен режим отладки (DEBUG), добавляем:
	// - middleware для логирования обновлений (Logger),
	// - middleware для отладочного логирования действий пользователя (DebugUserActions).
	if cfg.Debug {
		bot.Use(middleware.Logger(customLogger))
		bot.Use(middleware.DebugUserActions(true))
	}

	// Регистрируем middleware, которые всегда применяются:
	// - AutoRespond: автоматически отвечает на callback-запросы, предотвращая "зависание" Telegram клиента.
	// - Recover: перехватывает паники в обработчиках, чтобы приложение не завершалось аварийно.
	bot.Use(
		middleware.AutoRespond(),
		middleware.Recover(),
	)

	// Создаем хранилище состояний пользователей.
	// В зависимости от конфигурации, если STORAGE_TYPE == "json" используется JSONStore (сохранение в файл),
	// иначе применяется in‑memory хранилище.
	store := database.NewStore(cfg.StorageType, "data/states.json")
	// Устанавливаем глобальное хранилище для доступа из middleware и других модулей.
	database.GlobalStore = store

	// Регистрируем обработчики команд и callback'ов.
	// Обработчики находятся в пакете handlers и получают экземпляр бота и хранилище состояний.
	handlers.RegisterHandlers(bot, store)

	// Если используется JSON-хранилище, запускаем восстановление активных таймеров.
	// Необходимо для возобновления работы сессий кандидатов после перезапуска приложения, без обнуления таймера.
	if cfg.StorageType == "json" {
		go helpers.RestoreActiveTimers(bot, cfg, store, handlers.FinishTest)
	}

	// Выводим информационное сообщение о запуске бота в заданном режиме (polling или webhook).
	log.Printf("Запуск бота в режиме %s...", cfg.Mode)
	// Запускаем основной цикл получения обновлений бота.
	bot.Start()
}
