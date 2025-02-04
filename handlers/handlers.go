package handlers

import (
	"fmt"
	"gopkg.in/telebot.v3"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/IT-Nick/config"
	"github.com/IT-Nick/database"
	//"github.com/IT-Nick/helpers"
	"github.com/IT-Nick/messages"
	"github.com/IT-Nick/pending" // новый пакет для pending-назачений
	"github.com/IT-Nick/report"
	"github.com/IT-Nick/tasks"
)

var (
	cfg             *config.Config
	store           database.Store
	taskManager     *tasks.Manager
	testAssignStore *pending.TestAssignmentStore
	roleAssignStore *pending.RoleAssignmentStore
)

// RegisterHandlers регистрирует обработчики для бота и принимает хранилище состояния.
func RegisterHandlers(bot *telebot.Bot, s database.Store) {
	store = s

	var err error
	cfg, err = config.LoadConfig()
	if err != nil {
		log.Fatalf("Не удалось загрузить конфигурацию: %v", err)
	}

	taskManager, err = tasks.NewManager("data/questions.json")
	if err != nil {
		log.Fatalf("Не удалось загрузить вопросы: %v", err)
	}

	// Инициализируем хранилища для отложенных назначений
	testAssignStore = pending.NewTestAssignmentStore("data/test_assignments.json")
	roleAssignStore = pending.NewRoleAssignmentStore("data/role_assignments.json")

	// Регистрируем обработчики команд и callback
	bot.Handle("/start", startHandler(bot))
	bot.Handle(telebot.OnText, textHandler(bot))
	bot.Handle(&telebot.InlineButton{Unique: "start_test"}, startTestHandler(bot))
	bot.Handle(&telebot.InlineButton{Unique: "answer"}, answerHandler(bot))
	// Для обратной совместимости через команду:
	bot.Handle("/assign", assignHandler(bot))
	bot.Handle("/assign_hr", assignHRHandler(bot))
	// Для inline-кнопок:
	bot.Handle(&telebot.InlineButton{Unique: "assign_hr"}, assignHRHandler(bot))
	bot.Handle(&telebot.InlineButton{Unique: "assign_test"}, assignHandler(bot))
}

// assignHandler – обрабатывает назначение теста кандидату.
// Если вызван через inline-кнопку (c.Callback() != nil),
// то переводит пользователя (HR или admin) в состояние "assign_test_waiting" и просит ввести @username кандидата.
func assignHandler(bot *telebot.Bot) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		sender := c.Sender()
		senderState, ok := store.Get(sender.ID)
		if !ok || (senderState.Role != "hr" && senderState.Role != "admin") {
			return c.Send("У вас нет прав для назначения теста.")
		}
		// Если вызов идёт через inline-кнопку
		if c.Callback() != nil {
			senderState.State = "assign_test_waiting"
			if err := store.Set(sender.ID, senderState); err != nil {
				return err
			}
			return c.Send("Пожалуйста, введите @username кандидата для назначения теста.")
		}
		// Фолбэк: если вызвано как команда с аргументами
		args := c.Args()
		if len(args) < 1 {
			return c.Send("Укажите @username кандидата для назначения теста. Пример: /assign @candidate")
		}
		targetUsername := strings.TrimPrefix(args[0], "@")
		if _, exists, err := testAssignStore.Get(targetUsername); err != nil {
			return c.Send("Ошибка при проверке назначения теста.")
		} else if exists {
			return c.Send(fmt.Sprintf("Кандидату @%s уже назначен тест.", targetUsername))
		}
		newAssignment := pending.TestAssignment{
			CandidateID:       0, // заполняется позже, когда кандидат нажмёт "Начать тест"
			CandidateUsername: targetUsername,
			AssignedBy:        sender.Username,
			AssignedAt:        time.Now(),
		}
		if err := testAssignStore.Set(targetUsername, newAssignment); err != nil {
			return c.Send("Ошибка при назначении теста.")
		}
		return c.Send(fmt.Sprintf("Тест успешно назначен кандидату @%s", targetUsername))
	}
}

