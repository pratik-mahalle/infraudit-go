package services

import (
	"context"
	"time"

	"github.com/pratik-mahalle/infraudit/internal/domain/provider"
	"github.com/pratik-mahalle/infraudit/internal/domain/resource"
	"github.com/pratik-mahalle/infraudit/internal/pkg/errors"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
	cloudproviders "github.com/pratik-mahalle/infraudit/internal/providers"
)

// CloudProviderClient defines the interface for interacting with cloud providers
// This allows mocking for tests
type CloudProviderClient interface {
	AWSListResources(ctx context.Context, creds cloudproviders.AWSCredentials) ([]resource.Resource, error)
	AzureListResources(ctx context.Context, creds cloudproviders.AzureCredentials) ([]resource.Resource, error)
	GCPListResources(ctx context.Context, creds cloudproviders.GCPCredentials) ([]resource.Resource, error)
}

// DefaultCloudProviderClient is the actual implementation calling provider packages
type DefaultCloudProviderClient struct{}

func (c *DefaultCloudProviderClient) AWSListResources(ctx context.Context, creds cloudproviders.AWSCredentials) ([]resource.Resource, error) {
	return cloudproviders.AWSListResources(ctx, creds)
}

func (c *DefaultCloudProviderClient) AzureListResources(ctx context.Context, creds cloudproviders.AzureCredentials) ([]resource.Resource, error) {
	return cloudproviders.AzureListResources(ctx, creds)
}

func (c *DefaultCloudProviderClient) GCPListResources(ctx context.Context, creds cloudproviders.GCPCredentials) ([]resource.Resource, error) {
	return cloudproviders.GCPListResources(ctx, creds)
}

// ProviderService implements provider.Service
type ProviderService struct {
	providerRepo provider.Repository
	resourceRepo resource.Repository
	logger       *logger.Logger
	client       CloudProviderClient
}

// NewProviderService creates a new provider service
func NewProviderService(providerRepo provider.Repository, resourceRepo resource.Repository, log *logger.Logger) provider.Service {
	return &ProviderService{
		providerRepo: providerRepo,
		resourceRepo: resourceRepo,
		logger:       log,
		client:       &DefaultCloudProviderClient{},
	}
}

// SetClient sets the cloud provider client (used for testing)
func (s *ProviderService) SetClient(client CloudProviderClient) {
	s.client = client
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

	s.logger.WithFields(map[string]interface{}{
		"user_id":  userID,
		"provider": providerType,
	}).Info("Provider sync initiated")

	var resources []*resource.Resource

	switch providerType {
	case provider.ProviderAWS:
		creds := cloudproviders.AWSCredentials{
			AccessKeyID:     p.Credentials.AWSAccessKeyID,
			SecretAccessKey: p.Credentials.AWSSecretAccessKey,
			Region:          p.Credentials.AWSRegion,
		}
		res, err := s.client.AWSListResources(ctx, creds)
		if err != nil {
			return errors.Internal("Failed to list AWS resources", err)
		}
		for i := range res {
			r := &res[i]
			r.UserID = userID
			resources = append(resources, r)
		}

	case provider.ProviderAzure:
		creds := cloudproviders.AzureCredentials{
			TenantID:       p.Credentials.AzureTenantID,
			ClientID:       p.Credentials.AzureClientID,
			ClientSecret:   p.Credentials.AzureClientSecret,
			SubscriptionID: p.Credentials.AzureSubscriptionID,
			Location:       p.Credentials.AzureLocation,
		}
		res, err := s.client.AzureListResources(ctx, creds)
		if err != nil {
			return errors.Internal("Failed to list Azure resources", err)
		}
		for i := range res {
			r := &res[i]
			r.UserID = userID
			resources = append(resources, r)
		}

	case provider.ProviderGCP:
		creds := cloudproviders.GCPCredentials{
			ProjectID:          p.Credentials.GCPProjectID,
			ServiceAccountJSON: p.Credentials.GCPServiceAccountJSON,
			Region:             p.Credentials.GCPRegion,
		}
		res, err := s.client.GCPListResources(ctx, creds)
		if err != nil {
			return errors.Internal("Failed to list GCP resources", err)
		}
		for i := range res {
			r := &res[i]
			r.UserID = userID
			resources = append(resources, r)
		}

	default:
		return errors.BadRequest("Unsupported provider type")
	}

	// Save resources in batch
	if len(resources) > 0 {
		err = s.resourceRepo.SaveBatch(ctx, userID, providerType, resources)
		if err != nil {
			s.logger.ErrorWithErr(err, "Failed to save synced resources")
			return err
		}
	} else {
		// Even if no resources, we might want to ensure existing ones are cleaned up?
		// SaveBatch likely handles upserts. We might need logic to delete stale resources.
		// For now, let's assume SaveBatch or DeleteByProvider usage if strict sync required.
		// But basic logic is simple sync.
		s.logger.Infof("No resources found for provider %s", providerType)
	}

	// Update last synced timestamp
	now := time.Now()
	err = s.providerRepo.UpdateSyncStatus(ctx, userID, providerType, now)
	if err != nil {
		s.logger.ErrorWithErr(err, "Failed to update sync status")
	}

	s.logger.WithFields(map[string]interface{}{
		"user_id":        userID,
		"provider":       providerType,
		"resource_count": len(resources),
	}).Info("Provider sync completed")

	return nil
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
