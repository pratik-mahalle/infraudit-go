package utils

import (
	"encoding/json"
	"net/http"

	"infraaudit/backend/internal/pkg/errors"
)

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Success bool        `json:"success"`
	Error   ErrorDetail `json:"error"`
}

// ErrorDetail contains error details
type ErrorDetail struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// WriteJSON writes a JSON response
func WriteJSON(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

// WriteSuccess writes a successful JSON response
func WriteSuccess(w http.ResponseWriter, status int, data interface{}) error {
	return WriteJSON(w, status, SuccessResponse{
		Success: true,
		Data:    data,
	})
}

// WriteSuccessWithMessage writes a successful JSON response with a message
func WriteSuccessWithMessage(w http.ResponseWriter, status int, message string, data interface{}) error {
	return WriteJSON(w, status, SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// WriteError writes an error JSON response from AppError
func WriteError(w http.ResponseWriter, err *errors.AppError) error {
	return WriteJSON(w, err.StatusCode, ErrorResponse{
		Success: false,
		Error: ErrorDetail{
			Code:    err.Code,
			Message: err.Message,
			Details: err.Details,
		},
	})
}

// WriteErrorMessage writes a simple error message
func WriteErrorMessage(w http.ResponseWriter, status int, code, message string) error {
	return WriteJSON(w, status, ErrorResponse{
		Success: false,
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}
