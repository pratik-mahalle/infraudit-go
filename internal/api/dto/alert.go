package dto

import "time"

// AlertDTO represents an alert in API responses
type AlertDTO struct {
	ID          int64     `json:"id"`
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Resource    string    `json:"resource,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateAlertRequest represents an alert creation request
type CreateAlertRequest struct {
	Type        string `json:"type" validate:"required,oneof=security compliance performance cost availability"`
	Severity    string `json:"severity" validate:"required,oneof=critical high medium low info"`
	Title       string `json:"title" validate:"required"`
	Description string `json:"description" validate:"required"`
	Resource    string `json:"resource,omitempty"`
	Status      string `json:"status,omitempty"`
}

// UpdateAlertRequest represents an alert update request
type UpdateAlertRequest struct {
	Type        *string `json:"type,omitempty"`
	Severity    *string `json:"severity,omitempty"`
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Resource    *string `json:"resource,omitempty"`
	Status      *string `json:"status,omitempty"`
}

// AlertListRequest represents alert list query parameters
type AlertListRequest struct {
	Type     string `json:"type,omitempty"`
	Severity string `json:"severity,omitempty"`
	Status   string `json:"status,omitempty"`
	Resource string `json:"resource,omitempty"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
}
