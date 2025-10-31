package alert

import "context"

// Service defines the interface for alert business logic
type Service interface {
	// Create creates a new alert
	Create(ctx context.Context, alert *Alert) (int64, error)

	// GetByID retrieves an alert by ID
	GetByID(ctx context.Context, userID int64, id int64) (*Alert, error)

	// Update updates an alert
	Update(ctx context.Context, userID int64, id int64, updates map[string]interface{}) error

	// Delete deletes an alert
	Delete(ctx context.Context, userID int64, id int64) error

	// List retrieves alerts with filters and pagination
	List(ctx context.Context, userID int64, filter Filter, limit, offset int) ([]*Alert, int64, error)

	// UpdateStatus updates alert status
	UpdateStatus(ctx context.Context, userID int64, id int64, status string) error

	// GetSummary gets alert summary by status
	GetSummary(ctx context.Context, userID int64) (map[string]int, error)
}
