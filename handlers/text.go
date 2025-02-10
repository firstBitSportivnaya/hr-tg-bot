package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/IT-Nick/pending"
	"gopkg.in/telebot.v3"
)

// textHandler обрабатывает входящие текстовые сообщения.
// Он используется для ввода @username кандидата после нажатия inline-кнопок,
// инициирующих назначение теста или роли HR.
func textHandler(bot *telebot.Bot) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		// Получаем информацию об отправителе сообщения.
		user := c.Sender()
		// Извлекаем текущее состояние пользователя из хранилища.
		us, ok := store.Get(user.ID)
		if !ok {
			// Если состояние не найдено, выходим без обработки.
			return nil
		}

		// Обработка сообщения в зависимости от текущего состояния пользователя.
		switch us.State {

		// Состояние "assign_test_waiting" означает, что система ожидает ввода @username кандидата для назначения теста.
		case "assign_test_waiting":
			candidateText := c.Text()
			candidateUsername := strings.TrimSpace(candidateText)
			if !strings.HasPrefix(candidateUsername, "@") {
				candidateUsername = "@" + candidateUsername
			}
			// Для унификации удаляем "@".
			candidateUsername = strings.TrimPrefix(candidateUsername, "@")

			// Проверяем, существует ли уже назначение для этого кандидата.
			if _, exists, err := testAssignStore.Get(candidateUsername); err != nil {
				c.Send("Ошибка при проверке назначения теста.")
			} else if exists {
				c.Send(fmt.Sprintf("Кандидату @%s уже назначен тест.", candidateUsername))
			} else {
				// Используем выбранный тип теста, который сохранился в состоянии HR.
				newAssignment := pending.TestAssignment{
					CandidateID:       0, // Будет заполнен, когда кандидат запустит тест.
					CandidateUsername: candidateUsername,
					AssignedByID:      user.ID,
					AssignedBy:        user.Username,
					AssignedAt:        time.Now(),
					TestType:          us.TestType,
				}
				if err := testAssignStore.Set(candidateUsername, newAssignment); err != nil {
					c.Send("Ошибка при назначении теста.")
				} else {
					c.Send(fmt.Sprintf("Тест успешно назначен кандидату @%s", candidateUsername))
				}
			}

			// После назначения переводим состояние HR в "welcome".
			us.State = "welcome"
			store.Set(user.ID, us)
			return nil

		// Состояние "assign_hr_waiting" означает, что система ожидает ввода @username кандидата для назначения роли HR.
		case "assign_hr_waiting":
			// Получаем введённый текст.
			candidateText := c.Text()
			// Обрезаем пробелы.
			candidateUsername := strings.TrimSpace(candidateText)
			// Добавляем "@" в начале, если его нет.
			if !strings.HasPrefix(candidateUsername, "@") {
				candidateUsername = "@" + candidateUsername
			}
			// Удаляем "@" для унификации.
			candidateUsername = strings.TrimPrefix(candidateUsername, "@")

			// Проверяем, существует ли уже назначение роли для данного кандидата.
			if existing, exists, err := roleAssignStore.Get(candidateUsername); err != nil {
				c.Send("Ошибка при проверке назначения роли.")
			} else if exists {
				// Если роль уже назначена, уведомляем пользователя.
				c.Send(fmt.Sprintf("Кандидату @%s уже назначена роль %s.", candidateUsername, existing.NewRole))
			} else {
				// Формируем новое назначение роли HR.
				newRoleAssign := pending.RoleAssignment{
					CandidateUsername: candidateUsername,
					NewRole:           "hr",
					AssignedBy:        user.Username,
					AssignedAt:        time.Now(),
				}
				// Сохраняем назначение роли в хранилище.
				if err := roleAssignStore.Set(candidateUsername, newRoleAssign); err != nil {
					c.Send("Ошибка при назначении роли HR.")
				} else {
					// Если назначение прошло успешно, отправляем подтверждение.
					c.Send(fmt.Sprintf("Роль HR успешно назначена кандидату @%s", candidateUsername))
				}
			}
			// После обработки возвращаем состояние пользователя в "welcome".
			us.State = "welcome"
			store.Set(user.ID, us)
			return nil

		case "assign_admin_waiting":
			candidateText := c.Text()
			candidateUsername := strings.TrimSpace(candidateText)
			if !strings.HasPrefix(candidateUsername, "@") {
				candidateUsername = "@" + candidateUsername
			}
			// Для унификации удаляем символ "@".
			candidateUsername = strings.TrimPrefix(candidateUsername, "@")

			// Проверяем, есть ли уже назначение для данного кандидата.
			if existing, exists, err := roleAssignStore.Get(candidateUsername); err != nil {
				c.Send("Ошибка при проверке назначения роли.")
			} else if exists {
				c.Send(fmt.Sprintf("Кандидату @%s уже назначена роль %s.", candidateUsername, existing.NewRole))
			} else {
				newRoleAssign := pending.RoleAssignment{
					CandidateUsername: candidateUsername,
					NewRole:           "admin",
					AssignedBy:        user.Username,
					AssignedAt:        time.Now(),
				}
				if err := roleAssignStore.Set(candidateUsername, newRoleAssign); err != nil {
					c.Send("Ошибка при назначении роли администратора.")
				} else {
					c.Send(fmt.Sprintf("Роль администратора успешно назначена кандидату @%s", candidateUsername))
				}
			}
			// Возвращаем состояние в "welcome".
			us.State = "welcome"
			store.Set(user.ID, us)
			return nil

		// Если состояние не соответствует ни одному из ожидаемых вариантов,
		// отправляем сообщение с инструкцией для начала теста.
		default:
			_, err := bot.Send(user, "Для начала теста нажмите кнопку «Начать тест».")
			return err
		}
	}
}
