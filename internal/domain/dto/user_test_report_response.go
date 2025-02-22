package dto

// UserTestReportResponse структура для отчета по тестам пользователя
type UserTestReportResponse struct {
	Username    string        `json:"username"`
	TelegramID  int64         `json:"telegram_id"`
	FullName    string        `json:"full_name"`
	TestHistory []TestHistory `json:"test_history"`
}

type TestHistory struct {
	UserTestID     int            `json:"user_test_id"`
	TestID         int            `json:"test_id"`
	TestName       string         `json:"test_name"`
	TestType       string         `json:"test_type"`
	Duration       int            `json:"duration"`
	QuestionCount  int            `json:"question_count"`
	Status         string         `json:"status"`
	StartTime      string         `json:"start_time"`
	EndTime        string         `json:"end_time"`
	CorrectAnswers int            `json:"correct_answers"`
	TotalQuestions int            `json:"total_questions"`
	TimerDeadline  string         `json:"timer_deadline"`
	AssignedBy     string         `json:"assigned_by"`
	Questions      []QuestionInfo `json:"questions"`
}

type QuestionInfo struct {
	QuestionID    int      `json:"question_id"`
	QuestionText  string   `json:"question_text"`
	AnswerType    string   `json:"answer_type"`
	CorrectAnswer string   `json:"correct_answer,omitempty"`
	TestOptions   []string `json:"test_options,omitempty"`
	UserAnswer    string   `json:"user_answer"`
	IsCorrect     bool     `json:"is_correct"`
	AnsweredAt    string   `json:"answered_at"`
}
