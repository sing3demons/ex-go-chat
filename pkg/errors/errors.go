package errors

import (
	"fmt"
	"net/http"
)

// AppError represents an application error with HTTP status code
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
	Details    any    `json:"details,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewAppError creates a new application error
func NewAppError(code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(details any) *AppError {
	e.Details = details
	return e
}

// Common error constructors
var (
	// Authentication errors
	ErrInvalidCredentials = func() *AppError {
		return NewAppError("INVALID_CREDENTIALS", "Invalid username/email or password", http.StatusUnauthorized)
	}
	ErrInvalidToken = func() *AppError {
		return NewAppError("INVALID_TOKEN", "Invalid or expired token", http.StatusUnauthorized)
	}
	ErrMissingToken = func() *AppError {
		return NewAppError("MISSING_TOKEN", "Authentication token is required", http.StatusUnauthorized)
	}

	// Authorization errors
	ErrUnauthorized = func(message string) *AppError {
		return NewAppError("UNAUTHORIZED", message, http.StatusForbidden)
	}
	ErrNotRoomMember = func() *AppError {
		return NewAppError("NOT_ROOM_MEMBER", "You are not a member of this room", http.StatusForbidden)
	}
	ErrNotMessageOwner = func() *AppError {
		return NewAppError("NOT_MESSAGE_OWNER", "You can only edit/delete your own messages", http.StatusForbidden)
	}

	// Validation errors
	ErrInvalidInput = func(message string) *AppError {
		return NewAppError("INVALID_INPUT", message, http.StatusBadRequest)
	}
	ErrDuplicateUsername = func() *AppError {
		return NewAppError("DUPLICATE_USERNAME", "Username already exists", http.StatusConflict)
	}
	ErrDuplicateEmail = func() *AppError {
		return NewAppError("DUPLICATE_EMAIL", "Email already exists", http.StatusConflict)
	}
	ErrWeakPassword = func() *AppError {
		return NewAppError("WEAK_PASSWORD", "Password must be at least 8 characters with uppercase, lowercase, and number", http.StatusBadRequest)
	}
	ErrInvalidEmail = func() *AppError {
		return NewAppError("INVALID_EMAIL", "Invalid email format", http.StatusBadRequest)
	}

	// Resource errors
	ErrNotFound = func(resource string) *AppError {
		return NewAppError("NOT_FOUND", fmt.Sprintf("%s not found", resource), http.StatusNotFound)
	}
	ErrUserNotFound = func() *AppError {
		return ErrNotFound("User")
	}
	ErrRoomNotFound = func() *AppError {
		return ErrNotFound("Room")
	}
	ErrMessageNotFound = func() *AppError {
		return ErrNotFound("Message")
	}

	// Database errors
	ErrDatabase = func(err error) *AppError {
		return NewAppError("DATABASE_ERROR", "Database operation failed", http.StatusInternalServerError).WithDetails(err.Error())
	}

	// Internal errors
	ErrInternal = func(message string) *AppError {
		return NewAppError("INTERNAL_ERROR", message, http.StatusInternalServerError)
	}
)
