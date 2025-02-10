package handlers

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/IT-Nick/database"
	"github.com/IT-Nick/helpers"
	"github.com/IT-Nick/messages"
	"github.com/IT-Nick/report"
	"gopkg.in/telebot.v3"
)

// startTestHandler обрабатывает запуск теста кандидатом.
// Функция проверяет наличие назначения теста, обновляет состояние пользователя, инициирует резервирование
// вопросов и запускает таймер, после чего отправляет первый вопрос кандидату.
func startTestHandler(bot *telebot.Bot) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		// Получаем информацию о пользователе, который запускает тест.
		user := c.Sender()
		candidateUsername := user.Username

		// Проверяем наличие отложенного назначения теста для данного кандидата.
		assignment, exists, err := testAssignStore.Get(candidateUsername)
		if err != nil || !exists {
			// Если тест не назначен, уведомляем кандидата об отсутствии назначения.
			return c.Send("Ваш HR менеджер еще не назначил вам тест.")
		}
		// Привязываем назначение к текущему пользователю.
		assignment.CandidateID = user.ID

		// Обновляем запись назначения теста (если требуется) и затем удаляем её,
		// так как тест фактически начинается.
		if err := testAssignStore.Set(candidateUsername, assignment); err != nil {
			return err
		}
		_ = testAssignStore.Delete(candidateUsername)

		// Извлекаем текущее состояние пользователя из хранилища.
		us, stateExists := store.Get(user.ID)
		// По умолчанию роль пользователя — "user".
		role := "user"
		// Если состояние существует и роль пользователя равна "admin" или "hr", используем её.
		if stateExists && (us.Role == "admin" || us.Role == "hr") {
			role = us.Role
		} else {
			// Если пользователь является администратором согласно конфигурации, назначаем ему роль "admin".
			for _, id := range cfg.AdminIDs {
				if user.ID == id {
					role = "admin"
					break
				}
			}
		}

		// Получаем набор вопросов для кандидата с учетом выбранного типа теста.
		// cfg.TestQuestions – общее число вопросов для теста, а assignment.TestType содержит выбранный тип (например, "logic").
		taskSet, err := taskManager.GetRandomTasks(cfg.TestQuestions, strconv.FormatInt(user.ID, 10), assignment.TestType)
		if err != nil {
			return err
		}

		// Инициализируем новое состояние пользователя для начала тестирования.
		newState := database.UserState{
			Role:              role,
			State:             "testing",
			CurrentQuestion:   0,
			Score:             0,
			TestTasks:         taskSet,
			Answers:           make(map[int]int),
			TelegramFirstName: user.FirstName,
			TelegramUsername:  user.Username,
			AssignedByID:      assignment.AssignedByID,
			AssignedBy:        assignment.AssignedBy,
			TestType:          assignment.TestType,
		}
		// Сохраняем новое состояние пользователя.
		if err := store.Set(user.ID, newState); err != nil {
			return err
		}

		// Если запуск теста инициирован через callback (нажатие inline-кнопки),
		// удаляем исходное сообщение для уменьшения засорения чата.
		if c.Callback() != nil {
			_ = c.Delete()
		}

		// Запускаем таймер теста, который периодически обновляет сообщение с оставшимся временем и информацией о прогрессе.
		helpers.StartTimerMessage(bot, user, cfg, func() string {
			st, ok := store.Get(user.ID)
			if !ok {
				return ""
			}
			return fmt.Sprintf("Время: %s\nВопрос %d из %d",
				helpers.RemainingTimeStr(st.TimerDeadline),
				st.CurrentQuestion+1,
				len(st.TestTasks))
		}, func() {
			// По истечении времени таймера вызывается завершение теста.
			if err := FinishTest(bot, user); err != nil {
				fmt.Fprintf(os.Stderr, "Ошибка завершения теста userID=%d: %v\n", user.ID, err)
			}
		})

		// Отправляем первый вопрос кандидату.
		if err := sendQuestion(bot, user, newState.CurrentQuestion); err != nil {
			log.Printf("Ошибка отправки вопроса: %v", err)
			return err
		}
		return nil
	}
}

// answerHandler обрабатывает ответы кандидата на тестовые вопросы.
// Функция получает данные callback'а, определяет выбранный вариант, обновляет состояние пользователя,
// а затем либо отправляет следующий вопрос, либо завершает тест.
func answerHandler(bot *telebot.Bot) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		// Получаем пользователя и его текущее состояние.
		user := c.Sender()
		us, ok := store.Get(user.ID)
		if !ok || us.State != "testing" {
			return c.Send("Вы не можете выполнить это действие на данном этапе.")
		}

		// Извлекаем данные из callback'а и разбиваем их на части.
		data := c.Callback().Data
		parts := strings.Split(data, "_")
		if len(parts) != 3 {
			return nil
		}

		// Преобразуем индексы вопроса и выбранного варианта ответа из строки в целое число.
		qIndex, err := strconv.Atoi(parts[1])
		if err != nil {
			return err
		}
		optionIndex, err := strconv.Atoi(parts[2])
		if err != nil {
			return err
		}

		// Инициализируем карту ответов, если она отсутствует, и сохраняем выбранный вариант ответа для текущего вопроса.
		if us.Answers == nil {
			us.Answers = make(map[int]int)
		}
		us.Answers[qIndex] = optionIndex

		// Проверяем, что индекс вопроса находится в пределах доступного диапазона.
		if qIndex >= len(us.TestTasks) {
			return fmt.Errorf("индекс вопроса %d вне диапазона", qIndex)
		}
		// Получаем текущий вопрос.
		task := us.TestTasks[qIndex]
		// Если выбранный вариант совпадает с правильным ответом, увеличиваем счет.
		if optionIndex == task.Answer {
			us.Score++
		}

		// Сохраняем идентификатор сообщения с вопросом для последующего удаления.
		oldMsgID := us.QuestionMessageID
		// Увеличиваем счетчик текущего вопроса.
		us.CurrentQuestion++
		// Обновляем состояние пользователя в хранилище.
		if err := store.Set(user.ID, us); err != nil {
			return err
		}

		// Удаляем предыдущее сообщение с вопросом, если оно существует.
		if oldMsgID != 0 {
			_ = bot.Delete(&telebot.Message{
				ID:   oldMsgID,
				Chat: &telebot.Chat{ID: user.ID},
			})
		}

		// Если остаются еще вопросы, отправляем следующий вопрос.
		if us.CurrentQuestion < len(us.TestTasks) {
			if err := sendQuestion(bot, user, us.CurrentQuestion); err != nil {
				return err
			}
			return nil
		}
		// Если вопросы закончились, завершаем тест.
		return FinishTest(bot, user)
	}
}

