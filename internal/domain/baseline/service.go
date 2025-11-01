package baseline

import "context"

// Service defines the business logic for baseline management
type Service interface {
	// CreateBaseline creates a new baseline for a resource
	CreateBaseline(ctx context.Context, baseline *Baseline) (int64, error)

	// GetBaseline retrieves baseline for a resource
	GetBaseline(ctx context.Context, userID int64, resourceID string, baselineType string) (*Baseline, error)

	// UpdateBaseline updates an existing baseline
	UpdateBaseline(ctx context.Context, userID int64, id int64, updates map[string]interface{}) error

	// DeleteBaseline deletes a baseline
	DeleteBaseline(ctx context.Context, userID int64, id int64) error

	// ListBaselines lists all baselines for a user
	ListBaselines(ctx context.Context, userID int64) ([]*Baseline, error)
}
