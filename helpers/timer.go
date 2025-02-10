package helpers

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/IT-Nick/config"
	"github.com/IT-Nick/database"
	"gopkg.in/telebot.v3"
)

// TimerMessageManager хранит функции отмены (cancel functions) для таймеров,
// ассоциированных с идентификаторами пользователей (user.ID). Это позволяет
// управлять активными таймерами (например, отменять старые таймеры при запуске нового).
var TimerMessageManager = make(map[int64]context.CancelFunc)

// StartTimerMessage запускает или перезапускает таймер для тестирования пользователя.
// Функция отправляет сообщение с информацией о времени, обновляет его каждую секунду и
// вызывает функцию onTimeout, если время истекло.
//
// Параметры:
// - bot: указатель на экземпляр телеграм-бота.
// - user: указатель на пользователя, для которого запускается таймер.
// - cfg: конфигурация приложения, содержащая длительность теста.
// - getTimerText: функция, которая возвращает текст для сообщения таймера.
// - onTimeout: функция, вызываемая при истечении времени теста.
func StartTimerMessage(
	bot *telebot.Bot,
	user *telebot.User,
	cfg *config.Config,
	getTimerText func() string,
	onTimeout func(),
) {
	if cancel, ok := TimerMessageManager[user.ID]; ok {
		cancel()
	}

	us, ok := database.GlobalStore.Get(user.ID)
	if !ok {
		return
	}

	testDuration := cfg.TestDuration
	if us.TestDuration > 0 {
		testDuration = us.TestDuration
	}

	if us.TimerDeadline.IsZero() || time.Now().After(us.TimerDeadline) {
		us.TimerDeadline = time.Now().Add(testDuration)
		_ = database.GlobalStore.Set(user.ID, us)
	}

	initialText := getTimerText()
	timerMsg, err := bot.Send(user, initialText)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка отправки сообщения таймера для userID=%d: %v\n", user.ID, err)
		return
	}

	us.TimerMessageID = timerMsg.ID
	_ = database.GlobalStore.Set(user.ID, us)

	ctx, cancel := context.WithCancel(context.Background())
	TimerMessageManager[user.ID] = cancel

	go func() {
		defer delete(TimerMessageManager, user.ID)
		msg := timerMsg
		for {
			select {
			case <-ctx.Done():
				return
			default:
				newText := getTimerText()
				_, err := bot.Edit(msg, newText)
				if err != nil && !containsNotModifiedError(err.Error()) {
					fmt.Fprintf(os.Stderr, "Ошибка редактирования сообщения таймера для userID=%d: %v\n", user.ID, err)
				}
				us, ok := database.GlobalStore.Get(user.ID)
				if !ok {
					return
				}
				remaining := us.TimerDeadline.Sub(time.Now())
				if remaining <= 0 {
					onTimeout()
					return
				}
				time.Sleep(time.Second)
			}
		}
	}()
}

// StopTestTimer останавливает активный таймер теста для пользователя.
// Функция отменяет обновление сообщения таймера, удаляет само сообщение и сбрасывает параметры таймера
// в состоянии пользователя.
func StopTestTimer(bot *telebot.Bot, user *telebot.User) {
	// Если существует активная функция отмены для таймера, вызываем её и удаляем запись.
	if cancel, ok := TimerMessageManager[user.ID]; ok {
		cancel()
		delete(TimerMessageManager, user.ID)
	}
	// Получаем текущее состояние пользователя.
	us, ok := database.GlobalStore.Get(user.ID)
	if !ok {
		return
	}
	// Если существует сообщение таймера, удаляем его.
	if us.TimerMessageID != 0 {
		_ = bot.Delete(&telebot.Message{
			ID:   us.TimerMessageID,
			Chat: &telebot.Chat{ID: user.ID},
		})
	}
	// Сбрасываем поля, связанные с таймером, и обновляем состояние пользователя.
	us.TimerMessageID = 0
	us.TimerDeadline = time.Time{}
	_ = database.GlobalStore.Set(user.ID, us)
}

// RemainingTimeStr возвращает строковое представление оставшегося времени до дедлайна.
// Если время истекло, возвращается "0s".
func RemainingTimeStr(deadline time.Time) string {
	remaining := deadline.Sub(time.Now())
	if remaining < 0 {
		remaining = 0
	}
	return remaining.Round(time.Second).String()
}

// RestoreActiveTimers восстанавливает активные таймеры для пользователей, находящихся в процессе тестирования,
// при перезапуске приложения. Функция проходит по всем сохраненным состояниям и для каждого пользователя,
// у которого состояние "testing" и установлен дедлайн, запускает таймер.
func RestoreActiveTimers(
	bot *telebot.Bot,
	cfg *config.Config,
	store database.Store,
	onTimeout func(bot *telebot.Bot, user *telebot.User) error,
) {
	// Пытаемся привести хранилище к JSONStore, т.к. восстановление таймеров работает только для него.
	jsonStore, ok := store.(*database.JSONStore)
	if !ok {
		return
	}
	// Загружаем все состояния пользователей из JSON-файла.
	allStates, err := jsonStore.LoadAllStates()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка восстановления состояний: %v\n", err)
		return
	}
	// Для каждого пользователя, у которого тест в процессе и дедлайн не истек, запускаем таймер.
	for userID, us := range allStates {
		if us.State == "testing" && !us.TimerDeadline.IsZero() && time.Now().Before(us.TimerDeadline) {
			user := &telebot.User{ID: userID}
			StartTimerMessage(bot, user, cfg, func() string {
				s, ok := store.Get(userID)
				if !ok {
					return ""
				}
				// Форматируем строку с оставшимся временем и текущим номером вопроса.
				return fmt.Sprintf("Время: %s\nВопрос %d из %d",
					RemainingTimeStr(s.TimerDeadline),
					s.CurrentQuestion+1,
					len(s.TestTasks))
			}, func() {
				// Вызываем onTimeout для пользователя и, при ошибке, выводим сообщение.
				if err := onTimeout(bot, user); err != nil {
					fmt.Fprintf(os.Stderr, "Ошибка завершения теста для userID=%d: %v\n", user.ID, err)
				}
			})
		}
	}
}

// containsNotModifiedError проверяет, содержит ли строка ошибки подстроку "message is not modified".
// Это используется для определения, является ли ошибка редактирования сообщения несущественной.
func containsNotModifiedError(errStr string) bool {
	return (errStr != "" && contains(errStr, "message is not modified"))
}

// contains проверяет, содержится ли подстрока substr в строке s.
// Реализовано простым перебором символов.
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return false
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
