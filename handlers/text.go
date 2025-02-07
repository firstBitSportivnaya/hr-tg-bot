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
			// Получаем текст сообщения, введённый пользователем.
			candidateText := c.Text()
			// Обрезаем пробельные символы по краям.
			candidateUsername := strings.TrimSpace(candidateText)
			// Если введённый текст не начинается с "@", добавляем его.
			if !strings.HasPrefix(candidateUsername, "@") {
				candidateUsername = "@" + candidateUsername
			}
			// Удаляем "@" для унификации хранения имени кандидата.
			candidateUsername = strings.TrimPrefix(candidateUsername, "@")

			// Проверяем, существует ли уже назначение теста для данного кандидата.
			if _, exists, err := testAssignStore.Get(candidateUsername); err != nil {
				// В случае ошибки отправляем сообщение об ошибке проверки.
				c.Send("Ошибка при проверке назначения теста.")
			} else if exists {
				// Если тест уже назначен, уведомляем пользователя.
				c.Send(fmt.Sprintf("Кандидату @%s уже назначен тест.", candidateUsername))
			} else {
				// Если назначение отсутствует, формируем новую запись для назначения теста.
				newAssignment := pending.TestAssignment{
					CandidateID:       0, // ID кандидата будет назначен позже при запуске теста.
					CandidateUsername: candidateUsername,
					AssignedByID:      user.ID,
					AssignedBy:        user.Username,
					AssignedAt:        time.Now(),
				}
				// Сохраняем новое назначение в хранилище.
				if err := testAssignStore.Set(candidateUsername, newAssignment); err != nil {
					c.Send("Ошибка при назначении теста.")
				} else {
					// Если всё прошло успешно, отправляем подтверждение.
					c.Send(fmt.Sprintf("Тест успешно назначен кандидату @%s", candidateUsername))
				}
			}
			// Возвращаем состояние пользователя в "welcome" после обработки.
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

		// Если состояние не соответствует ни одному из ожидаемых вариантов,
		// отправляем сообщение с инструкцией для начала теста.
		default:
			_, err := bot.Send(user, "Для начала теста нажмите кнопку «Начать тест».")
			return err
		}
	}
}
