package resource

import "context"

// Service defines the interface for resource business logic
type Service interface {
	// GetByID retrieves a resource by ID
	GetByID(ctx context.Context, userID int64, resourceID string) (*Resource, error)

	// List retrieves resources with filters and pagination
	List(ctx context.Context, userID int64, filter Filter, limit, offset int) ([]*Resource, int64, error)

	// Create creates a new resource
	Create(ctx context.Context, resource *Resource) error

	// Update updates a resource
	Update(ctx context.Context, userID int64, resourceID string, updates map[string]interface{}) error

	// Delete deletes a resource
	Delete(ctx context.Context, userID int64, resourceID string) error

	// SyncProviderResources syncs resources from a cloud provider
	SyncProviderResources(ctx context.Context, userID int64, provider string) error
}
