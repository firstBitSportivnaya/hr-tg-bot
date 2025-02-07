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
	"gopkg.in/telebot.v3"
)

// assignHandler обрабатывает назначение теста кандидату.
// При нажатии inline-кнопки пользователем с ролью HR или admin, его состояние переводится в режим "assign_test_waiting",
// после чего ему отправляется сообщение с просьбой ввести @username кандидата для назначения теста.
func assignHandler() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		// Получаем информацию об отправителе (пользователе, инициировавшем действие).
		sender := c.Sender()
		// Извлекаем текущее состояние пользователя из глобального хранилища.
		senderState, ok := store.Get(sender.ID)
		// Если состояние не найдено или роль пользователя не позволяет назначать тест (допустимы только "hr" и "admin"),
		// отправляем уведомление об отсутствии прав.
		if !ok || (senderState.Role != "hr" && senderState.Role != "admin") {
			return c.Send("У вас нет прав для назначения теста.")
		}
		// Если обновление является callback-запросом (например, при нажатии inline-кнопки),
		// переводим состояние пользователя в режим ожидания ввода @username кандидата.
		if c.Callback() != nil {
			senderState.State = "assign_test_waiting"
			// Сохраняем обновленное состояние в хранилище.
			if err := store.Set(sender.ID, senderState); err != nil {
				return err
			}
			// Отправляем сообщение с инструкцией по вводу @username кандидата.
			return c.Send("Пожалуйста, введите @username кандидата для назначения теста.")
		}
		return nil
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
