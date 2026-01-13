package dto

import "time"

// ProviderDTO represents a cloud provider account in API responses
type ProviderDTO struct {
	Provider    string     `json:"provider"`
	IsConnected bool       `json:"is_connected"`
	LastSynced  *time.Time `json:"last_synced,omitempty"`
}

// ConnectProviderRequest represents a provider connection request
type ConnectProviderRequest struct {
	Provider string `json:"provider" validate:"required,oneof=aws gcp azure"`

	// AWS credentials
	AWSAccessKeyID     *string `json:"aws_access_key_id,omitempty"`
	AWSSecretAccessKey *string `json:"aws_secret_access_key,omitempty"`
	AWSRegion          *string `json:"aws_region,omitempty"`

	// GCP credentials
	GCPProjectID          *string `json:"gcp_project_id,omitempty"`
	GCPServiceAccountJSON *string `json:"gcp_service_account_json,omitempty"`
	GCPRegion             *string `json:"gcp_region,omitempty"`

	// Azure credentials
	AzureTenantID       *string `json:"azure_tenant_id,omitempty"`
	AzureClientID       *string `json:"azure_client_id,omitempty"`
	AzureClientSecret   *string `json:"azure_client_secret,omitempty"`
	AzureSubscriptionID *string `json:"azure_subscription_id,omitempty"`
	AzureLocation       *string `json:"azure_location,omitempty"`
}

// SyncProviderRequest represents a provider sync request
type SyncProviderRequest struct {
	Provider string `json:"provider" validate:"required,oneof=aws gcp azure"`
}

// ProviderStatusResponse represents provider status information
type ProviderStatusResponse struct {
	Provider      string     `json:"provider"`
	IsConnected   bool       `json:"is_connected"`
	LastSynced    *time.Time `json:"last_synced,omitempty"`
	ResourceCount int        `json:"resource_count"`
	Status        string     `json:"status"`
	Message       string     `json:"message,omitempty"`
}
