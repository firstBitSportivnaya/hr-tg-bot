package app

import (
	"fmt"
	"github.com/IT-Nick/internal/app/handlers/http/active_tests_handler"
	"github.com/IT-Nick/internal/app/handlers/http/update_user_role_handler"
	"github.com/IT-Nick/internal/app/handlers/http/user_test_report_handler"
	"github.com/IT-Nick/internal/app/handlers/telegram/answer_handler"
	"github.com/IT-Nick/internal/app/handlers/telegram/assign_tests/assign_handler"
	"github.com/IT-Nick/internal/app/handlers/telegram/assign_tests/assign_next_page_handler"
	"github.com/IT-Nick/internal/app/handlers/telegram/assign_tests/assign_prev_page_handler"
	"github.com/IT-Nick/internal/app/handlers/telegram/assign_tests/assign_test_handler"
	"github.com/IT-Nick/internal/app/handlers/telegram/assign_tests/select_test_handler"
	"github.com/IT-Nick/internal/app/handlers/telegram/start_handler"
	"github.com/IT-Nick/internal/app/handlers/telegram/start_test_handler"
	msgRepo "github.com/IT-Nick/internal/domain/messages/repository"
	msgService "github.com/IT-Nick/internal/domain/messages/service"
	rolesRepo "github.com/IT-Nick/internal/domain/roles/repository"
	rolesService "github.com/IT-Nick/internal/domain/roles/service"
	testsRepo "github.com/IT-Nick/internal/domain/tests/repository"
	testsService "github.com/IT-Nick/internal/domain/tests/service"
	"github.com/IT-Nick/internal/domain/users/repository"
	"github.com/IT-Nick/internal/domain/users/service"
	"github.com/IT-Nick/internal/infra/config"
	"github.com/IT-Nick/internal/infra/timer"
	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/telebot.v4"
	"log"
	"net/http"
	"strings"
	"time"
)

type LocalStatesHelpers struct {
	pageState       map[int64]int
	assignTestState map[int64]int
}

type Services struct {
	userService    *service.UserService
	messageService *msgService.MessageService
	roleService    *rolesService.RoleService
	testService    *testsService.TestService
}

type App struct {
	config       *config.Config
	bot          *telebot.Bot
	db           *pgxpool.Pool
	server       *http.Server
	timerUpdater *timer.Updater

	Services
	states LocalStatesHelpers
}

func NewApp(configPath string) (*App, error) {

	log.Println("Local time:", time.Now())
	log.Println("UTC time:", time.Now().UTC())
	log.Println("Local timezone:", time.Local)

	configImpl, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("config.LoadConfig: %w", err)
	}

	db, err := InitDatabase(configImpl)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	app := &App{
		config: configImpl,
		db:     db,
		states: LocalStatesHelpers{
			pageState:       make(map[int64]int),
			assignTestState: make(map[int64]int),
		},
	}

	app.initServices()
	return app, nil
}

// Функция для инициализации сервисов и репозиториев
func (app *App) initServices() {
	// Инициализация репозиториев
	userRepo := repository.NewUserRepository(app.db)
	messageRepo := msgRepo.NewMessageRepository(app.db)
	rolePermissionRepo := rolesRepo.NewRolePermissionRepository(app.db)
	testRepo := testsRepo.NewTestRepository(app.db)

	// Инициализация сервисов
	app.userService = service.NewUserService(userRepo, rolePermissionRepo)
	app.messageService = msgService.NewMessageService(messageRepo)
	app.roleService = rolesService.NewRoleService(rolePermissionRepo)
	app.testService = testsService.NewTestService(testRepo, userRepo)
}

