package cost

import (
	"encoding/json"
	"time"
)

// Cost represents a cost record for a resource
type Cost struct {
	ID           string          `json:"id"`
	UserID       int64           `json:"user_id"`
	ResourceID   *string         `json:"resource_id,omitempty"`
	Provider     string          `json:"provider"`
	Region       string          `json:"region,omitempty"`
	ServiceName  string          `json:"service_name"`
	ResourceType string          `json:"resource_type,omitempty"`
	CostDate     time.Time       `json:"cost_date"`
	DailyCost    float64         `json:"daily_cost"`
	MonthlyCost  float64         `json:"monthly_cost,omitempty"`
	Currency     string          `json:"currency"`
	CostDetails  json.RawMessage `json:"cost_details,omitempty"`
	Tags         json.RawMessage `json:"tags,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// CostSummary represents aggregated cost data
type CostSummary struct {
	Provider   string             `json:"provider"`
	TotalCost  float64            `json:"total_cost"`
	Currency   string             `json:"currency"`
	Period     string             `json:"period"`
	StartDate  time.Time          `json:"start_date"`
	EndDate    time.Time          `json:"end_date"`
	ByService  map[string]float64 `json:"by_service"`
	ByRegion   map[string]float64 `json:"by_region"`
	ByResource map[string]float64 `json:"by_resource,omitempty"`
}

// CostTrend represents cost changes over time
type CostTrend struct {
	Period        string          `json:"period"` // daily, weekly, monthly
	CurrentCost   float64         `json:"current_cost"`
	PreviousCost  float64         `json:"previous_cost"`
	Change        float64         `json:"change"` // absolute change
	ChangePercent float64         `json:"change_percent"`
	Trend         string          `json:"trend"` // up, down, stable
	DataPoints    []CostDataPoint `json:"data_points"`
}

// CostDataPoint represents a single cost data point
type CostDataPoint struct {
	Date time.Time `json:"date"`
	Cost float64   `json:"cost"`
}

// CostForecast represents predicted future costs
type CostForecast struct {
	Provider        string    `json:"provider"`
	Period          string    `json:"period"`
	ForecastedCost  float64   `json:"forecasted_cost"`
	ConfidenceLevel float64   `json:"confidence_level"` // 0-1
	LowerBound      float64   `json:"lower_bound"`
	UpperBound      float64   `json:"upper_bound"`
	Currency        string    `json:"currency"`
	EndDate         time.Time `json:"end_date"`
}

// CostAnomaly represents unusual cost patterns
type CostAnomaly struct {
	ID           string    `json:"id"`
	UserID       int64     `json:"user_id"`
	Provider     string    `json:"provider"`
	ServiceName  string    `json:"service_name"`
	ResourceID   *string   `json:"resource_id,omitempty"`
	AnomalyType  string    `json:"anomaly_type"` // spike, drop, unusual_pattern
	ExpectedCost float64   `json:"expected_cost"`
	ActualCost   float64   `json:"actual_cost"`
	Deviation    float64   `json:"deviation"` // percentage deviation
	Severity     string    `json:"severity"`
	DetectedAt   time.Time `json:"detected_at"`
	Status       string    `json:"status"` // open, reviewed, resolved
	Notes        string    `json:"notes,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// AnomalyType constants
const (
	AnomalyTypeSpike          = "spike"
	AnomalyTypeDrop           = "drop"
	AnomalyTypeUnusualPattern = "unusual_pattern"
)

// AnomalyStatus constants
const (
	AnomalyStatusOpen     = "open"
	AnomalyStatusReviewed = "reviewed"
	AnomalyStatusResolved = "resolved"
)

// CostOptimization represents a cost savings opportunity
type CostOptimization struct {
	ID               string          `json:"id"`
	UserID           int64           `json:"user_id"`
	Provider         string          `json:"provider"`
	ResourceID       *string         `json:"resource_id,omitempty"`
	ResourceType     string          `json:"resource_type"`
	OptimizationType string          `json:"optimization_type"`
	Title            string          `json:"title"`
	Description      string          `json:"description"`
	CurrentCost      float64         `json:"current_cost"`
	EstimatedSavings float64         `json:"estimated_savings"`
	SavingsPercent   float64         `json:"savings_percent"`
	Implementation   string          `json:"implementation"` // easy, moderate, complex
	Status           string          `json:"status"`         // pending, applied, dismissed
	Details          json.RawMessage `json:"details,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

// OptimizationType constants
const (
	OptTypeRightsize          = "rightsize"
	OptTypeUnused             = "unused"
	OptTypeReservedInstance   = "reserved_instance"
	OptTypeSavingsPlan        = "savings_plan"
	OptTypeStorageClass       = "storage_class"
	OptTypeIdleResource       = "idle_resource"
	OptTypeRegionOptimization = "region_optimization"
)

// OptimizationStatus constants
const (
	OptStatusPending   = "pending"
	OptStatusApplied   = "applied"
	OptStatusDismissed = "dismissed"
)

// Filter contains cost query filters
type Filter struct {
	Provider    string
	ServiceName string
	ResourceID  string
	Region      string
	StartDate   *time.Time
	EndDate     *time.Time
	GroupBy     string // service, region, resource, tag
}

// Provider constants
const (
	ProviderAWS   = "aws"
	ProviderGCP   = "gcp"
	ProviderAzure = "azure"
)
