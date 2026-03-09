package dto

import "time"

// AnomalyDTO represents a cost anomaly in API responses
type AnomalyDTO struct {
	ID                  int64     `json:"id"`
	AnomalyType         string    `json:"anomaly_type,omitempty"`
	ServiceName         string    `json:"service_name,omitempty"`
	Region              string    `json:"region,omitempty"`
	Severity            string    `json:"severity"`
	DeviationPercentage float64   `json:"deviation_percentage"`
	ExpectedCost        float64   `json:"expected_cost"`
	ActualCost          float64   `json:"actual_cost"`
	Description         string    `json:"description,omitempty"`
	DetectedAt          time.Time `json:"detected_at"`
	Status              string    `json:"status"`
}

// CreateAnomalyRequest represents an anomaly creation request
type CreateAnomalyRequest struct {
	AnomalyType         string  `json:"anomaly_type"`
	ServiceName         string  `json:"service_name"`
	Region              string  `json:"region,omitempty"`
	Severity            string  `json:"severity" validate:"required,oneof=critical high medium low"`
	DeviationPercentage float64 `json:"deviation_percentage" validate:"required"`
	ExpectedCost        float64 `json:"expected_cost" validate:"required"`
	ActualCost          float64 `json:"actual_cost" validate:"required"`
	Description         string  `json:"description,omitempty"`
	Status              string  `json:"status,omitempty"`
}

// UpdateAnomalyRequest represents an anomaly update request
type UpdateAnomalyRequest struct {
	AnomalyType         *string  `json:"anomaly_type,omitempty"`
	Severity            *string  `json:"severity,omitempty"`
	DeviationPercentage *float64 `json:"deviation_percentage,omitempty"`
	ExpectedCost        *float64 `json:"expected_cost,omitempty"`
	ActualCost          *float64 `json:"actual_cost,omitempty"`
	Description         *string  `json:"description,omitempty"`
	Status              *string  `json:"status,omitempty"`
}

// AnomalyListRequest represents anomaly list query parameters
type AnomalyListRequest struct {
	Type     string `json:"type,omitempty"`
	Severity string `json:"severity,omitempty"`
	Status   string `json:"status,omitempty"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
}
