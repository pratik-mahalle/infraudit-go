package recommendation

import "time"

// Recommendation represents a cost or performance recommendation
type Recommendation struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Type        string    `json:"type"`
	Priority    string    `json:"priority"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Savings     float64   `json:"savings,omitempty"`
	Effort      string    `json:"effort"`
	Impact      string    `json:"impact"`
	Category    string    `json:"category"`
	Resources   []string  `json:"resources,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}

// Recommendation types
const (
	TypeCostOptimization        = "cost_optimization"
	TypePerformanceOptimization = "performance_optimization"
	TypeSecurityImprovement     = "security_improvement"
	TypeReliabilityImprovement  = "reliability_improvement"
	TypeCompliance              = "compliance"
)

// Priority levels
const (
	PriorityCritical = "critical"
	PriorityHigh     = "high"
	PriorityMedium   = "medium"
	PriorityLow      = "low"
)

// Effort levels
const (
	EffortLow    = "low"
	EffortMedium = "medium"
	EffortHigh   = "high"
)

// Impact levels
const (
	ImpactHigh   = "high"
	ImpactMedium = "medium"
	ImpactLow    = "low"
)

// Filter contains recommendation filtering options
type Filter struct {
	Type     string
	Priority string
	Category string
}
