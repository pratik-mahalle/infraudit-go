package integration

import (
	"context"
	"testing"

	"github.com/pratik-mahalle/infraudit/internal/domain/provider"
	"github.com/pratik-mahalle/infraudit/internal/domain/resource"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
	"github.com/pratik-mahalle/infraudit/internal/repository/postgres"
	"github.com/pratik-mahalle/infraudit/internal/services"
	"github.com/pratik-mahalle/infraudit/internal/testutil"
)

func TestProviderSync(t *testing.T) {
	// Setup DB
	db := testutil.NewTestDB(t)
	defer testutil.CleanupDB(db)

	logger := logger.New(logger.Config{Level: "error", Format: "console"})
	providerRepo := postgres.NewProviderRepository(db)
	resourceRepo := postgres.NewResourceRepository(db)

	// Setup Service
	svc := services.NewProviderService(providerRepo, resourceRepo, logger)
	providerSvc, ok := svc.(*services.ProviderService)
	if !ok {
		t.Fatal("Failed to cast provider service")
	}

	ctx := context.Background()
	userID := int64(1)

	// Create User (referential integrity)
	_, err := db.Exec("INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)", userID, "test@example.com", "hash")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// 1. Test Sync with AWS
	t.Run("Sync AWS Success", func(t *testing.T) {
		// Create Connection
		creds := provider.Credentials{
			AWSAccessKeyID:     "AKIA...",
			AWSSecretAccessKey: "secret...",
			AWSRegion:          "us-east-1",
		}
		// Upsert provider manually or via Connect service (Connect tests connection, so lets insert manually)
		p := &provider.Provider{
			UserID:      userID,
			Provider:    provider.ProviderAWS,
			IsConnected: true,
			Credentials: creds,
		}
		err := providerRepo.Upsert(ctx, p)
		if err != nil {
			t.Fatalf("Failed to setup provider: %v", err)
		}

		// Setup Mock
		mockRes := []resource.Resource{
			{
				ResourceID: "i-1234567890abcdef0",
				Name:       "test-instance",
				Type:       resource.TypeEC2Instance,
				Region:     "us-east-1",
				Status:     resource.StatusActive,
			},
		}
		mockClient := &MockCloudProviderClient{
			AWSResources: mockRes,
		}
		providerSvc.SetClient(mockClient)

		// Execute Sync
		err = providerSvc.Sync(ctx, userID, provider.ProviderAWS)
		if err != nil {
			t.Fatalf("Sync failed: %v", err)
		}

		// Verify CloudProvider call
		if !mockClient.AWSListCalled {
			t.Error("Expected AWSListResources to be called")
		}

		// Verify DB
		savedResources, err := resourceRepo.ListByProvider(ctx, userID, provider.ProviderAWS)
		if err != nil {
			t.Fatalf("Failed to list resources: %v", err)
		}

		if len(savedResources) != 1 {
			t.Errorf("Expected 1 resource, got %d", len(savedResources))
		}
		if savedResources[0].ResourceID != "i-1234567890abcdef0" {
			t.Errorf("Expected resource ID i-1234567890abcdef0, got %s", savedResources[0].ResourceID)
		}

		// Verify Sync Status Updated
		statuses, err := providerSvc.GetSyncStatus(ctx, userID)
		if err != nil {
			t.Fatalf("Failed to get sync status: %v", err)
		}
		if len(statuses) != 1 {
			t.Fatalf("Expected 1 status, got %d", len(statuses))
		}
		if statuses[0].LastSynced.IsZero() {
			t.Error("LastSynced should be updated")
		}
	})

	// 2. Test Sync Failure handles gracefully
	t.Run("Sync Failure", func(t *testing.T) {
		// Mock Error
		mockClient := &MockCloudProviderClient{
			Err: context.DeadlineExceeded,
		}
		providerSvc.SetClient(mockClient)

		err := providerSvc.Sync(ctx, userID, provider.ProviderAWS)
		if err == nil {
			t.Error("Expected error from Sync")
		}
	})
}
