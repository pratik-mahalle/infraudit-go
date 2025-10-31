package anomaly

import "context"

// Service defines the interface for anomaly business logic
type Service interface {
	// Create creates a new anomaly record
	Create(ctx context.Context, anomaly *Anomaly) (int64, error)

	// GetByID retrieves an anomaly by ID
	GetByID(ctx context.Context, userID int64, id int64) (*Anomaly, error)

	// Update updates an anomaly record
	Update(ctx context.Context, userID int64, id int64, updates map[string]interface{}) error

	// Delete deletes an anomaly record
	Delete(ctx context.Context, userID int64, id int64) error

	// List retrieves anomalies with filters and pagination
	List(ctx context.Context, userID int64, filter Filter, limit, offset int) ([]*Anomaly, int64, error)

	// UpdateStatus updates anomaly status
	UpdateStatus(ctx context.Context, userID int64, id int64, status string) error

	// DetectAnomalies detects cost anomalies for a user
	DetectAnomalies(ctx context.Context, userID int64) error

	// GetSummary gets anomaly summary by severity
	GetSummary(ctx context.Context, userID int64) (map[string]int, error)
}
