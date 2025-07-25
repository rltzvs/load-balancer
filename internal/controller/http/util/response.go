package util

import (
	"encoding/json"
	"net/http"
)

func RespondJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if payload != nil {
		json.NewEncoder(w).Encode(payload)
	}
}

func RespondError(w http.ResponseWriter, statusCode int, message string) {
	type errorResponse struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	RespondJSON(w, statusCode, errorResponse{Code: statusCode, Message: message})
}
