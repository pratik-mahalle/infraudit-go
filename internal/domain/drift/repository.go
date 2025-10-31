package drift

import "context"

// Repository defines the interface for drift data access
type Repository interface {
	// Create creates a new drift record
	Create(ctx context.Context, drift *Drift) (int64, error)

	// GetByID retrieves a drift by ID
	GetByID(ctx context.Context, userID int64, id int64) (*Drift, error)

	// Update updates a drift record
	Update(ctx context.Context, drift *Drift) error

	// Delete deletes a drift record
	Delete(ctx context.Context, userID int64, id int64) error

	// List retrieves drifts with filters
	List(ctx context.Context, userID int64, filter Filter) ([]*Drift, error)

	// ListWithPagination retrieves drifts with filters and pagination
	ListWithPagination(ctx context.Context, userID int64, filter Filter, limit, offset int) ([]*Drift, int64, error)

	// CountBySeverity counts drifts by severity
	CountBySeverity(ctx context.Context, userID int64) (map[string]int, error)
}