// assignHRHandler – обрабатывает назначение роли HR кандидату.
// Если вызван через inline-кнопку, переводит пользователя (только admin) в состояние "assign_hr_waiting"
// и просит ввести @username кандидата.
func assignHRHandler(bot *telebot.Bot) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		sender := c.Sender()
		senderState, ok := store.Get(sender.ID)
		if !ok || senderState.Role != "admin" {
			return c.Send("У вас нет прав для назначения HR.")
		}
		if c.Callback() != nil {
			senderState.State = "assign_hr_waiting"
			if err := store.Set(sender.ID, senderState); err != nil {
				return err
			}
			return c.Send("Пожалуйста, введите @username кандидата для назначения HR.")
		}
		args := c.Args()
		if len(args) < 1 {
			return c.Send("Укажите @username кандидата для назначения HR. Пример: /assign_hr @candidate")
		}
		targetUsername := strings.TrimPrefix(args[0], "@")
		if existing, exists, err := roleAssignStore.Get(targetUsername); err != nil {
			return c.Send("Ошибка при проверке назначения роли.")
		} else if exists {
			return c.Send(fmt.Sprintf("Кандидату @%s уже назначена роль %s.", targetUsername, existing.NewRole))
		}
		newRoleAssign := pending.RoleAssignment{
			CandidateUsername: targetUsername,
			NewRole:           "hr",
			AssignedBy:        sender.Username,
			AssignedAt:        time.Now(),
		}
		if err := roleAssignStore.Set(targetUsername, newRoleAssign); err != nil {
			return c.Send("Ошибка при назначении роли HR.")
		}
		return c.Send(fmt.Sprintf("Роль HR успешно назначена кандидату @%s", targetUsername))
	}
}

// startHandler – обрабатывает команду /start для кандидата, HR и admin.
func startHandler(bot *telebot.Bot) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		user := c.Sender()
		existingState, exists := store.Get(user.ID)
		role := "user"
		if exists {
			role = existingState.Role
		}
		// Если пользователь есть в списке администраторов, роль admin
		for _, id := range cfg.AdminIDs {
			if user.ID == id {
				role = "admin"
				break
			}
		}

		// Если для кандидата отложено назначение роли (например, HR), обновляем роль
		if roleAssign, ok, _ := roleAssignStore.Get(user.Username); ok {
			role = roleAssign.NewRole
			_ = roleAssignStore.Delete(user.Username)
		}

		// Если для кандидата есть назначение теста, переводим его состояние в "assigned"
		stateStr := "welcome"
		if testAssign, ok, _ := testAssignStore.Get(user.Username); ok {
			stateStr = "assigned"
			newState := database.UserState{
				Role:              role,
				State:             stateStr,
				TelegramFirstName: user.FirstName,
				TelegramUsername:  user.Username,
				AssignedBy:        testAssign.AssignedBy,
			}
			if err := store.Set(user.ID, newState); err != nil {
				return err
			}
			_ = testAssignStore.Delete(user.Username)
		} else {
			if err := store.Set(user.ID, database.UserState{
				Role:              role,
				State:             stateStr,
				TelegramFirstName: user.FirstName,
				TelegramUsername:  user.Username,
			}); err != nil {
				return err
			}
		}

		// Формируем приветственное сообщение и кнопки
		welcome := fmt.Sprintf(messages.WelcomeFmt, cfg.TestQuestions, int(cfg.TestDuration.Minutes()))
		rm := &telebot.ReplyMarkup{}
		startTestBtn := telebot.InlineButton{
			Text:   messages.StartTestButton,
			Unique: "start_test",
			Data:   "start",
		}
		rows := [][]telebot.InlineButton{
			{startTestBtn},
		}
		if role == "hr" {
			assignTestBtn := telebot.InlineButton{
				Text:   "Назначить тест кандидату",
				Unique: "assign_test",
				Data:   "assign_test",
			}
			rows = append(rows, []telebot.InlineButton{assignTestBtn})
		} else if role == "admin" {
			assignTestBtn := telebot.InlineButton{
				Text:   "Назначить тест кандидату",
				Unique: "assign_test",
				Data:   "assign_test",
			}
			assignHRBtn := telebot.InlineButton{
				Text:   "Назначить HR",
				Unique: "assign_hr",
				Data:   "assign_hr",
			}
			rows = append(rows, []telebot.InlineButton{assignTestBtn, assignHRBtn})
		}
		rm.InlineKeyboard = rows

		_, err := bot.Send(user, welcome, rm)
		return err
	}
}

