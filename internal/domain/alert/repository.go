package alert

import "context"

// Repository defines the interface for alert data access
type Repository interface {
	// Create creates a new alert
	Create(ctx context.Context, alert *Alert) (int64, error)

	// GetByID retrieves an alert by ID
	GetByID(ctx context.Context, userID int64, id int64) (*Alert, error)

	// Update updates an alert
	Update(ctx context.Context, alert *Alert) error

	// Delete deletes an alert
	Delete(ctx context.Context, userID int64, id int64) error

	// List retrieves alerts with filters
	List(ctx context.Context, userID int64, filter Filter) ([]*Alert, error)

	// ListWithPagination retrieves alerts with filters and pagination
	ListWithPagination(ctx context.Context, userID int64, filter Filter, limit, offset int) ([]*Alert, int64, error)

	// CountByStatus counts alerts by status
	CountByStatus(ctx context.Context, userID int64) (map[string]int, error)
}
