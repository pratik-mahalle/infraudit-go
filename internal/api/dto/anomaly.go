package dto

import "time"

// AnomalyDTO represents a cost anomaly in API responses
type AnomalyDTO struct {
	ID           int64     `json:"id"`
	ResourceID   string    `json:"resource_id"`
	AnomalyType  string    `json:"anomaly_type"`
	Severity     string    `json:"severity"`
	Percentage   int       `json:"percentage"`
	PreviousCost int       `json:"previous_cost"`
	CurrentCost  int       `json:"current_cost"`
	DetectedAt   time.Time `json:"detected_at"`
	Status       string    `json:"status"`
}

// CreateAnomalyRequest represents an anomaly creation request
type CreateAnomalyRequest struct {
	ResourceID   string `json:"resource_id" validate:"required"`
	AnomalyType  string `json:"anomaly_type" validate:"required"`
	Severity     string `json:"severity" validate:"required,oneof=critical high medium low"`
	Percentage   int    `json:"percentage" validate:"required"`
	PreviousCost int    `json:"previous_cost" validate:"required"`
	CurrentCost  int    `json:"current_cost" validate:"required"`
	Status       string `json:"status,omitempty"`
}

// UpdateAnomalyRequest represents an anomaly update request
type UpdateAnomalyRequest struct {
	ResourceID   *string `json:"resource_id,omitempty"`
	AnomalyType  *string `json:"anomaly_type,omitempty"`
	Severity     *string `json:"severity,omitempty"`
	Percentage   *int    `json:"percentage,omitempty"`
	PreviousCost *int    `json:"previous_cost,omitempty"`
	CurrentCost  *int    `json:"current_cost,omitempty"`
	Status       *string `json:"status,omitempty"`
}

// AnomalyListRequest represents anomaly list query parameters
type AnomalyListRequest struct {
	ResourceID string `json:"resource_id,omitempty"`
	Type       string `json:"type,omitempty"`
	Severity   string `json:"severity,omitempty"`
	Status     string `json:"status,omitempty"`
	Page       int    `json:"page,omitempty"`
	PageSize   int    `json:"page_size,omitempty"`
}
