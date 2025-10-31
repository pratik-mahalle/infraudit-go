package client

import "fmt"

// APIError represents an error returned by the API
type APIError struct {
	StatusCode int                    `json:"-"`
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("API error [%s]: %s (status: %d)", e.Code, e.Message, e.StatusCode)
	}
	return fmt.Sprintf("API error: %s (status: %d)", e.Message, e.StatusCode)
}

// IsNotFound returns true if the error is a 404 not found error
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == 404
}

// IsUnauthorized returns true if the error is a 401 unauthorized error
func (e *APIError) IsUnauthorized() bool {
	return e.StatusCode == 401
}

// IsForbidden returns true if the error is a 403 forbidden error
func (e *APIError) IsForbidden() bool {
	return e.StatusCode == 403
}

// IsValidationError returns true if the error is a 400 validation error
func (e *APIError) IsValidationError() bool {
	return e.StatusCode == 400
}

// IsServerError returns true if the error is a 5xx server error
func (e *APIError) IsServerError() bool {
	return e.StatusCode >= 500
}
