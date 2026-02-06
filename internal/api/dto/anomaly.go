package dto

import "time"

// AnomalyDTO represents a cost anomaly in API responses
type AnomalyDTO struct {
	ID           int64     `json:"id"`
	ResourceID   string    `json:"resourceId"`
	AnomalyType  string    `json:"anomalyType"`
	Severity     string    `json:"severity"`
	Percentage   int       `json:"percentage"`
	PreviousCost int       `json:"previousCost"`
	CurrentCost  int       `json:"currentCost"`
	DetectedAt   time.Time `json:"detectedAt"`
	Status       string    `json:"status"`
}

// CreateAnomalyRequest represents an anomaly creation request
type CreateAnomalyRequest struct {
	ResourceID   string `json:"resourceId" validate:"required"`
	AnomalyType  string `json:"anomalyType" validate:"required"`
	Severity     string `json:"severity" validate:"required,oneof=critical high medium low"`
	Percentage   int    `json:"percentage" validate:"required"`
	PreviousCost int    `json:"previousCost" validate:"required"`
	CurrentCost  int    `json:"currentCost" validate:"required"`
	Status       string `json:"status,omitempty"`
}

// UpdateAnomalyRequest represents an anomaly update request
type UpdateAnomalyRequest struct {
	ResourceID   *string `json:"resourceId,omitempty"`
	AnomalyType  *string `json:"anomalyType,omitempty"`
	Severity     *string `json:"severity,omitempty"`
	Percentage   *int    `json:"percentage,omitempty"`
	PreviousCost *int    `json:"previousCost,omitempty"`
	CurrentCost  *int    `json:"currentCost,omitempty"`
	Status       *string `json:"status,omitempty"`
}

// AnomalyListRequest represents anomaly list query parameters
type AnomalyListRequest struct {
	ResourceID string `json:"resourceId,omitempty"`
	Type       string `json:"type,omitempty"`
	Severity   string `json:"severity,omitempty"`
	Status     string `json:"status,omitempty"`
	Page       int    `json:"page,omitempty"`
	PageSize   int    `json:"pageSize,omitempty"`
}
