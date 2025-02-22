package user_test_report_handler

import (
	"encoding/json"
	"fmt"
	"github.com/IT-Nick/internal/domain/dto"
	testsService "github.com/IT-Nick/internal/domain/tests/service"
	"github.com/IT-Nick/internal/domain/users/service"
	httpError "github.com/IT-Nick/pkg/http"
	"net/http"
)

// UserTestReportRequest структура для данных запроса
type UserTestReportRequest struct {
	Username string `json:"username"`
}

// UserTestReportHandler структура для обработчика
type UserTestReportHandler struct {
	userService *service.UserService
	testService *testsService.TestService
}

// NewUserTestReportHandler создает новый экземпляр обработчика
func NewUserTestReportHandler(userService *service.UserService, testService *testsService.TestService) *UserTestReportHandler {
	return &UserTestReportHandler{
		userService: userService,
		testService: testService,
	}
}

// ServeHTTP метод для обработки запроса
func (h *UserTestReportHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Декодируем тело запроса
	var request UserTestReportRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpError.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Проверяем, что username указан
	if request.Username == "" {
		httpError.ErrorResponse(w, http.StatusBadRequest, "Missing username in request body")
		return
	}

	// Получаем пользователя по username (telegram_username)
	ctx := r.Context()
	user, err := h.userService.GetUserByUsername(ctx, request.Username)
	if err != nil {
		httpError.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to find user: %v", err))
		return
	}
	if user == nil {
		httpError.ErrorResponse(w, http.StatusNotFound, fmt.Sprintf("User %s not found", request.Username))
		return
	}

	// Получаем отчет по тестам пользователя
	report, err := h.testService.GetUserTestReport(ctx, user.ID)
	if err != nil {
		httpError.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to generate report: %v", err))
		return
	}

	// Формируем полный отчет
	fullName := fmt.Sprintf("%s %s %s", user.RealFirstName, user.RealSecondName, user.RealSurname)
	response := dto.UserTestReportResponse{
		Username:    request.Username,
		TelegramID:  *user.TelegramID,
		FullName:    fullName,
		TestHistory: report,
	}

	// Отправляем успешный ответ
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		httpError.ErrorResponse(w, http.StatusInternalServerError, "Failed to encode response")
		return
	}
}
