package dto

import "time"

// DriftDTO represents a security drift in API responses
// Uses camelCase for frontend compatibility
type DriftDTO struct {
	ID              int64                  `json:"id"`
	ResourceID      int64                  `json:"resourceId"`
	ResourceIDStr   string                 `json:"resourceIdStr,omitempty"`
	DriftType       string                 `json:"driftType"`
	Severity        string                 `json:"severity"`
	Details         map[string]interface{} `json:"details,omitempty"`
	Description     string                 `json:"description"`
	DetectedAt      time.Time              `json:"detectedAt"`
	Status          string                 `json:"status"`
	BaselineConfig  string                 `json:"baselineConfig,omitempty"`
	CurrentConfig   string                 `json:"currentConfig,omitempty"`
	RemediationTips []string               `json:"remediationTips,omitempty"`
}

// CreateDriftRequest represents a drift creation request
type CreateDriftRequest struct {
	ResourceID string `json:"resourceId" validate:"required"`
	DriftType  string `json:"driftType" validate:"required"`
	Severity   string `json:"severity" validate:"required,oneof=critical high medium low"`
	Details    string `json:"details" validate:"required"`
	Status     string `json:"status,omitempty"`
}

// UpdateDriftRequest represents a drift update request
type UpdateDriftRequest struct {
	ResourceID *string `json:"resourceId,omitempty"`
	DriftType  *string `json:"driftType,omitempty"`
	Severity   *string `json:"severity,omitempty"`
	Details    *string `json:"details,omitempty"`
	Status     *string `json:"status,omitempty"`
}

// DriftListRequest represents drift list query parameters
type DriftListRequest struct {
	ResourceID string `json:"resourceId,omitempty"`
	DriftType  string `json:"driftType,omitempty"`
	Severity   string `json:"severity,omitempty"`
	Status     string `json:"status,omitempty"`
	Page       int    `json:"page,omitempty"`
	PageSize   int    `json:"pageSize,omitempty"`
}

// DriftSummaryDTO represents drift summary statistics
type DriftSummaryDTO struct {
	Total      int            `json:"total"`
	Critical   int            `json:"critical"`
	High       int            `json:"high"`
	Medium     int            `json:"medium"`
	Low        int            `json:"low"`
	Open       int            `json:"open"`
	Remediated int            `json:"remediated"`
	ByType     map[string]int `json:"byType,omitempty"`
}
