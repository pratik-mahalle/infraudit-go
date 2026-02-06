package dto

import "time"

// ProviderDTO represents a cloud provider account in API responses
type ProviderDTO struct {
	Provider    string     `json:"provider"`
	IsConnected bool       `json:"isConnected"`
	LastSynced  *time.Time `json:"lastSynced,omitempty"`
}

// ConnectProviderRequest represents a provider connection request
// Field names match the frontend ProviderCredentials interface
type ConnectProviderRequest struct {
	Provider string `json:"provider" validate:"required,oneof=aws gcp azure"`

	// AWS credentials
	AWSAccessKeyID     *string `json:"accessKeyId,omitempty"`
	AWSSecretAccessKey *string `json:"secretAccessKey,omitempty"`
	AWSRegion          *string `json:"region,omitempty"`

	// GCP credentials
	GCPProjectID          *string `json:"projectId,omitempty"`
	GCPServiceAccountJSON *string `json:"credentials,omitempty"`

	// Azure credentials
	AzureTenantID       *string `json:"tenantId,omitempty"`
	AzureClientID       *string `json:"clientId,omitempty"`
	AzureClientSecret   *string `json:"clientSecret,omitempty"`
	AzureSubscriptionID *string `json:"subscriptionId,omitempty"`
}

// SyncProviderRequest represents a provider sync request
type SyncProviderRequest struct {
	Provider string `json:"provider" validate:"required,oneof=aws gcp azure"`
}

// ProviderStatusResponse represents provider status information
type ProviderStatusResponse struct {
	Provider      string     `json:"provider"`
	IsConnected   bool       `json:"isConnected"`
	LastSynced    *time.Time `json:"lastSynced,omitempty"`
	ResourceCount int        `json:"resourceCount"`
	Status        string     `json:"status"`
	Message       string     `json:"message,omitempty"`
}
