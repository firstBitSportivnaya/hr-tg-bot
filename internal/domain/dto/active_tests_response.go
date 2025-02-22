package dto

// ActiveTestsResponse структура для отчета по активным тестам
type ActiveTestsResponse struct {
	TotalActiveUsers int              `json:"total_active_users"`
	ActiveTests      []ActiveTestInfo `json:"active_tests"`
}

type ActiveTestInfo struct {
	TelegramUsername string             `json:"telegram_username"`
	FullName         string             `json:"full_name"`
	TestID           int                `json:"test_id"`
	TestName         string             `json:"test_name"`
	TestType         string             `json:"test_type"`
	Duration         int                `json:"duration"`
	CurrentQuestion  QuestionInfoActive `json:"current_question"`
	PreviousAnswers  []AnswerInfo       `json:"previous_answers"`
	CorrectAnswers   int                `json:"correct_answers"`
	TotalQuestions   int                `json:"total_questions"`
	RemainingTime    string             `json:"remaining_time"`
	Status           *string            `json:"status,omitempty"`
}

type QuestionInfoActive struct {
	QuestionID   int      `json:"question_id"`
	QuestionText string   `json:"question_text"`
	AnswerType   string   `json:"answer_type"`
	TestOptions  []string `json:"test_options"`
}

type AnswerInfo struct {
	QuestionID   int    `json:"question_id"`
	QuestionText string `json:"question_text"`
	UserAnswer   string `json:"user_answer"`
	IsCorrect    bool   `json:"is_correct"`
	AnsweredAt   string `json:"answered_at"`
}
