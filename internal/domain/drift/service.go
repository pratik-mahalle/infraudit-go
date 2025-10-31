package drift

import "context"

// Service defines the interface for drift business logic
type Service interface {
	// Create creates a new drift record
	Create(ctx context.Context, drift *Drift) (int64, error)

	// GetByID retrieves a drift by ID
	GetByID(ctx context.Context, userID int64, id int64) (*Drift, error)

	// Update updates a drift record
	Update(ctx context.Context, userID int64, id int64, updates map[string]interface{}) error

	// Delete deletes a drift record
	Delete(ctx context.Context, userID int64, id int64) error

	// List retrieves drifts with filters and pagination
	List(ctx context.Context, userID int64, filter Filter, limit, offset int) ([]*Drift, int64, error)

	// UpdateStatus updates drift status
	UpdateStatus(ctx context.Context, userID int64, id int64, status string) error

	// DetectDrifts detects configuration drifts for a user
	DetectDrifts(ctx context.Context, userID int64) error

	// GetSummary gets drift summary by severity
	GetSummary(ctx context.Context, userID int64) (map[string]int, error)
}
