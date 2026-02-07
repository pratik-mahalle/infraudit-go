package provider

import "time"

// Provider represents a cloud provider account
type Provider struct {
	ID           int64      `json:"id"`
	UserID       int64      `json:"user_id"`
	Provider     string     `json:"provider"`
	IsConnected  bool       `json:"is_connected"`
	LastSynced   *time.Time `json:"last_synced,omitempty"`
	Credentials  Credentials `json:"-"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// Credentials contains provider-specific credentials
type Credentials struct {
	// AWS
	AWSAccessKeyID     string `json:"aws_access_key_id,omitempty"`
	AWSSecretAccessKey string `json:"aws_secret_access_key,omitempty"`
	AWSRegion          string `json:"aws_region,omitempty"`

	// GCP
	GCPProjectID          string `json:"gcp_project_id,omitempty"`
	GCPServiceAccountJSON string `json:"gcp_service_account_json,omitempty"`
	GCPRegion             string `json:"gcp_region,omitempty"`
	GCPBillingDataset     string `json:"gcp_billing_dataset,omitempty"`
	GCPBillingAccountID   string `json:"gcp_billing_account_id,omitempty"`

	// Azure
	AzureTenantID       string `json:"azure_tenant_id,omitempty"`
	AzureClientID       string `json:"azure_client_id,omitempty"`
	AzureClientSecret   string `json:"azure_client_secret,omitempty"`
	AzureSubscriptionID string `json:"azure_subscription_id,omitempty"`
	AzureLocation       string `json:"azure_location,omitempty"`
}

// Provider types
const (
	ProviderAWS          = "aws"
	ProviderGCP          = "gcp"
	ProviderAzure        = "azure"
	ProviderDigitalOcean = "digitalocean"
)

// SyncStatus represents the sync status
type SyncStatus struct {
	Provider     string     `json:"provider"`
	IsConnected  bool       `json:"is_connected"`
	LastSynced   *time.Time `json:"last_synced,omitempty"`
	ResourceCount int       `json:"resource_count,omitempty"`
	Status       string     `json:"status"`
	Message      string     `json:"message,omitempty"`
}
