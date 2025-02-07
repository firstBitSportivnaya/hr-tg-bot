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
		// Получаем данные о пользователе, инициировавшем команду.
		user := c.Sender()

		// Попытка извлечь ранее сохраненное состояние пользователя из хранилища.
		existingState, exists := store.Get(user.ID)

		// По умолчанию роль пользователя - "user".
		role := "user"
		if exists {
			// Если состояние существует, берем роль из сохраненного состояния.
			role = existingState.Role
		}

		// Если пользователь содержится в списке администраторов из конфигурации, назначаем роль "admin".
		for _, id := range cfg.AdminIDs {
			if user.ID == id {
				role = "admin"
				break
			}
		}

		// Если для пользователя имеется отложенное назначение роли (например, HR),
		// обновляем роль на указанную в отложенном назначении и удаляем запись о назначении.
		if roleAssign, ok, _ := roleAssignStore.Get(user.Username); ok {
			role = roleAssign.NewRole
			_ = roleAssignStore.Delete(user.Username)
		}

		// Устанавливаем начальное состояние пользователя.
		// По умолчанию состояние - "welcome".
		stateStr := "welcome"

		// Если для пользователя имеется отложенное назначение теста (например, HR назначил тест кандидату),
		// обновляем состояние на "assigned" и сохраняем дополнительные данные о назначении.
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
			// Если нет отложенного назначения теста, сохраняем стандартное состояние.
			if err := store.Set(user.ID, database.UserState{
				Role:              role,
				State:             stateStr,
				TelegramFirstName: user.FirstName,
				TelegramUsername:  user.Username,
			}); err != nil {
				return err
			}
		}

		// Формирование приветственного сообщения с динамическими данными:
		// количество вопросов в тесте и продолжительность теста.
		welcome := fmt.Sprintf(messages.WelcomeFmt, cfg.TestQuestions, int(cfg.TestDuration.Minutes()))

		// Создаем объект разметки для inline-кнопок.
		rm := &telebot.ReplyMarkup{}
		// Создаем кнопку для начала теста.
		startTestBtn := telebot.InlineButton{
			Text:   messages.StartTestButton,
			Unique: "start_test",
			Data:   "start",
		}

		// Инициализируем клавиатуру с первой строкой, содержащей кнопку "Начать тест".
		rows := [][]telebot.InlineButton{
			{startTestBtn},
		}

		// Если роль пользователя "hr", добавляем кнопку для назначения теста кандидату.
		if role == "hr" {
			assignTestBtn := telebot.InlineButton{
				Text:   "Назначить тест кандидату",
				Unique: "assign_test",
				Data:   "assign_test",
			}
			rows = append(rows, []telebot.InlineButton{assignTestBtn})
		} else if role == "admin" {
			// Если роль пользователя "admin", добавляем две кнопки:
			// одну для назначения теста, другую для назначения роли HR.
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
		// Устанавливаем сформированную клавиатуру в объект ReplyMarkup.
		rm.InlineKeyboard = rows

		// Отправляем пользователю приветственное сообщение с inline-кнопками.
		_, err := bot.Send(user, welcome, rm)
		return err
	}
}
