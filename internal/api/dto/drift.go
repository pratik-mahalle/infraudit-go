package dto

import "time"

// DriftDTO represents a security drift in API responses
type DriftDTO struct {
	ID         int64     `json:"id"`
	ResourceID string    `json:"resource_id"`
	DriftType  string    `json:"drift_type"`
	Severity   string    `json:"severity"`
	Details    string    `json:"details"`
	DetectedAt time.Time `json:"detected_at"`
	Status     string    `json:"status"`
}

// CreateDriftRequest represents a drift creation request
type CreateDriftRequest struct {
	ResourceID string `json:"resource_id" validate:"required"`
	DriftType  string `json:"drift_type" validate:"required"`
	Severity   string `json:"severity" validate:"required,oneof=critical high medium low"`
	Details    string `json:"details" validate:"required"`
	Status     string `json:"status,omitempty"`
}

// UpdateDriftRequest represents a drift update request
type UpdateDriftRequest struct {
	ResourceID *string `json:"resource_id,omitempty"`
	DriftType  *string `json:"drift_type,omitempty"`
	Severity   *string `json:"severity,omitempty"`
	Details    *string `json:"details,omitempty"`
	Status     *string `json:"status,omitempty"`
}

// DriftListRequest represents drift list query parameters
type DriftListRequest struct {
	ResourceID string `json:"resource_id,omitempty"`
	DriftType  string `json:"drift_type,omitempty"`
	Severity   string `json:"severity,omitempty"`
	Status     string `json:"status,omitempty"`
	Page       int    `json:"page,omitempty"`
	PageSize   int    `json:"page_size,omitempty"`
}
