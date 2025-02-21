package model

// Role представляет роль пользователя
type Role struct {
	ID   int    `json:"id"`
	Name string `json:"role_name"`
}

// Permission представляет право, которое может быть привязано к роли
type Permission struct {
	ID   int    `json:"id"`
	Name string `json:"permission_name"`
}
