package http

import (
	"encoding/json"
	"net/http"
)

type errorResponse struct {
	Message string
}

func ErrorResponse(w http.ResponseWriter, statusCode int, message string) error {
	w.WriteHeader(statusCode)
	w.Header().Add("Content-Type", "application/json")

	return json.NewEncoder(w).Encode(&errorResponse{Message: message})
}
