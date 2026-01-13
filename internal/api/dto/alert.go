package dto

import "time"

// AlertDTO represents an alert in API responses
// Uses camelCase for frontend compatibility
type AlertDTO struct {
	ID          int64     `json:"id"`
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Title       string    `json:"title"`
	Message     string    `json:"message"`
	Description string    `json:"description,omitempty"`
	ResourceID  int64     `json:"resourceId,omitempty"`
	Resource    string    `json:"resource,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
}

// CreateAlertRequest represents an alert creation request
type CreateAlertRequest struct {
	Type        string `json:"type" validate:"required,oneof=security compliance performance cost availability resource"`
	Severity    string `json:"severity" validate:"required,oneof=critical high medium low info"`
	Title       string `json:"title" validate:"required"`
	Message     string `json:"message" validate:"required"`
	Description string `json:"description,omitempty"`
	ResourceID  int64  `json:"resourceId,omitempty"`
	Resource    string `json:"resource,omitempty"`
	Status      string `json:"status,omitempty"`
}

// UpdateAlertRequest represents an alert update request
type UpdateAlertRequest struct {
	Type        *string `json:"type,omitempty"`
	Severity    *string `json:"severity,omitempty"`
	Title       *string `json:"title,omitempty"`
	Message     *string `json:"message,omitempty"`
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
	PageSize int    `json:"pageSize,omitempty"`
}

// AlertSummaryDTO represents alert summary statistics
type AlertSummaryDTO struct {
	Total        int            `json:"total"`
	Critical     int            `json:"critical"`
	High         int            `json:"high"`
	Medium       int            `json:"medium"`
	Low          int            `json:"low"`
	Open         int            `json:"open"`
	Acknowledged int            `json:"acknowledged"`
	Resolved     int            `json:"resolved"`
	ByType       map[string]int `json:"byType,omitempty"`
}
