package handlers

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse represents the JSON error response structure
type ErrorResponse struct {
	Error string `json:"error"`
}

// RespondJSON writes a JSON response with the provided status code and data
func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

// RespondError writes a JSON error response with the provided status code and message
func RespondError(w http.ResponseWriter, status int, message string) {
	RespondJSON(w, status, ErrorResponse{Error: message})
}
