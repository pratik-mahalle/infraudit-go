package resource

import "context"

// Repository defines the interface for resource data access
type Repository interface {
	// Create creates a new resource
	Create(ctx context.Context, resource *Resource) error

	// GetByID retrieves a resource by ID
	GetByID(ctx context.Context, userID int64, resourceID string) (*Resource, error)

	// Update updates a resource
	Update(ctx context.Context, resource *Resource) error

	// Delete deletes a resource
	Delete(ctx context.Context, userID int64, resourceID string) error

	// List retrieves resources with filters and pagination
	List(ctx context.Context, userID int64, filter Filter, limit, offset int) ([]*Resource, int64, error)

	// ListByProvider retrieves resources by provider
	ListByProvider(ctx context.Context, userID int64, provider string) ([]*Resource, error)

	// SaveBatch saves multiple resources (used for sync)
	SaveBatch(ctx context.Context, userID int64, provider string, resources []*Resource) error

	// DeleteByProvider deletes all resources for a provider
	DeleteByProvider(ctx context.Context, userID int64, provider string) error
}
