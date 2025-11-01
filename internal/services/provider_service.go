package services

import (
	"context"
	"time"

	"github.com/pratik-mahalle/infraudit/internal/domain/provider"
	"github.com/pratik-mahalle/infraudit/internal/domain/resource"
	"github.com/pratik-mahalle/infraudit/internal/pkg/errors"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
)

// ProviderService implements provider.Service
type ProviderService struct {
	providerRepo provider.Repository
	resourceRepo resource.Repository
	logger       *logger.Logger
}

// NewProviderService creates a new provider service
func NewProviderService(providerRepo provider.Repository, resourceRepo resource.Repository, log *logger.Logger) provider.Service {
	return &ProviderService{
		providerRepo: providerRepo,
		resourceRepo: resourceRepo,
		logger:       log,
	}
}

// Connect connects a cloud provider account
func (s *ProviderService) Connect(ctx context.Context, userID int64, providerType string, credentials provider.Credentials) error {
	p := &provider.Provider{
		UserID:      userID,
		Provider:    providerType,
		IsConnected: true,
		Credentials: credentials,
	}

	// Test connection before saving
	if err := s.TestConnection(ctx, providerType, credentials); err != nil {
		s.logger.ErrorWithErr(err, "Provider connection test failed")
		return errors.ProviderAuthError(providerType, err)
	}

	err := s.providerRepo.Upsert(ctx, p)
	if err != nil {
		s.logger.ErrorWithErr(err, "Failed to save provider")
		return err
	}

	s.logger.WithFields(map[string]interface{}{
		"user_id":  userID,
		"provider": providerType,
	}).Info("Provider connected")

	return nil
}

// Disconnect disconnects a cloud provider account
func (s *ProviderService) Disconnect(ctx context.Context, userID int64, providerType string) error {
	// Delete provider account
	err := s.providerRepo.Delete(ctx, userID, providerType)
	if err != nil {
		s.logger.ErrorWithErr(err, "Failed to disconnect provider")
		return err
	}

	// Delete associated resources
	err = s.resourceRepo.DeleteByProvider(ctx, userID, providerType)
	if err != nil {
		s.logger.Warnf("Failed to delete resources for provider %s: %v", providerType, err)
	}

	s.logger.WithFields(map[string]interface{}{
		"user_id":  userID,
		"provider": providerType,
	}).Info("Provider disconnected")

	return nil
}

// List retrieves all connected providers for a user
func (s *ProviderService) List(ctx context.Context, userID int64) ([]*provider.Provider, error) {
	return s.providerRepo.List(ctx, userID)
}

// GetByProvider retrieves a specific provider account
func (s *ProviderService) GetByProvider(ctx context.Context, userID int64, providerType string) (*provider.Provider, error) {
	return s.providerRepo.GetByProvider(ctx, userID, providerType)
}

// TestConnection tests the connection to a provider
func (s *ProviderService) TestConnection(ctx context.Context, providerType string, credentials provider.Credentials) error {
	// This would use the actual cloud provider SDKs to test credentials
	// For now, basic validation
	switch providerType {
	case provider.ProviderAWS:
		if credentials.AWSAccessKeyID == "" || credentials.AWSSecretAccessKey == "" {
			return errors.BadRequest("AWS credentials are required")
		}
	case provider.ProviderGCP:
		if credentials.GCPProjectID == "" || credentials.GCPServiceAccountJSON == "" {
			return errors.BadRequest("GCP credentials are required")
		}
	case provider.ProviderAzure:
		if credentials.AzureTenantID == "" || credentials.AzureClientID == "" || credentials.AzureClientSecret == "" {
			return errors.BadRequest("Azure credentials are required")
		}
	default:
		return errors.BadRequest("Unsupported provider type")
	}

	s.logger.WithFields(map[string]interface{}{
		"provider": providerType,
	}).Info("Provider connection test passed")

	return nil
}

// Sync syncs resources from a provider
func (s *ProviderService) Sync(ctx context.Context, userID int64, providerType string) error {
	p, err := s.providerRepo.GetByProvider(ctx, userID, providerType)
	if err != nil {
		return err
	}

	if !p.IsConnected {
		return errors.BadRequest("Provider is not connected")
	}

	// Update last synced timestamp
	now := time.Now()
	err = s.providerRepo.UpdateSyncStatus(ctx, userID, providerType, now)
	if err != nil {
		s.logger.ErrorWithErr(err, "Failed to update sync status")
	}

	s.logger.WithFields(map[string]interface{}{
		"user_id":  userID,
		"provider": providerType,
	}).Info("Provider sync initiated")

	// Actual sync would happen here using the cloud provider SDKs
	return errors.Internal("Provider sync not yet fully implemented", nil)
}

// GetSyncStatus gets the sync status for all providers
func (s *ProviderService) GetSyncStatus(ctx context.Context, userID int64) ([]*provider.SyncStatus, error) {
	providers, err := s.providerRepo.List(ctx, userID)
	if err != nil {
		return nil, err
	}

	statuses := make([]*provider.SyncStatus, len(providers))
	for i, p := range providers {
		// Get resource count for this provider
		resources, err := s.resourceRepo.ListByProvider(ctx, userID, p.Provider)
		resourceCount := 0
		if err == nil {
			resourceCount = len(resources)
		}

		status := "connected"
		if !p.IsConnected {
			status = "disconnected"
		}

		statuses[i] = &provider.SyncStatus{
			Provider:      p.Provider,
			IsConnected:   p.IsConnected,
			LastSynced:    p.LastSynced,
			ResourceCount: resourceCount,
			Status:        status,
		}
	}

	return statuses, nil
}
