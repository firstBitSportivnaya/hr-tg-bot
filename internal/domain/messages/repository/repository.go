package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MessageRepository реализация интерфейса для работы с сообщениями
type MessageRepository struct {
	db *pgxpool.Pool
}

// NewMessageRepository создает новый экземпляр MessageRepository
func NewMessageRepository(db *pgxpool.Pool) *MessageRepository {
	return &MessageRepository{db: db}
}

// GetMessageByKey возвращает текст сообщения по ключу
func (r *MessageRepository) GetMessageByKey(ctx context.Context, messageKey string) (string, error) {
	var messageText string
	err := r.db.QueryRow(ctx, "SELECT message_text FROM messages WHERE message_key=$1", messageKey).
		Scan(&messageText)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("message with key %s not found", messageKey)
		}
		return "", fmt.Errorf("failed to get message: %w", err)
	}
	return messageText, nil
}
