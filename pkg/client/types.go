package client

import "time"

// Resource represents a cloud resource
type Resource struct {
	ID           int64                  `json:"id"`
	UserID       int64                  `json:"user_id"`
	ProviderID   int64                  `json:"provider_id"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   string                 `json:"resource_id"`
	Name         string                 `json:"name"`
	Region       string                 `json:"region"`
	Status       string                 `json:"status"`
	Tags         map[string]string      `json:"tags,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Cost         *ResourceCost          `json:"cost,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// ResourceCost represents cost information for a resource
type ResourceCost struct {
	Daily   float64 `json:"daily"`
	Monthly float64 `json:"monthly"`
	Yearly  float64 `json:"yearly"`
}

// Provider represents a cloud provider account
type Provider struct {
	ID           int64                  `json:"id"`
	UserID       int64                  `json:"user_id"`
	Name         string                 `json:"name"`
	ProviderType string                 `json:"provider_type"` // aws, gcp, azure
	Status       string                 `json:"status"`        // connected, disconnected, error
	LastSyncedAt *time.Time             `json:"last_synced_at,omitempty"`
	Credentials  map[string]interface{} `json:"credentials,omitempty"`
	Config       map[string]interface{} `json:"config,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// Alert represents a security or operational alert
type Alert struct {
	ID          int64                  `json:"id"`
	UserID      int64                  `json:"user_id"`
	ResourceID  *int64                 `json:"resource_id,omitempty"`
	Type        string                 `json:"type"`        // security, compliance, performance
	Severity    string                 `json:"severity"`    // critical, high, medium, low
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Status      string                 `json:"status"`      // open, acknowledged, resolved
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
}

// Recommendation represents a cost or performance optimization recommendation
type Recommendation struct {
	ID                int64                  `json:"id"`
	UserID            int64                  `json:"user_id"`
	ResourceID        *int64                 `json:"resource_id,omitempty"`
	Type              string                 `json:"type"` // cost, performance, security
	Title             string                 `json:"title"`
	Description       string                 `json:"description"`
	Impact            string                 `json:"impact"` // high, medium, low
	EstimatedSavings  float64                `json:"estimated_savings,omitempty"`
	Status            string                 `json:"status"` // pending, applied, dismissed
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

// Drift represents a security configuration drift
type Drift struct {
	ID             int64                  `json:"id"`
	UserID         int64                  `json:"user_id"`
	ResourceID     int64                  `json:"resource_id"`
	DriftType      string                 `json:"drift_type"` // configuration, security, compliance
	Severity       string                 `json:"severity"`   // critical, high, medium, low
	Description    string                 `json:"description"`
	ExpectedConfig map[string]interface{} `json:"expected_config"`
	ActualConfig   map[string]interface{} `json:"actual_config"`
	Status         string                 `json:"status"` // detected, investigating, resolved
	DetectedAt     time.Time              `json:"detected_at"`
	ResolvedAt     *time.Time             `json:"resolved_at,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// Anomaly represents a cost anomaly
type Anomaly struct {
	ID               int64                  `json:"id"`
	UserID           int64                  `json:"user_id"`
	ResourceID       *int64                 `json:"resource_id,omitempty"`
	AnomalyType      string                 `json:"anomaly_type"` // cost_spike, unusual_usage
	Severity         string                 `json:"severity"`     // critical, high, medium, low
	Description      string                 `json:"description"`
	ExpectedValue    float64                `json:"expected_value"`
	ActualValue      float64                `json:"actual_value"`
	Deviation        float64                `json:"deviation"` // Percentage deviation
	Status           string                 `json:"status"`    // detected, investigating, resolved
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	DetectedAt       time.Time              `json:"detected_at"`
	ResolvedAt       *time.Time             `json:"resolved_at,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// ListOptions contains common options for list operations
type ListOptions struct {
	Page     int    `json:"page,omitempty"`      // Page number (1-based)
	PageSize int    `json:"page_size,omitempty"` // Items per page
	Sort     string `json:"sort,omitempty"`      // Sort field
	Order    string `json:"order,omitempty"`     // Sort order: asc, desc
	Search   string `json:"search,omitempty"`    // Search query
}

// ListResponse represents a paginated list response
type ListResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	Total      int64       `json:"total"`
	TotalPages int         `json:"total_pages"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Version   string            `json:"version,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services,omitempty"`
}
