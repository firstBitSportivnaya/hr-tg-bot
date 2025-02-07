/*
MIT License

Copyright (c) 2025 Первый Бит

Данная лицензия разрешает использование, копирование, изменение, слияние, публикацию, распространение,
лицензирование и/или продажу копий программного обеспечения при соблюдении следующих условий:

В вышеуказанном уведомлении об авторских правах и данном уведомлении о разрешении должны быть включены все копии
или значимые части программного обеспечения.

ПРОГРАММНОЕ ОБЕСПЕЧЕНИЕ ПРЕДОСТАВЛЯЕТСЯ "КАК ЕСТЬ", БЕЗ ГАРАНТИЙ ЛЮБОГО РОДА, ЯВНЫХ ИЛИ ПОДРАЗУМЕВАЕМЫХ,
ВКЛЮЧАЯ, НО НЕ ОГРАНИЧИВАЯСЬ, ГАРАНТИЯМИ КОММЕРЧЕСКОЙ ПРИГОДНОСТИ, СООТВЕТСТВИЯ ДЛЯ ОПРЕДЕЛЕННОЙ ЦЕЛИ И
НЕНАРУШЕНИЯ ПРАВ. НИ В КОЕМ СЛУЧАЕ АВТОРЫ ИЛИ ПРАВООБЛАДАТЕЛИ НЕ НЕСУТ ОТВЕТСТВЕННОСТИ ПО ИСКАМ,
УСЛОВИЯМ, ДАМГЕ или другим обязательствам, возникающим из, или в связи с использованием, или иным образом
связанным с данным программным обеспечением.
*/

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
// В данной структуре хранятся ключевые настройки, необходимые для работы Telegram-бота,
// включая токен, режим работы, настройки вебхука, параметры теста и настройки хранилища.
type Config struct {
	Token        string        // Telegram-бот токен, обязательный параметр.
	Mode         string        // Режим работы: "webhook" или "polling". Определяет способ получения обновлений.
	WebhookURL   string        // Публичный URL для вебхуков (используется, если Mode == "webhook").
	ListenAddr   string        // Адрес и порт для прослушивания входящих запросов вебхука.
	PollInterval time.Duration // Интервал для лонгпуллинга (используется, если Mode == "polling").
	Debug        bool          // Флаг отладочного режима; при true включается подробное логирование.

	TestQuestions int           // Количество вопросов в тесте для кандидатов.
	TestDuration  time.Duration // Время, отведенное на прохождение теста (в минутах).

	StorageType string  // Тип хранилища состояния: "memory" для in‑memory или "json" для хранения в файле.
	AdminIDs    []int64 // Список Telegram‑ID администраторов, разделенных запятой.
}

// LoadConfig загружает конфигурацию из файла .env (если он существует) и переменных окружения.
// При успешном выполнении возвращает указатель на структуру Config, в противном случае — ошибку.
func LoadConfig() (*Config, error) {
	// Загружаем переменные окружения из файла .env (если файл существует).
	_ = godotenv.Load()

	// Получаем обязательный токен бота.
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("переменная TELEGRAM_BOT_TOKEN не задана")
	}

	// Определяем режим работы бота (polling или webhook). По умолчанию используется polling.
	mode := os.Getenv("BOT_MODE")
	if mode == "" {
		mode = "polling"
	}

	// Получаем URL для вебхуков (используется при режиме webhook).
	webhookURL := os.Getenv("WEBHOOK_URL")

	// Получаем адрес для прослушивания вебхуков. Если не задан, используется значение по умолчанию ":8443".
	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = ":8443"
	}

	// Устанавливаем интервал для лонгпуллинга. По умолчанию — 2 секунды.
	pollInterval := 2 * time.Second
	if piStr := os.Getenv("POLL_INTERVAL"); piStr != "" {
		if pi, err := strconv.Atoi(piStr); err == nil {
			pollInterval = time.Duration(pi) * time.Second
		}
	}

	// Определяем, включен ли режим отладки.
	debug := false
	if debugStr := os.Getenv("DEBUG"); debugStr != "" {
		if debugStr == "true" || debugStr == "1" {
			debug = true
		}
	}

	// Получаем количество вопросов для теста. По умолчанию — 20 вопросов.
	testQ := 20
	if tqs := os.Getenv("TEST_QUESTIONS"); tqs != "" {
		if n, err := strconv.Atoi(tqs); err == nil {
			testQ = n
		}
	}

	// Получаем время на прохождение теста. По умолчанию — 20 минут.
	testDur := 20 * time.Minute
	if tds := os.Getenv("TEST_DURATION"); tds != "" {
		if n, err := strconv.Atoi(tds); err == nil {
			testDur = time.Duration(n) * time.Minute
		}
	}

	// Определяем тип хранилища состояния: "memory" или "json". Если не задано, используется "memory".
	storageType := os.Getenv("STORAGE_TYPE")
	if storageType == "" {
		storageType = "memory"
	}

	// Обрабатываем список администраторов (Telegram ID), разделенных запятой.
	var adminIDs []int64
	if adminIDsStr := os.Getenv("ADMIN_IDS"); adminIDsStr != "" {
		// Разбиваем строку по запятой и обрабатываем каждую часть.
		parts := strings.Split(adminIDsStr, ",")
		for _, s := range parts {
			s = strings.TrimSpace(s)
			if id, err := strconv.ParseInt(s, 10, 64); err == nil {
				adminIDs = append(adminIDs, id)
			}
		}
	}

	// Возвращаем указатель на структуру Config, заполненную значениями из переменных окружения.
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
