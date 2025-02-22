package model

import "time"

// Answer представляет ответ пользователя на вопрос теста
type Answer struct {
	ID         int       `json:"id"`
	UserTestID int       `json:"user_test_id"`
	QuestionID int       `json:"question_id"`
	UserAnswer string    `json:"user_answer"`
	IsCorrect  bool      `json:"is_correct"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
