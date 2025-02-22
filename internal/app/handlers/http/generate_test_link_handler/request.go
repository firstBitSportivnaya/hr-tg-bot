package generate_test_link_handler

// GenerateTestLinkRequest структура для данных запроса
type GenerateTestLinkRequest struct {
	TestID   int    `json:"test_id"`
	Username string `json:"username"`
}
