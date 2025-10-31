package provider

import "context"

// Service defines the interface for provider business logic
type Service interface {
	// Connect connects a cloud provider account
	Connect(ctx context.Context, userID int64, providerType string, credentials Credentials) error

	// Disconnect disconnects a cloud provider account
	Disconnect(ctx context.Context, userID int64, providerType string) error

	// List retrieves all connected providers for a user
	List(ctx context.Context, userID int64) ([]*Provider, error)

	// GetByProvider retrieves a specific provider account
	GetByProvider(ctx context.Context, userID int64, providerType string) (*Provider, error)

	// TestConnection tests the connection to a provider
	TestConnection(ctx context.Context, providerType string, credentials Credentials) error

	// Sync syncs resources from a provider
	Sync(ctx context.Context, userID int64, providerType string) error

	// GetSyncStatus gets the sync status for all providers
	GetSyncStatus(ctx context.Context, userID int64) ([]*SyncStatus, error)
}
