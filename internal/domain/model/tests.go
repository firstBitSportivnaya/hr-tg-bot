package model

type Test struct {
	ID            int    `json:"id"`
	TestName      string `json:"test_name"`
	TestType      string `json:"test_type"`
	Duration      int    `json:"duration"`
	QuestionCount int    `json:"question_count"`
}