// textHandler – обрабатывает текстовые сообщения.
func textHandler(bot *telebot.Bot) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		user := c.Sender()
		us, ok := store.Get(user.ID)
		if !ok {
			return nil
		}
		switch us.State {
		case "assign_test_waiting":
			candidateText := c.Text()
			candidateUsername := strings.TrimSpace(candidateText)
			if !strings.HasPrefix(candidateUsername, "@") {
				candidateUsername = "@" + candidateUsername
			}
			candidateUsername = strings.TrimPrefix(candidateUsername, "@")
			if _, exists, err := testAssignStore.Get(candidateUsername); err != nil {
				c.Send("Ошибка при проверке назначения теста.")
			} else if exists {
				c.Send(fmt.Sprintf("Кандидату @%s уже назначен тест.", candidateUsername))
			} else {
				newAssignment := pending.TestAssignment{
					CandidateID:       0,
					CandidateUsername: candidateUsername,
					AssignedBy:        user.Username,
					AssignedAt:        time.Now(),
				}
				if err := testAssignStore.Set(candidateUsername, newAssignment); err != nil {
					c.Send("Ошибка при назначении теста.")
				} else {
					c.Send(fmt.Sprintf("Тест успешно назначен кандидату @%s", candidateUsername))
				}
			}
			us.State = "welcome"
			store.Set(user.ID, us)
			return nil

		case "assign_hr_waiting":
			candidateText := c.Text()
			candidateUsername := strings.TrimSpace(candidateText)
			if !strings.HasPrefix(candidateUsername, "@") {
				candidateUsername = "@" + candidateUsername
			}
			candidateUsername = strings.TrimPrefix(candidateUsername, "@")
			if existing, exists, err := roleAssignStore.Get(candidateUsername); err != nil {
				c.Send("Ошибка при проверке назначения роли.")
			} else if exists {
				c.Send(fmt.Sprintf("Кандидату @%s уже назначена роль %s.", candidateUsername, existing.NewRole))
			} else {
				newRoleAssign := pending.RoleAssignment{
					CandidateUsername: candidateUsername,
					NewRole:           "hr",
					AssignedBy:        user.Username,
					AssignedAt:        time.Now(),
				}
				if err := roleAssignStore.Set(candidateUsername, newRoleAssign); err != nil {
					c.Send("Ошибка при назначении роли HR.")
				} else {
					c.Send(fmt.Sprintf("Роль HR успешно назначена кандидату @%s", candidateUsername))
				}
			}
			us.State = "welcome"
			store.Set(user.ID, us)
			return nil

		default:
			_, err := bot.Send(user, "Для начала теста нажмите кнопку «Начать тест».")
			return err
		}
	}
}

func startTestHandler(bot *telebot.Bot) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		user := c.Sender()
		candidateUsername := user.Username

		// Ищем назначение по username, т.к. при назначении мы его сохраняли по username
		assignment, exists, err := testAssignStore.Get(candidateUsername)
		if err != nil || !exists {
			return c.Send("Ваш HR менеджер еще не назначил вам тест.")
		}

		// Обновляем поле CandidateID, теперь известен реальный id пользователя
		assignment.CandidateID = user.ID
		if err := testAssignStore.Set(candidateUsername, assignment); err != nil {
			return err
		}

		// Получаем тестовый набор вопросов
		taskSet, err := taskManager.GetRandomTasks(cfg.TestQuestions, strconv.FormatInt(user.ID, 10))
		if err != nil {
			return err
		}
		state := database.UserState{
			Role:              "user",
			State:             "testing",
			CurrentQuestion:   0,
			Score:             0,
			TestTasks:         taskSet,
			Answers:           make(map[int]int),
			TelegramFirstName: user.FirstName,
			TelegramUsername:  user.Username,
			AssignedBy:        assignment.AssignedBy, // сохраняем, кто назначил тест
		}
		if err := store.Set(user.ID, state); err != nil {
			return err
		}
		return sendQuestion(bot, user, 0)
	}
}

