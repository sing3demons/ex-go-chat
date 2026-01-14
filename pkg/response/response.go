package response

import (
	"encoding/json"
	"net/http"

	"realtime-chat-system/pkg/errors"
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
	Error   string      `json:"error"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// JSON sends a JSON response
func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// Success sends a successful JSON response
func Success(w http.ResponseWriter, data interface{}, message string) {
	JSON(w, http.StatusOK, SuccessResponse{
		Success: true,
		Data:    data,
		Message: message,
	})
}

// Created sends a 201 Created response
func Created(w http.ResponseWriter, data interface{}, message string) {
	JSON(w, http.StatusCreated, SuccessResponse{
		Success: true,
		Data:    data,
		Message: message,
	})
}

// Error sends an error JSON response
func Error(w http.ResponseWriter, err error) {
	// Check if it's an AppError
	if appErr, ok := err.(*errors.AppError); ok {
		JSON(w, appErr.StatusCode, ErrorResponse{
			Success: false,
			Error:   "error",
			Code:    appErr.Code,
			Message: appErr.Message,
			Details: appErr.Details,
		})
		return
	}

	// Default to internal server error
	JSON(w, http.StatusInternalServerError, ErrorResponse{
		Success: false,
		Error:   "error",
		Code:    "INTERNAL_ERROR",
		Message: err.Error(),
	})
}

// BadRequest sends a 400 Bad Request response
func BadRequest(w http.ResponseWriter, message string) {
	JSON(w, http.StatusBadRequest, ErrorResponse{
		Success: false,
		Error:   "error",
		Code:    "BAD_REQUEST",
		Message: message,
	})
}

// Unauthorized sends a 401 Unauthorized response
func Unauthorized(w http.ResponseWriter, message string) {
	JSON(w, http.StatusUnauthorized, ErrorResponse{
		Success: false,
		Error:   "error",
		Code:    "UNAUTHORIZED",
		Message: message,
	})
}

// Forbidden sends a 403 Forbidden response
func Forbidden(w http.ResponseWriter, message string) {
	JSON(w, http.StatusForbidden, ErrorResponse{
		Success: false,
		Error:   "error",
		Code:    "FORBIDDEN",
		Message: message,
	})
}

// NotFound sends a 404 Not Found response
func NotFound(w http.ResponseWriter, message string) {
	JSON(w, http.StatusNotFound, ErrorResponse{
		Success: false,
		Error:   "error",
		Code:    "NOT_FOUND",
		Message: message,
	})
}

func DecodeJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}
