package anomaly

import "time"

// Anomaly represents a cost anomaly
type Anomaly struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	ResourceID   string    `json:"resource_id"`
	AnomalyType  string    `json:"anomaly_type"`
	Severity     string    `json:"severity"`
	Percentage   int       `json:"percentage"`
	PreviousCost int       `json:"previous_cost"`
	CurrentCost  int       `json:"current_cost"`
	DetectedAt   time.Time `json:"detected_at"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
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