// answerHandler – обрабатывает ответы на вопросы кандидата.
func answerHandler(bot *telebot.Bot) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		user := c.Sender()
		// Получаем актуальное состояние пользователя из хранилища
		us, ok := store.Get(user.ID)
		if !ok || us.State != "testing" {
			return c.Send("Вы не можете выполнить это действие на данном этапе.")
		}

		data := c.Callback().Data
		parts := strings.Split(data, "_")
		if len(parts) != 3 {
			return nil
		}
		qIndex, err := strconv.Atoi(parts[1])
		if err != nil {
			return err
		}
		optionIndex, err := strconv.Atoi(parts[2])
		if err != nil {
			return err
		}

		// Сохраняем выбранный ответ
		if us.Answers == nil {
			us.Answers = make(map[int]int)
		}
		us.Answers[qIndex] = optionIndex

		if qIndex >= len(us.TestTasks) {
			return fmt.Errorf("индекс вопроса %d вне диапазона", qIndex)
		}
		task := us.TestTasks[qIndex]
		if optionIndex == task.Answer {
			us.Score++
		}
		us.CurrentQuestion++

		// Если вопросы ещё остались – сохраняем состояние и отправляем следующий вопрос.
		if us.CurrentQuestion < len(us.TestTasks) {
			if err := store.Set(user.ID, us); err != nil {
				return err
			}
			return sendQuestion(bot, user, us.CurrentQuestion)
		}

		// Если вопросы закончились:
		taskManager.ReleaseCandidateTasks(strconv.FormatInt(user.ID, 10))
		finalMsg := fmt.Sprintf(messages.TestFinishedFmt, us.Score, len(us.TestTasks))
		us.State = "finished"
		if err := store.Set(user.ID, us); err != nil {
			return err
		}

		// Формируем отчёт с данными теста
		reportData := report.ReportData{
			UserID:            user.ID,
			TelegramFirstName: user.FirstName,
			TelegramUsername:  user.Username,
			Role:              us.Role,
			State:             us.State,
			Score:             us.Score,
			TotalQuestions:    len(us.TestTasks),
			Answers:           us.Answers,
			TestTasks:         us.TestTasks,
		}
		reportFile, err := report.GeneratePDFReport(reportData)
		if err != nil {
			log.Printf("Ошибка генерации отчета: %v", err)
		} else {
			doc := &telebot.Document{
				File:     telebot.FromDisk(reportFile),
				FileName: reportFile,
			}
			if _, err = bot.Send(user, doc); err != nil {
				log.Printf("Ошибка отправки отчета: %v", err)
			}
		}

		// Удаляем назначение теста из pending-хранилища по username
		if err := testAssignStore.Delete(user.Username); err != nil {
			log.Printf("Ошибка удаления назначения теста: %v", err)
		}

		_, err = bot.Send(user, finalMsg)
		return err
	}
}

// sendQuestion отправляет кандидату вопрос с вариантами ответов.
func sendQuestion(bot *telebot.Bot, user *telebot.User, qIndex int) error {
	us, ok := store.Get(user.ID)
	if !ok || qIndex >= len(us.TestTasks) {
		return nil
	}
	task := us.TestTasks[qIndex]
	questionText := fmt.Sprintf("Вопрос %d:\n%s", qIndex+1, task.Text)
	rm := &telebot.ReplyMarkup{}
	var buttons []telebot.InlineButton
	for i, opt := range task.Options {
		data := fmt.Sprintf("answer_%d_%d", qIndex, i)
		btn := telebot.InlineButton{
			Text:   opt,
			Unique: "answer",
			Data:   data,
		}
		buttons = append(buttons, btn)
	}
	rm.InlineKeyboard = [][]telebot.InlineButton{buttons}
	_, err := bot.Send(user, questionText, rm)
	return err
}
