package model

import "time"

type UserTest struct {
	ID                   int        `json:"id"`
	UserID               int        `json:"user_id,omitempty"`
	TestID               int        `json:"test_id"`
	AssignedBy           int        `json:"assigned_by"`
	PendingUsername      *string    `json:"pending_username,omitempty"`
	CurrentQuestionIndex int        `json:"current_question_index,omitempty"`
	CorrectAnswersCount  int        `json:"correct_answers_count,omitempty"`
	MessageID            *int       `json:"message_id,omitempty"`
	TimerDeadline        time.Time  `json:"timer_deadline,omitempty"`
	StartTime            time.Time  `json:"start_time,omitempty"`
	EndTime              *time.Time `json:"end_time,omitempty"`
	Status               *string    `json:"status,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}
