package generate_test_link_handler

import (
	"encoding/json"
	"fmt"
	rolesService "github.com/IT-Nick/internal/domain/roles/service"
	testsService "github.com/IT-Nick/internal/domain/tests/service"
	usersService "github.com/IT-Nick/internal/domain/users/service"
	httpError "github.com/IT-Nick/pkg/http"
	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
	"net/http"
	"os"
)

// GenerateTestLinkHandler структура для обработчика
type GenerateTestLinkHandler struct {
	testService *testsService.TestService
	userService *usersService.UserService
	roleService *rolesService.RoleService
	botUsername string
	baseURL     string
}

// NewGenerateTestLinkHandler создает новый экземпляр обработчика
func NewGenerateTestLinkHandler(
	testService *testsService.TestService,
	userService *usersService.UserService,
	roleService *rolesService.RoleService,
	botUsername, baseURL string,
) *GenerateTestLinkHandler {
	return &GenerateTestLinkHandler{
		testService: testService,
		userService: userService,
		roleService: roleService,
		botUsername: botUsername,
		baseURL:     baseURL,
	}
}

// ServeHTTP метод для обработки запроса
func (h *GenerateTestLinkHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpError.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req GenerateTestLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Проверяем, что username и test_id указаны
	if req.Username == "" || req.TestID <= 0 {
		httpError.ErrorResponse(w, http.StatusBadRequest, "Missing username or test_id")
		return
	}

	// Проверяем, что пользователь существует
	ctx := r.Context()
	user, err := h.userService.GetUserByUsername(ctx, req.Username)
	if err != nil || user == nil {
		httpError.ErrorResponse(w, http.StatusUnauthorized, "User not found")
		return
	}

	// Проверяем, что пользователь имеет разрешение generate_qr
	permissions, err := h.roleService.GetPermissionsForUser(ctx, user.RoleID)
	if err != nil {
		httpError.ErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve permissions")
		return
	}

	hasGenerateQRPermission := false
	for _, perm := range permissions {
		if perm == "generate_qr" {
			hasGenerateQRPermission = true
			break
		}
	}

	if !hasGenerateQRPermission {
		httpError.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized: user does not have permission to generate test links")
		return
	}

	// Проверяем, существует ли тест
	test, err := h.testService.GetTestByID(ctx, req.TestID)
	if err != nil || test == nil {
		httpError.ErrorResponse(w, http.StatusNotFound, fmt.Sprintf("Test with ID %d not found", req.TestID))
		return
	}

	// Генерируем уникальный токен
	token := uuid.New().String()

	// Сохраняем токен в базе
	err = h.testService.SaveTestLink(ctx, req.TestID, token)
	if err != nil {
		httpError.ErrorResponse(w, http.StatusInternalServerError, "Failed to save test link")
		return
	}

	// Формируем ссылку
	link := fmt.Sprintf("https://t.me/%s?start=test_%d_%s_%s", h.botUsername, req.TestID, token, req.Username)

	// Генерируем QR-код
	qrCodeFilename := fmt.Sprintf("test_%d_%s_%s.png", req.TestID, token, req.Username)
	qrCodePath := qrCodeFilename

	err = qrcode.WriteFile(link, qrcode.Medium, 256, qrCodePath)
	if err != nil {
		httpError.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to generate QR code: %v", err))
		return
	}

	// Формируем URL для скачивания QR-кода
	qrCodeURL := fmt.Sprintf("%s/qr/%s", h.baseURL, qrCodeFilename)

	// Отправляем успешный ответ
	response := GenerateTestLinkResponse{
		Link:      link,
		QRCodeURL: qrCodeURL,
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		httpError.ErrorResponse(w, http.StatusInternalServerError, "Failed to encode response")
		return
	}

	// Удаляем временный файл QR-кода
	os.Remove(qrCodePath)
}
