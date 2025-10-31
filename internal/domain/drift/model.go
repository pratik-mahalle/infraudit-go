package drift

import "time"

// Drift represents a security configuration drift
type Drift struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	ResourceID string    `json:"resource_id"`
	DriftType  string    `json:"drift_type"`
	Severity   string    `json:"severity"`
	Details    string    `json:"details"`
	DetectedAt time.Time `json:"detected_at"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at,omitempty"`
}

// Drift types
const (
	TypeConfigurationChange = "configuration_change"
	TypeSecurityGroup       = "security_group"
	TypeIAMPolicy           = "iam_policy"
	TypeNetworkRule         = "network_rule"
	TypeEncryption          = "encryption"
	TypeCompliance          = "compliance"
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

// Filter contains drift filtering options
type Filter struct {
	ResourceID string
	DriftType  string
	Severity   string
	Status     string
}
