package model

import "time"

// Question представляет вопрос теста
type Question struct {
	ID            int       `json:"id"`
	TestID        int       `json:"test_id"`
	QuestionText  string    `json:"question_text"`
	AnswerType    string    `json:"answer_type"` // "single", "multiple", "text"
	CorrectAnswer string    `json:"correct_answer"`
	TestOptions   []string  `json:"test_options"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