// ListenAndServeTelegram запускает сервер Telegram бота
func (app *App) ListenAndServeTelegram() error {
	bot, err := telebot.NewBot(telebot.Settings{
		Token:  app.config.TelegramBot.Token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		return fmt.Errorf("telebot.NewBot: %w", err)
	}
	app.bot = bot

	app.timerUpdater = timer.NewTimerUpdater(app.bot, app.testService)

	app.bootstrapHandlersTelegram()

	go app.bot.Start()

	return nil
}

// bootstrapHandlersTelegram - регистрирует обработчики для бота
func (app *App) bootstrapHandlersTelegram() {
	app.bot.Handle("/start",
		start_handler.NewStartHandler(
			app.userService,
			app.messageService,
			app.roleService,
			app.testService,
		).GetHandlerFunc())

	// Обработчики назначения теста кандидату (с обработчиками пагинации).
	// OnCallback обработчик принимает айди теста.
	app.bot.Handle(&telebot.InlineButton{Unique: "assign_test"},
		assign_handler.NewAssignStartPageHandler(
			app.userService,
			app.testService,
			app.states.pageState,
		).GetHandlerFunc())
	app.bot.Handle(&telebot.InlineButton{Unique: "next_page"},
		assign_next_page_handler.NewAssignNextPageHandler(
			app.userService,
			app.testService,
			app.states.pageState,
		).GetHandlerFunc())
	app.bot.Handle(&telebot.InlineButton{Unique: "prev_page"},
		assign_prev_page_handler.NewAssignPrevPageHandler(
			app.userService,
			app.testService,
			app.states.pageState,
		).GetHandlerFunc())
	app.bot.Handle(&telebot.InlineButton{Unique: "start_page"}, func(c telebot.Context) error {
		if c.Sender() != nil {
			return c.Send("Вы находитесь в начале списка.")
		}
		return nil
	})
	app.bot.Handle(&telebot.InlineButton{Unique: "end_page"}, func(c telebot.Context) error {
		if c.Sender() != nil {
			return c.Send("Вы находитесь в конце списка.")
		}
		return nil
	})
	app.bot.Handle(telebot.OnCallback, func(c telebot.Context) error {
		data := c.Callback().Data

		cleanedData := strings.TrimSpace(data)
		cleanedData = strings.ReplaceAll(cleanedData, "\f", "")
		cleanedData = strings.ReplaceAll(cleanedData, "\\f", "")

		// Проверяем, начинается ли callback с "test_"
		if strings.HasPrefix(cleanedData, "test_") {
			return select_test_handler.NewSelectTestHandler(
				app.userService,
				app.testService,
				app.states.assignTestState).Handle(c)
		}

		// Проверяем callback для ответа на вопрос
		if strings.HasPrefix(cleanedData, "answer_") {
			return answer_handler.NewAnswerHandler(app.bot, app.testService).Handle(c)
		}

		return nil
	})

	app.bot.Handle(telebot.OnText,
		assign_test_handler.NewAssignTestHandler(
			app.userService,
			app.testService,
			app.states.assignTestState,
		).GetHandlerFunc())

	// Обработчик запуска теста (с логикой нахождения назначенных тестов кандидату)
	app.bot.Handle(&telebot.InlineButton{Unique: "start_test"},
		start_test_handler.NewStartTestHandler(
			app.bot,
			app.testService,
			app.messageService,
			app.userService,
			app.timerUpdater,
		).GetHandlerFunc())
}

// ListenAndServeHTTP запускает HTTP сервер
func (app *App) ListenAndServeHTTP() error {
	mx := http.NewServeMux()

	mx.Handle("POST /users/update-role", update_user_role_handler.NewUpdateUserRoleHandler(app.userService, app.roleService))
	mx.Handle("POST /reports/user", user_test_report_handler.NewUserTestReportHandler(app.userService, app.testService))
	mx.Handle("GET /reports/active-tests", active_tests_handler.NewActiveTestsHandler(app.userService, app.testService))

	app.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%s", app.config.Server.Host, app.config.Server.Port),
		Handler: mx,
	}

	return app.server.ListenAndServe()
}

// ListenAndServe запускает оба сервера (Telegram и HTTP)
func (app *App) ListenAndServe() error {
	// Запускаем Telegram сервер
	if err := app.ListenAndServeTelegram(); err != nil {
		return fmt.Errorf("failed to start Telegram bot: %w", err)
	}

	// Запускаем HTTP сервер
	if err := app.ListenAndServeHTTP(); err != nil {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	return nil
}
