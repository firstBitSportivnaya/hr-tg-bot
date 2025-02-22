package model

import "time"

type User struct {
	ID                int       `json:"id"`
	RoleID            int       `json:"role_id"`
	TelegramID        *int64    `json:"telegram_id,omitempty"`
	TelegramUsername  string    `json:"telegram_username"`
	TelegramFirstName *string   `json:"telegram_first_name,omitempty"`
	RealFirstName     *string   `json:"real_first_name,omitempty"`
	RealSecondName    *string   `json:"real_second_name,omitempty"`
	RealSurname       *string   `json:"real_surname,omitempty"`
	CurrentState      *string   `json:"current_state,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
