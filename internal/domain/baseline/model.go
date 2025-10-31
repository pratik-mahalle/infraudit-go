package baseline

import "time"

// Baseline represents a stored configuration snapshot of a resource
type Baseline struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"user_id"`
	ResourceID    string    `json:"resource_id"`
	Provider      string    `json:"provider"`
	ResourceType  string    `json:"resource_type"`
	Configuration string    `json:"configuration"` // JSON string of resource configuration
	BaselineType  string    `json:"baseline_type"` // manual, automatic, approved
	Description   string    `json:"description,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"`
}

// Baseline types
const (
	TypeManual    = "manual"    // Manually created baseline
	TypeAutomatic = "automatic" // Auto-created on first scan
	TypeApproved  = "approved"  // Approved configuration state
)
