package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config содержит параметры конфигурации приложения.
type Config struct {
	Token        string        // Telegram-бот токен
	Mode         string        // Режим работы: "webhook" или "polling"
	WebhookURL   string        // Публичный URL для вебхуков (используется при webhook)
	ListenAddr   string        // Адрес и порт для прослушивания вебхуков
	PollInterval time.Duration // Интервал для лонгпуллинга
	Debug        bool          // Включение отладочного логирования

	TestQuestions int           // Количество вопросов в тесте
	TestDuration  time.Duration // Время на прохождение теста (в минутах)

	StorageType string  // Тип хранения состояния: "memory" или "json"
	AdminIDs    []int64 // Список Telegram‑ID администраторов (через запятую)
}

// LoadConfig загружает конфигурацию из .env и переменных окружения.
func LoadConfig() (*Config, error) {
	// Загружаем .env (если существует)
	_ = godotenv.Load()

	// Обязательный токен.
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("переменная TELEGRAM_BOT_TOKEN не задана")
	}

	// Режим работы (polling или webhook).
	mode := os.Getenv("BOT_MODE")
	if mode == "" {
		mode = "polling"
	}

	// URL для вебхуков.
	webhookURL := os.Getenv("WEBHOOK_URL")

	// Адрес для прослушивания вебхуков.
	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = ":8443"
	}

	// Интервал для лонгпуллинга.
	pollInterval := 2 * time.Second
	if piStr := os.Getenv("POLL_INTERVAL"); piStr != "" {
		if pi, err := strconv.Atoi(piStr); err == nil {
			pollInterval = time.Duration(pi) * time.Second
		}
	}

	// Отладочный режим.
	debug := false
	if debugStr := os.Getenv("DEBUG"); debugStr != "" {
		if debugStr == "true" || debugStr == "1" {
			debug = true
		}
	}

	// Количество вопросов в тесте.
	testQ := 20
	if tqs := os.Getenv("TEST_QUESTIONS"); tqs != "" {
		if n, err := strconv.Atoi(tqs); err == nil {
			testQ = n
		}
	}

	// Время теста (в минутах).
	testDur := 20 * time.Minute
	if tds := os.Getenv("TEST_DURATION"); tds != "" {
		if n, err := strconv.Atoi(tds); err == nil {
			testDur = time.Duration(n) * time.Minute
		}
	}

	// Тип хранения состояния.
	storageType := os.Getenv("STORAGE_TYPE")
	if storageType == "" {
		storageType = "memory"
	}

	// Список администраторов (Telegram ID) через запятую.
	var adminIDs []int64
	if adminIDsStr := os.Getenv("ADMIN_IDS"); adminIDsStr != "" {
		parts := strings.Split(adminIDsStr, ",")
		for _, s := range parts {
			s = strings.TrimSpace(s)
			if id, err := strconv.ParseInt(s, 10, 64); err == nil {
				adminIDs = append(adminIDs, id)
			}
		}
	}

	return &Config{
		Token:         token,
		Mode:          mode,
		WebhookURL:    webhookURL,
		ListenAddr:    listenAddr,
		PollInterval:  pollInterval,
		Debug:         debug,
		TestQuestions: testQ,
		TestDuration:  testDur,
		StorageType:   storageType,
		AdminIDs:      adminIDs,
	}, nil
}
