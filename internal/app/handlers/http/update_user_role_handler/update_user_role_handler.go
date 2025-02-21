package update_user_role_handler

import (
	"encoding/json"
	"fmt"
	roleService "github.com/IT-Nick/internal/domain/roles/service"
	"github.com/IT-Nick/internal/domain/users/service"
	httpError "github.com/IT-Nick/pkg/http"
	"net/http"
)

// UpdateUserRoleRequest структура для данных запроса
type UpdateUserRoleRequest struct {
	Username string `json:"username"`
	RoleName string `json:"role_name"`
}

// UpdateUserRoleHandler структура для обработчика
type UpdateUserRoleHandler struct {
	userService *service.UserService
	roleService *roleService.RoleService
}

// NewUpdateUserRoleHandler создает новый экземпляр обработчика
func NewUpdateUserRoleHandler(userService *service.UserService, roleService *roleService.RoleService) *UpdateUserRoleHandler {
	return &UpdateUserRoleHandler{
		userService: userService,
		roleService: roleService,
	}
}

// ServeHTTP метод для обработки запроса
func (h *UpdateUserRoleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Декодируем запрос
	var request UpdateUserRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpError.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Получаем роль по имени
	role, err := h.roleService.GetRoleByRoleName(r.Context(), request.RoleName)
	if err != nil {
		httpError.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to find role: %v", err))
		return
	}

	// Обновляем роль пользователя
	userID, err := h.userService.UpdateUserRole(r.Context(), request.Username, role.Name)
	if err != nil {
		httpError.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update user role: %v", err))
		return
	}

	// Отправляем успешный ответ
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	response := map[string]interface{}{
		"message": fmt.Sprintf("User %s role updated to %s", request.Username, role.Name),
		"user_id": userID,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		httpError.ErrorResponse(w, http.StatusInternalServerError, "Failed to encode response")
		return
	}
}
