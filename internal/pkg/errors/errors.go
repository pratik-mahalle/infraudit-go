package errors

import (
	"fmt"
	"net/http"
)

// AppError represents an application error with additional context
type AppError struct {
	Code       string      `json:"code"`
	Message    string      `json:"message"`
	StatusCode int         `json:"-"`
	Internal   error       `json:"-"`
	Details    interface{} `json:"details,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Internal)
	}
	return e.Message
}

// Unwrap returns the internal error for errors.Is and errors.As
func (e *AppError) Unwrap() error {
	return e.Internal
}

// Common error codes
const (
	ErrCodeInternal          = "INTERNAL_ERROR"
	ErrCodeBadRequest        = "BAD_REQUEST"
	ErrCodeUnauthorized      = "UNAUTHORIZED"
	ErrCodeForbidden         = "FORBIDDEN"
	ErrCodeNotFound          = "NOT_FOUND"
	ErrCodeConflict          = "CONFLICT"
	ErrCodeValidation        = "VALIDATION_ERROR"
	ErrCodeDatabase          = "DATABASE_ERROR"
	ErrCodeProviderAuth      = "PROVIDER_AUTH_ERROR"
	ErrCodeProviderAPI       = "PROVIDER_API_ERROR"
	ErrCodeRateLimited       = "RATE_LIMITED"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
)

// New creates a new AppError
func New(code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// Wrap wraps an error with an AppError
func Wrap(err error, code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Internal:   err,
	}
}

// WithDetails adds details to an AppError
func (e *AppError) WithDetails(details interface{}) *AppError {
	e.Details = details
	return e
}

// Common error constructors

// Internal creates an internal server error
func Internal(message string, err error) *AppError {
	return Wrap(err, ErrCodeInternal, message, http.StatusInternalServerError)
}

// BadRequest creates a bad request error
func BadRequest(message string) *AppError {
	return New(ErrCodeBadRequest, message, http.StatusBadRequest)
}

// Unauthorized creates an unauthorized error
func Unauthorized(message string) *AppError {
	return New(ErrCodeUnauthorized, message, http.StatusUnauthorized)
}

// Forbidden creates a forbidden error
func Forbidden(message string) *AppError {
	return New(ErrCodeForbidden, message, http.StatusForbidden)
}

// NotFound creates a not found error
func NotFound(resource string) *AppError {
	return New(ErrCodeNotFound, fmt.Sprintf("%s not found", resource), http.StatusNotFound)
}

// Conflict creates a conflict error
func Conflict(message string) *AppError {
	return New(ErrCodeConflict, message, http.StatusConflict)
}

// ValidationError creates a validation error
func ValidationError(message string, details interface{}) *AppError {
	return New(ErrCodeValidation, message, http.StatusBadRequest).WithDetails(details)
}

// DatabaseError creates a database error
func DatabaseError(message string, err error) *AppError {
	return Wrap(err, ErrCodeDatabase, message, http.StatusInternalServerError)
}

// ProviderAuthError creates a provider authentication error
func ProviderAuthError(provider string, err error) *AppError {
	return Wrap(err, ErrCodeProviderAuth,
		fmt.Sprintf("Failed to authenticate with %s", provider),
		http.StatusUnauthorized)
}

// ProviderAPIError creates a provider API error
func ProviderAPIError(provider string, err error) *AppError {
	return Wrap(err, ErrCodeProviderAPI,
		fmt.Sprintf("Failed to communicate with %s API", provider),
		http.StatusBadGateway)
}

// RateLimited creates a rate limited error
func RateLimited(message string) *AppError {
	return New(ErrCodeRateLimited, message, http.StatusTooManyRequests)
}

// ServiceUnavailable creates a service unavailable error
func ServiceUnavailable(message string) *AppError {
	return New(ErrCodeServiceUnavailable, message, http.StatusServiceUnavailable)
}
