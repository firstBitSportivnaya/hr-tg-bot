package service

import (
	"context"
	"fmt"
	"github.com/IT-Nick/internal/domain/messages/repository"
	"github.com/IT-Nick/internal/domain/model"
)

// MessageService содержит логику для работы с сообщениями
type MessageService struct {
	messageRepo *repository.MessageRepository
}

// NewMessageService создает новый экземпляр MessageService
func NewMessageService(messageRepo *repository.MessageRepository) *MessageService {
	return &MessageService{messageRepo: messageRepo}
}

// GetMessageByKey возвращает сообщение по ключу из базы данных
func (s *MessageService) GetMessageByKey(ctx context.Context, messageKey string) (string, error) {
	message, err := s.messageRepo.GetMessageByKey(ctx, messageKey)
	if err != nil {
		return "", fmt.Errorf("failed to get message by key: %w", err)
	}
	return message, nil
}

// GetButtons возвращает мапу с константными кнопками, где значения подгружаются из базы
func (s *MessageService) GetButtons(ctx context.Context) (map[string]string, error) {
	buttons := make(map[string]string)

	// Получаем текст для каждой кнопки из базы данных
	for _, key := range []string{model.StartTestKey, model.AssignHRKey, model.AssignAdminKey, model.AssignTestKey} {
		text, err := s.messageRepo.GetMessageByKey(ctx, key)
		if err != nil {
			return nil, fmt.Errorf("failed to get button text for key %s: %w", key, err)
		}
		buttons[key] = text
	}

	return buttons, nil
}
