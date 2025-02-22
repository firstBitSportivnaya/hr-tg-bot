package timer

import (
	"context"
	"fmt"
	testsService "github.com/IT-Nick/internal/domain/tests/service"
	usersService "github.com/IT-Nick/internal/domain/users/service"
	"gopkg.in/telebot.v4"
	"log"
	"time"
)

type Updater struct {
	bot         *telebot.Bot
	testService *testsService.TestService
	userService *usersService.UserService
}

func NewTimerUpdater(bot *telebot.Bot, testService *testsService.TestService, userService *usersService.UserService) *Updater {
	return &Updater{
		bot:         bot,
		testService: testService,
		userService: userService,
	}
}

// UpdateTimer обновляет сообщение с таймером, номером вопроса и статусом теста
func (tu *Updater) UpdateTimer(ctx context.Context, userID int64, messageID int, deadline time.Time, userTestID int, totalQuestions int) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Контекст отменен, завершаем обновление таймера
			log.Printf("Timer update canceled for user %d", userID)
			return
		case <-ticker.C:
			// Вычисляем оставшееся время
			timeLeft := time.Until(deadline)
			log.Printf("Deadline: %s, Time left: %s", deadline, timeLeft)
			if timeLeft <= 0 {
				// Время вышло, проверяем статус теста
				_, _, status, err := tu.testService.GetUserTestState(ctx, userTestID)
				if err != nil {
					log.Printf("Failed to get user test state for user %d: %v", userID, err)
					return
				}

				if status != "finished" {
					// Обновляем статус теста на "finished" и сохраняем end_time
					endTime := time.Now()
					err = tu.testService.UpdateUserTestStatus(ctx, userTestID, "finished")
					err = tu.testService.UpdateUserTestEndTime(ctx, userTestID, endTime)
					if err != nil {
						log.Printf("Failed to update test status for user %d: %v", userID, err)
					}

					// Получаем пользователя, который назначил тест
					userTest, err := tu.userService.GetUserTestByID(ctx, userTestID)
					if err != nil {
						log.Printf("Failed to get user test for user (ошибка в таймере) %d: %v", userID, err)
					}

					assignedByTgId, err := tu.userService.GetUserByID(ctx, userTest.AssignedBy)
					if err != nil {
						log.Printf("Ошибка получения HR менеджера в таймере: %v", err)
					}

					user, err := tu.userService.GetUserByTelegramID(ctx, userID)
					if err != nil {
						log.Printf("Ошибка получения пользователя по айди телеграм в таймере: %v", err)
					}

					test, err := tu.testService.GetLastTestForUserWithFinishStatus(ctx, user.TelegramUsername)
					if err != nil {
						log.Printf("GetLastTestForUserWithFinishStatus в таймере: %v", err)
					}
					// Отправляем сообщение о завершении теста пользователю assigned_by
					_, err = tu.bot.Send(&telebot.User{ID: *assignedByTgId.TelegramID}, fmt.Sprintf("⚡️ Кандидат *%s* завершил выполнение теста *%s*.", user.TelegramUsername, test.TestName), &telebot.SendOptions{
						ParseMode: telebot.ModeMarkdown,
					})

					// Отправляем сообщение о завершении времени
					_, err = tu.bot.Edit(&telebot.Message{
						ID:   messageID,
						Chat: &telebot.Chat{ID: userID},
					}, "⏰ Время вышло!", &telebot.SendOptions{
						ParseMode: telebot.ModeMarkdown,
					})
					if err != nil {
						log.Printf("Failed to update timer message for user %d: %w", userID, err)
					}
				}
				return
			}

			// Получаем текущее состояние теста (индекс текущего вопроса)
			currentQuestionIndex, _, status, err := tu.testService.GetUserTestState(ctx, userTestID)
			if err != nil {
				log.Printf("Failed to get user test state for user %d: %v", userID, err)
				continue
			}

			// Если тест уже завершен, прекращаем обновление таймера
			if status == "finished" {
				log.Printf("Test already finished for user %d", userID)
				return
			}

			// Вычисляем минуты и секунды
			minutes := int(timeLeft.Minutes())
			seconds := int(timeLeft.Seconds()) % 60

			// Формируем текст сообщения с таймером и номером вопроса
			timerText := fmt.Sprintf(
				"⏰ Тест начался! Оставшееся время: %02d:%02d, Вопрос %d/%d",
				minutes, seconds, currentQuestionIndex+1, totalQuestions,
			)

			// Обновляем сообщение с таймером
			_, err = tu.bot.Edit(&telebot.Message{
				ID:   messageID,
				Chat: &telebot.Chat{ID: userID},
			}, timerText, &telebot.SendOptions{
				ParseMode: telebot.ModeMarkdown,
			})
			if err != nil {
				log.Printf("Failed to update timer message for user %d: %w", userID, err)
			}
		}
	}
}
