package services

import (
	"context"
	"testing"

	"github.com/pratik-mahalle/infraudit/internal/domain/provider"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
	"github.com/pratik-mahalle/infraudit/internal/testutil"
)

func TestProviderService_Connect(t *testing.T) {
	providerRepo := testutil.NewMockProviderRepository()
	resourceRepo := testutil.NewMockResourceRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := NewProviderService(providerRepo, resourceRepo, log)

	tests := []struct {
		name         string
		userID       int64
		providerType string
		credentials  provider.Credentials
		wantErr      bool
	}{
		{
			name:         "connect AWS with valid credentials",
			userID:       1,
			providerType: provider.ProviderAWS,
			credentials: provider.Credentials{
				AWSAccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
				AWSSecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			},
			wantErr: false,
		},
		{
			name:         "connect GCP with valid credentials",
			userID:       1,
			providerType: provider.ProviderGCP,
			credentials: provider.Credentials{
				GCPProjectID:          "my-gcp-project",
				GCPServiceAccountJSON: `{"type": "service_account"}`,
			},
			wantErr: false,
		},
		{
			name:         "connect Azure with valid credentials",
			userID:       1,
			providerType: provider.ProviderAzure,
			credentials: provider.Credentials{
				AzureTenantID:     "tenant-123",
				AzureClientID:     "client-456",
				AzureClientSecret: "secret-789",
			},
			wantErr: false,
		},
		{
			name:         "connect AWS with missing credentials",
			userID:       1,
			providerType: provider.ProviderAWS,
			credentials:  provider.Credentials{}, // Empty
			wantErr:      true,
		},
		{
			name:         "connect unsupported provider",
			userID:       1,
			providerType: "unsupported",
			credentials: provider.Credentials{
				AWSAccessKeyID: "test",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := service.Connect(ctx, tt.userID, tt.providerType, tt.credentials)

			if (err != nil) != tt.wantErr {
				t.Errorf("Connect() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProviderService_Disconnect(t *testing.T) {
	providerRepo := testutil.NewMockProviderRepository()
	resourceRepo := testutil.NewMockResourceRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := NewProviderService(providerRepo, resourceRepo, log)

	ctx := context.Background()

	// First connect a provider
	err := service.Connect(ctx, 1, provider.ProviderAWS, provider.Credentials{
		AWSAccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
		AWSSecretAccessKey: "secret",
	})
	if err != nil {
		t.Fatalf("Failed to connect provider: %v", err)
	}

	// Verify it's connected
	p, err := service.GetByProvider(ctx, 1, provider.ProviderAWS)
	if err != nil {
		t.Fatalf("Failed to get provider: %v", err)
	}
	if !p.IsConnected {
		t.Error("Provider should be connected")
	}

	// Disconnect
	err = service.Disconnect(ctx, 1, provider.ProviderAWS)
	if err != nil {
		t.Errorf("Disconnect() error = %v", err)
	}

	// Verify disconnection
	_, err = service.GetByProvider(ctx, 1, provider.ProviderAWS)
	if err == nil {
		t.Error("Provider should not exist after disconnect")
	}
}

func TestProviderService_List(t *testing.T) {
	providerRepo := testutil.NewMockProviderRepository()
	resourceRepo := testutil.NewMockResourceRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := NewProviderService(providerRepo, resourceRepo, log)

	ctx := context.Background()

	// Connect multiple providers for user 1
	service.Connect(ctx, 1, provider.ProviderAWS, provider.Credentials{
		AWSAccessKeyID:     "key1",
		AWSSecretAccessKey: "secret1",
	})
	service.Connect(ctx, 1, provider.ProviderGCP, provider.Credentials{
		GCPProjectID:          "project1",
		GCPServiceAccountJSON: "{}",
	})

	// Connect provider for user 2
	service.Connect(ctx, 2, provider.ProviderAzure, provider.Credentials{
		AzureTenantID:     "tenant",
		AzureClientID:     "client",
		AzureClientSecret: "secret",
	})

	// List providers for user 1
	providers, err := service.List(ctx, 1)
	if err != nil {
		t.Errorf("List() error = %v", err)
		return
	}

	if len(providers) != 2 {
		t.Errorf("List() returned %v providers, want %v", len(providers), 2)
	}

	// List providers for user 2
	providers2, err := service.List(ctx, 2)
	if err != nil {
		t.Errorf("List() error = %v", err)
		return
	}

	if len(providers2) != 1 {
		t.Errorf("List() for user 2 returned %v providers, want %v", len(providers2), 1)
	}
}

func TestProviderService_TestConnection(t *testing.T) {
	providerRepo := testutil.NewMockProviderRepository()
	resourceRepo := testutil.NewMockResourceRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := NewProviderService(providerRepo, resourceRepo, log)

	ctx := context.Background()

	tests := []struct {
		name         string
		providerType string
		credentials  provider.Credentials
		wantErr      bool
	}{
		{
			name:         "valid AWS credentials",
			providerType: provider.ProviderAWS,
			credentials: provider.Credentials{
				AWSAccessKeyID:     "key",
				AWSSecretAccessKey: "secret",
			},
			wantErr: false,
		},
		{
			name:         "missing AWS access key",
			providerType: provider.ProviderAWS,
			credentials: provider.Credentials{
				AWSSecretAccessKey: "secret",
			},
			wantErr: true,
		},
		{
			name:         "valid GCP credentials",
			providerType: provider.ProviderGCP,
			credentials: provider.Credentials{
				GCPProjectID:          "project",
				GCPServiceAccountJSON: "{}",
			},
			wantErr: false,
		},
		{
			name:         "missing GCP project",
			providerType: provider.ProviderGCP,
			credentials: provider.Credentials{
				GCPServiceAccountJSON: "{}",
			},
			wantErr: true,
		},
		{
			name:         "valid Azure credentials",
			providerType: provider.ProviderAzure,
			credentials: provider.Credentials{
				AzureTenantID:     "tenant",
				AzureClientID:     "client",
				AzureClientSecret: "secret",
			},
			wantErr: false,
		},
		{
			name:         "missing Azure tenant",
			providerType: provider.ProviderAzure,
			credentials: provider.Credentials{
				AzureClientID:     "client",
				AzureClientSecret: "secret",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.TestConnection(ctx, tt.providerType, tt.credentials)

			if (err != nil) != tt.wantErr {
				t.Errorf("TestConnection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProviderService_GetSyncStatus(t *testing.T) {
	providerRepo := testutil.NewMockProviderRepository()
	resourceRepo := testutil.NewMockResourceRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := NewProviderService(providerRepo, resourceRepo, log)

	ctx := context.Background()

	// Connect a provider
	service.Connect(ctx, 1, provider.ProviderAWS, provider.Credentials{
		AWSAccessKeyID:     "key",
		AWSSecretAccessKey: "secret",
	})

	// Get sync status
	statuses, err := service.GetSyncStatus(ctx, 1)
	if err != nil {
		t.Errorf("GetSyncStatus() error = %v", err)
		return
	}

	if len(statuses) != 1 {
		t.Errorf("GetSyncStatus() returned %v statuses, want %v", len(statuses), 1)
		return
	}

	status := statuses[0]
	if status.Provider != provider.ProviderAWS {
		t.Errorf("GetSyncStatus() provider = %v, want %v", status.Provider, provider.ProviderAWS)
	}
	if !status.IsConnected {
		t.Error("GetSyncStatus() provider should be connected")
	}
	if status.Status != "connected" {
		t.Errorf("GetSyncStatus() status = %v, want %v", status.Status, "connected")
	}
}
