package alert

import "time"

// Alert represents a security or operational alert
type Alert struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Resource    string    `json:"resource,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}

// Alert types
const (
	TypeSecurity     = "security"
	TypeCompliance   = "compliance"
	TypePerformance  = "performance"
	TypeCost         = "cost"
	TypeAvailability = "availability"
)

// Alert severity levels
const (
	SeverityCritical = "critical"
	SeverityHigh     = "high"
	SeverityMedium   = "medium"
	SeverityLow      = "low"
	SeverityInfo     = "info"
)

// Alert status
const (
	StatusOpen       = "open"
	StatusAcknowledged = "acknowledged"
	StatusInProgress = "in_progress"
	StatusResolved   = "resolved"
	StatusClosed     = "closed"
)

// Filter contains alert filtering options
type Filter struct {
	Type     string
	Severity string
	Status   string
	Resource string
}
