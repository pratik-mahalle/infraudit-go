package provider

import (
	"context"
	"time"
)

// Repository defines the interface for provider data access
type Repository interface {
	// Upsert creates or updates a provider account
	Upsert(ctx context.Context, provider *Provider) error

	// GetByProvider retrieves a provider account by provider type
	GetByProvider(ctx context.Context, userID int64, providerType string) (*Provider, error)

	// List retrieves all provider accounts for a user
	List(ctx context.Context, userID int64) ([]*Provider, error)

	// Delete deletes a provider account
	Delete(ctx context.Context, userID int64, providerType string) error

	// UpdateSyncStatus updates the sync status
	UpdateSyncStatus(ctx context.Context, userID int64, providerType string, lastSynced time.Time) error

	// UpdateConnectionStatus updates the connection status
	UpdateConnectionStatus(ctx context.Context, userID int64, providerType string, isConnected bool) error
}
