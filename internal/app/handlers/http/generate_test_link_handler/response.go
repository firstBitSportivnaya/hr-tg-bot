package generate_test_link_handler

// GenerateTestLinkResponse структура для ответа
type GenerateTestLinkResponse struct {
	Link      string `json:"link"`
	QRCodeURL string `json:"qr_code_url"`
}
