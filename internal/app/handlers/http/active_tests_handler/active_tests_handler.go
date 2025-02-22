package active_tests_handler

import (
	"encoding/json"
	"fmt"
	"github.com/IT-Nick/internal/domain/dto"
	testsService "github.com/IT-Nick/internal/domain/tests/service"
	"github.com/IT-Nick/internal/domain/users/service"
	httpError "github.com/IT-Nick/pkg/http"
	"net/http"
)

// ActiveTestsHandler структура для обработчика
type ActiveTestsHandler struct {
	userService *service.UserService
	testService *testsService.TestService
}

// NewActiveTestsHandler создает новый экземпляр обработчика
func NewActiveTestsHandler(userService *service.UserService, testService *testsService.TestService) *ActiveTestsHandler {
	return &ActiveTestsHandler{
		userService: userService,
		testService: testService,
	}
}

// ServeHTTP метод для обработки запроса
func (h *ActiveTestsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Получаем список активных тестов
	activeTests, err := h.testService.GetActiveTests(ctx)
	if err != nil {
		httpError.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get active tests: %v", err))
		return
	}

	// Формируем отчет
	response := dto.ActiveTestsResponse{
		TotalActiveUsers: len(activeTests),
		ActiveTests:      activeTests,
	}

	// Отправляем успешный ответ
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		httpError.ErrorResponse(w, http.StatusInternalServerError, "Failed to encode response")
		return
	}
}
