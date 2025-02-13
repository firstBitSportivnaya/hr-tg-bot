package handlers

import (
	"fmt"
	"github.com/IT-Nick/database/json"
	"gopkg.in/telebot.v3"
	"strings"
)

// assignHandler обрабатывает назначение теста кандидату.
// При нажатии inline-кнопки пользователем с ролью HR или admin, его состояние переводится в режим "assign_test_waiting",
// после чего ему отправляется сообщение с просьбой ввести @username кандидата для назначения теста.
func assignHandler() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		sender := c.Sender()
		senderState, ok := store.Get(sender.ID)
		if !ok || (senderState.Role != "hr" && senderState.Role != "admin") {
			return c.Send("У вас нет прав для назначения теста.")
		}

		// Если это callback, то загружаем типы тестов и выводим их в виде inline-кнопок.
		if c.Callback() != nil {
			testTypes, err := json.LoadTestTypes("data/test_types.json")
			if err != nil {
				return c.Send("Ошибка загрузки типов тестов.")
			}
			rm := &telebot.ReplyMarkup{}
			var buttons []telebot.InlineButton
			for _, tt := range testTypes {
				data := fmt.Sprintf("select_type_%s", tt.Type)
				btn := telebot.InlineButton{
					Text:   tt.Description,
					Unique: "select_test_type",
					Data:   data,
				}
				buttons = append(buttons, btn)
			}
			// Например, выводим кнопки в одну строку
			rm.InlineKeyboard = [][]telebot.InlineButton{buttons}

			// Обновляем состояние отправителя (например, "assign_test_select_type")
			senderState.State = "assign_test_select_type"
			if err := store.Set(sender.ID, senderState); err != nil {
				return err
			}
			return c.Send("Выберите тип теста:", rm)
		}
		return nil
	}
}

// selectTestTypeHandler обрабатывает выбор типа теста HR/admin'ом.
// После нажатия на кнопку с типом теста:
//   - Извлекается выбранный тип,
//   - Обновляется pending-запись назначения теста,
//   - Изменяется состояние отправителя (теперь ожидается ввод @username кандидата),
//   - Исходное сообщение с кнопками удаляется из чата,
//   - Отправляется уведомление о выбранном типе.
func selectTestTypeHandler(bot *telebot.Bot) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		data := c.Callback().Data
		parts := strings.Split(data, "_")
		if len(parts) < 3 {
			return c.Send("Неверные данные выбора типа теста.")
		}
		selectedType := parts[2]

		// Не создаём pending‑запись сразу,
		// а сохраняем выбранный тип теста в состоянии отправителя.
		senderState, ok := store.Get(c.Sender().ID)
		if !ok {
			return c.Send("Ошибка состояния пользователя.")
		}
		senderState.TestType = selectedType
		senderState.State = "assign_test_waiting"
		if err := store.Set(c.Sender().ID, senderState); err != nil {
			return err
		}

		// Удаляем сообщение с кнопками, чтобы оно не оставалось в чате.
		if err := c.Delete(); err != nil {
			fmt.Printf("Ошибка удаления сообщения: %v\n", err)
		}

		notifyText := fmt.Sprintf("Выбран тип теста: '%s'. Теперь введите @username кандидата для назначения теста.", selectedType)
		bot.Send(c.Sender(), notifyText)
		return c.Respond()
	}
}

// assignHRHandler обрабатывает назначение роли HR кандидату.
// При нажатии inline-кнопки пользователем с ролью admin его состояние переводится в режим "assign_hr_waiting",
// после чего ему отправляется сообщение с просьбой ввести @username кандидата для назначения HR.
func assignHRHandler() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		// Получаем информацию об отправителе.
		sender := c.Sender()
		// Извлекаем состояние пользователя из глобального хранилища.
		senderState, ok := store.Get(sender.ID)
		// Если состояние не найдено или роль пользователя не admin,
		// отправляем сообщение о недостатке прав для назначения HR.
		if !ok || senderState.Role != "admin" {
			return c.Send("У вас нет прав для назначения HR.")
		}
		// Если обновление является callback-запросом,
		// переводим состояние пользователя в режим ожидания ввода @username кандидата для назначения HR.
		if c.Callback() != nil {
			senderState.State = "assign_hr_waiting"
			// Сохраняем обновленное состояние в хранилище.
			if err := store.Set(sender.ID, senderState); err != nil {
				return err
			}
			// Отправляем сообщение с инструкцией по вводу @username кандидата.
			return c.Send("Пожалуйста, введите @username кандидата для назначения HR.")
		}
		return nil
	}
}

// assignAdminHandler обрабатывает назначение роли администратора.
// Если администратор нажимает inline‑кнопку, его состояние переводится в режим "assign_admin_waiting"
// и ему отправляется сообщение с просьбой ввести @username кандидата для назначения администратора.
func assignAdminHandler() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		sender := c.Sender()
		senderState, ok := store.Get(sender.ID)
		if !ok || senderState.Role != "admin" {
			return c.Send("У вас нет прав для назначения администратора.")
		}
		// Если это callback‑запрос, переводим состояние и просим ввести @username кандидата.
		if c.Callback() != nil {
			senderState.State = "assign_admin_waiting"
			if err := store.Set(sender.ID, senderState); err != nil {
				return err
			}
			return c.Send("Пожалуйста, введите @username кандидата для назначения администратора.")
		}
		return nil
	}
}