// sendQuestion отправляет пользователю тестовый вопрос с набором вариантов ответов в виде inline-кнопок.
// Функция формирует текст вопроса, создает клавиатуру с вариантами и сохраняет ID отправленного сообщения для дальнейшего управления.
func sendQuestion(bot *telebot.Bot, user *telebot.User, qIndex int) error {
	// Получаем текущее состояние пользователя.
	us, ok := store.Get(user.ID)
	if !ok || qIndex >= len(us.TestTasks) {
		return nil
	}

	// Извлекаем тестовый вопрос по индексу.
	task := us.TestTasks[qIndex]
	// Формируем текст вопроса с номером.
	questionText := fmt.Sprintf("Вопрос %d:\n%s", qIndex+1, task.Text)
	// Инициализируем объект разметки для inline-кнопок.
	rm := &telebot.ReplyMarkup{}
	var buttons []telebot.InlineButton
	// Для каждого варианта ответа формируем кнопку с данными callback'а.
	for i, opt := range task.Options {
		data := fmt.Sprintf("answer_%d_%d", qIndex, i)
		btn := telebot.InlineButton{
			Text:   opt,
			Unique: "answer",
			Data:   data,
		}
		buttons = append(buttons, btn)
	}
	// Устанавливаем кнопки в одну строку клавиатуры.
	rm.InlineKeyboard = [][]telebot.InlineButton{buttons}

	// Отправляем сообщение с вопросом и inline-клавиатурой.
	msg, err := bot.Send(user, questionText, rm)
	if err != nil {
		return err
	}

	// Сохраняем ID сообщения с вопросом в состоянии пользователя для последующего удаления.
	us.QuestionMessageID = msg.ID
	_ = store.Set(user.ID, us)
	return nil
}

// FinishTest завершает тестирование пользователя.
// Функция обновляет состояние, освобождает зарезервированные вопросы, генерирует PDF-отчёт,
// отправляет итоговое сообщение и, при необходимости, пересылает отчёт назначившему HR/admin.
func FinishTest(bot *telebot.Bot, user *telebot.User) error {
	// Получаем текущее состояние пользователя.
	us, ok := store.Get(user.ID)
	if !ok {
		return nil
	}
	// Освобождаем вопросы, зарезервированные для данного пользователя.
	taskManager.ReleaseCandidateTasks(strconv.FormatInt(user.ID, 10))

	// Формируем итоговое сообщение по завершении теста.
	finalMsg := fmt.Sprint(messages.TestFinishedFmt)
	// Обновляем состояние пользователя на "finished".
	us.State = "finished"
	if err := store.Set(user.ID, us); err != nil {
		return err
	}

	// Формируем данные для отчёта, включающие информацию о пользователе, результатах теста и выбранных ответах.
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
		TestType:          us.TestType,
	}
	// Генерируем PDF-отчёт на основе полученных данных.
	reportFile, err := report.GeneratePDFReport(reportData)
	if err != nil {
		log.Printf("Ошибка генерации отчета: %v", err)
	}

	// Удаляем запись о назначении теста для пользователя, так как тест завершён.
	if err := testAssignStore.Delete(user.Username); err != nil {
		log.Printf("Ошибка удаления назначения теста: %v", err)
	}

	// Останавливаем таймер теста.
	helpers.StopTestTimer(bot, user)
	// Отправляем итоговое сообщение пользователю.
	_, _ = bot.Send(user, finalMsg)

	// Если отчёт успешно сгенерирован, создаем документ и отправляем его назначившему (HR/admin).
	if reportFile != "" {
		doc := &telebot.Document{
			File:     telebot.FromDisk(reportFile),
			FileName: reportFile,
		}
		// Отправка отчёта кандидату (опционально, можно раскомментировать при необходимости).
		// _, _ = bot.Send(user, doc)

		// Если AssignedByID не равен 0, отправляем отчёт назначившему HR или администратору.
		if us.AssignedByID != 0 {
			hrUser := &telebot.User{ID: us.AssignedByID}
			if _, err := bot.Send(hrUser, doc); err != nil {
				log.Printf("Ошибка отправки отчета HR (ID=%d): %v", hrUser.ID, err)
			}
		}
	}

	return nil
}
