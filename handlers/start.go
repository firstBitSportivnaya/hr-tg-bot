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

	"github.com/IT-Nick/database"
	"github.com/IT-Nick/messages"
	"gopkg.in/telebot.v3"
)

// startHandler обрабатывает команду /start для различных типов пользователей: user, HR и admin.
// Функция определяет роль пользователя, устанавливает начальное состояние и формирует приветственное сообщение
// с соответствующими inline-кнопками для дальнейших действий.
func startHandler(bot *telebot.Bot) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		user := c.Sender()
		existingState, exists := store.Get(user.ID)
		role := "user"
		if exists {
			role = existingState.Role
		}
		for _, id := range cfg.AdminIDs {
			if user.ID == id {
				role = "admin"
				break
			}
		}
		if roleAssign, ok, _ := roleAssignStore.Get(user.Username); ok {
			role = roleAssign.NewRole
			_ = roleAssignStore.Delete(user.Username)
		}

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
			assignAdminBtn := telebot.InlineButton{
				Text:   "Назначить администратора",
				Unique: "assign_admin",
				Data:   "assign_admin",
			}
			rows = append(rows, []telebot.InlineButton{assignTestBtn, assignHRBtn, assignAdminBtn})
		}
		rm.InlineKeyboard = rows
		_, err := bot.Send(user, welcome, rm)
		return err
	}
}
