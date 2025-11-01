package baseline

import "context"

// Repository defines the interface for baseline data access
type Repository interface {
	// Create creates a new baseline
	Create(ctx context.Context, baseline *Baseline) (int64, error)

	// GetByResourceID retrieves baseline for a resource
	GetByResourceID(ctx context.Context, userID int64, resourceID string, baselineType string) (*Baseline, error)

	// Update updates a baseline
	Update(ctx context.Context, baseline *Baseline) error

	// Delete deletes a baseline
	Delete(ctx context.Context, userID int64, id int64) error

	// List retrieves baselines for a user
	List(ctx context.Context, userID int64) ([]*Baseline, error)
}
