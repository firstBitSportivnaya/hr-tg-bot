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

package handlers

import (
	"log"

	"github.com/IT-Nick/config"
	"github.com/IT-Nick/database"
	"github.com/IT-Nick/pending"
	"github.com/IT-Nick/tasks"
	"gopkg.in/telebot.v3"
)

// Глобальные переменные для хранения зависимостей, используемых обработчиками бота.
// Эти переменные инициализируются при вызове RegisterHandlers и используются для доступа к общим ресурсам.
var (
	cfg             *config.Config               // Глобальная конфигурация приложения.
	store           database.Store               // Хранилище состояний пользователей.
	taskManager     *tasks.Manager               // Менеджер тестовых вопросов, используется для получения случайного набора вопросов.
	testAssignStore *pending.TestAssignmentStore // Хранилище для отложенных назначений теста кандидатам.
	roleAssignStore *pending.RoleAssignmentStore // Хранилище для отложенных назначений ролей (например, HR).
)

// RegisterHandlers инициализирует зависимости и регистрирует обработчики команд и callback'ов для Telegram-бота.
// Функция принимает экземпляр бота и хранилище состояний, которое используется для отслеживания сессий пользователей.
func RegisterHandlers(bot *telebot.Bot, s database.Store) {
	// Устанавливаем глобальное хранилище состояний.
	store = s

	var err error

	// Загружаем конфигурацию приложения из переменных окружения и файла .env.
	cfg, err = config.LoadConfig()
	if err != nil {
		log.Fatalf("Не удалось загрузить конфигурацию: %v", err)
	}

	// Инициализируем менеджер тестовых вопросов, загружая данные из файла "data/questions.json".
	taskManager, err = tasks.NewManager("data/questions.json")
	if err != nil {
		log.Fatalf("Не удалось загрузить вопросы: %v", err)
	}

	// Инициализируем хранилища для отложенных назначений теста и ролей.
	testAssignStore = pending.NewTestAssignmentStore("data/test_assignments.json")
	roleAssignStore = pending.NewRoleAssignmentStore("data/role_assignments.json")

	// Регистрируем обработчик команды /start.
	// Этот обработчик отвечает за приветствие пользователя, установку его начального состояния,
	// а также за определение его роли (например, user, HR, admin).
	bot.Handle("/start", startHandler(bot))

	// Регистрируем обработчик текстовых сообщений.
	// Он используется для ввода @username кандидата после активации inline-кнопок для назначения теста или ролей.
	bot.Handle(telebot.OnText, textHandler(bot))

	// Регистрируем обработчик для inline-кнопки "start_test".
	// При активации данной кнопки начинается тестирование: назначается тест и отправляется первый вопрос.
	bot.Handle(&telebot.InlineButton{Unique: "start_test"}, startTestHandler(bot))

	// Регистрируем обработчик для inline-кнопки "answer".
	// Этот обработчик обрабатывает выбор варианта ответа пользователем на тестовый вопрос.
	bot.Handle(&telebot.InlineButton{Unique: "answer"}, answerHandler(bot))

	// Регистрируем обработчики для inline-кнопок назначения теста и назначения роли HR.
	// - assignHandler: используется для назначения теста кандидату пользователем с ролью HR или admin.
	// - assignHRHandler: используется для назначения роли HR кандидату, инициируется администратором.
	bot.Handle(&telebot.InlineButton{Unique: "assign_test"}, assignHandler())
	bot.Handle(&telebot.InlineButton{Unique: "select_test_type"}, selectTestTypeHandler(bot))

	bot.Handle(&telebot.InlineButton{Unique: "assign_hr"}, assignHRHandler())

	bot.Handle(&telebot.InlineButton{Unique: "assign_admin"}, assignAdminHandler())

}
