package anomaly

import "time"

// Anomaly represents a cost anomaly
type Anomaly struct {
	ID                  int64     `json:"id"`
	UserID              int64     `json:"user_id"`
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
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at,omitempty"`
}

// Anomaly types
const (
	TypeCostSpike     = "cost_spike"
	TypeUnusualUsage  = "unusual_usage"
	TypeBudgetOverrun = "budget_overrun"
)

// Severity levels
const (
	SeverityCritical = "critical"
	SeverityHigh     = "high"
	SeverityMedium   = "medium"
	SeverityLow      = "low"
)

// Status
const (
	StatusDetected     = "detected"
	StatusAcknowledged = "acknowledged"
	StatusResolved     = "resolved"
	StatusIgnored      = "ignored"
)

// Filter contains anomaly filtering options
type Filter struct {
	ResourceID string
	Type       string
	Severity   string
	Status     string
}
