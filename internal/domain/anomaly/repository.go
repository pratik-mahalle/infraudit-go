package anomaly

import "context"

// Repository defines the interface for anomaly data access
type Repository interface {
	// Create creates a new anomaly record
	Create(ctx context.Context, anomaly *Anomaly) (int64, error)

	// GetByID retrieves an anomaly by ID
	GetByID(ctx context.Context, userID int64, id int64) (*Anomaly, error)

	// Update updates an anomaly record
	Update(ctx context.Context, anomaly *Anomaly) error

	// Delete deletes an anomaly record
	Delete(ctx context.Context, userID int64, id int64) error

	// List retrieves anomalies with filters
	List(ctx context.Context, userID int64, filter Filter) ([]*Anomaly, error)

	// ListWithPagination retrieves anomalies with filters and pagination
	ListWithPagination(ctx context.Context, userID int64, filter Filter, limit, offset int) ([]*Anomaly, int64, error)

	// CountBySeverity counts anomalies by severity
	CountBySeverity(ctx context.Context, userID int64) (map[string]int, error)
}
