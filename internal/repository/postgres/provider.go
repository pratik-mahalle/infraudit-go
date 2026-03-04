package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/pratik-mahalle/infraudit/internal/domain/provider"
	"github.com/pratik-mahalle/infraudit/internal/pkg/errors"
)

// ProviderRepository implements provider.Repository
type ProviderRepository struct {
	db *sql.DB
}

// NewProviderRepository creates a new provider repository
func NewProviderRepository(db *sql.DB) provider.Repository {
	return &ProviderRepository{db: db}
}

// Upsert creates or updates a provider account
func (r *ProviderRepository) Upsert(ctx context.Context, p *provider.Provider) error {
	now := time.Now()
	p.UpdatedAt = now

	query := `
		INSERT INTO provider_accounts (
			user_id, provider, is_connected, last_synced,
			aws_access_key_id, aws_secret_access_key, aws_region,
			gcp_project_id, gcp_service_account_json, gcp_region,
			azure_tenant_id, azure_client_id, azure_client_secret, azure_subscription_id, azure_location
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT(user_id, provider) DO UPDATE SET
			is_connected = excluded.is_connected,
			last_synced = excluded.last_synced,
			aws_access_key_id = excluded.aws_access_key_id,
			aws_secret_access_key = excluded.aws_secret_access_key,
			aws_region = excluded.aws_region,
			gcp_project_id = excluded.gcp_project_id,
			gcp_service_account_json = excluded.gcp_service_account_json,
			gcp_region = excluded.gcp_region,
			azure_tenant_id = excluded.azure_tenant_id,
			azure_client_id = excluded.azure_client_id,
			azure_client_secret = excluded.azure_client_secret,
			azure_subscription_id = excluded.azure_subscription_id,
			azure_location = excluded.azure_location
	`

	_, err := r.db.ExecContext(ctx, query,
		p.UserID, p.Provider, p.IsConnected, p.LastSynced,
		p.Credentials.AWSAccessKeyID, p.Credentials.AWSSecretAccessKey, p.Credentials.AWSRegion,
		p.Credentials.GCPProjectID, p.Credentials.GCPServiceAccountJSON, p.Credentials.GCPRegion,
		p.Credentials.AzureTenantID, p.Credentials.AzureClientID, p.Credentials.AzureClientSecret,
		p.Credentials.AzureSubscriptionID, p.Credentials.AzureLocation,
	)
	if err != nil {
		return errors.DatabaseError("Failed to upsert provider", err)
	}

	return nil
}

// GetByProvider retrieves a provider account by provider type
func (r *ProviderRepository) GetByProvider(ctx context.Context, userID int64, providerType string) (*provider.Provider, error) {
	query := `
		SELECT user_id, provider, is_connected, last_synced,
			aws_access_key_id, aws_secret_access_key, aws_region,
			gcp_project_id, gcp_service_account_json, gcp_region,
			azure_tenant_id, azure_client_id, azure_client_secret, azure_subscription_id, azure_location
		FROM provider_accounts
		WHERE user_id = $1 AND provider = $2
	`

	var p provider.Provider
	var creds provider.Credentials

	err := r.db.QueryRowContext(ctx, query, userID, providerType).Scan(
		&p.UserID, &p.Provider, &p.IsConnected, &p.LastSynced,
		&creds.AWSAccessKeyID, &creds.AWSSecretAccessKey, &creds.AWSRegion,
		&creds.GCPProjectID, &creds.GCPServiceAccountJSON, &creds.GCPRegion,
		&creds.AzureTenantID, &creds.AzureClientID, &creds.AzureClientSecret,
		&creds.AzureSubscriptionID, &creds.AzureLocation,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Provider")
	}
	if err != nil {
		return nil, errors.DatabaseError("Failed to get provider", err)
	}

	p.Credentials = creds

	return &p, nil
}

// List retrieves all provider accounts for a user
func (r *ProviderRepository) List(ctx context.Context, userID int64) ([]*provider.Provider, error) {
	query := `
		SELECT user_id, provider, is_connected, last_synced,
			aws_access_key_id, aws_secret_access_key, aws_region,
			gcp_project_id, gcp_service_account_json, gcp_region,
			azure_tenant_id, azure_client_id, azure_client_secret, azure_subscription_id, azure_location
		FROM provider_accounts
		WHERE user_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.DatabaseError("Failed to list providers", err)
	}
	defer rows.Close()

	var providers []*provider.Provider
	for rows.Next() {
		var p provider.Provider
		var creds provider.Credentials

		err := rows.Scan(
			&p.UserID, &p.Provider, &p.IsConnected, &p.LastSynced,
			&creds.AWSAccessKeyID, &creds.AWSSecretAccessKey, &creds.AWSRegion,
			&creds.GCPProjectID, &creds.GCPServiceAccountJSON, &creds.GCPRegion,
			&creds.AzureTenantID, &creds.AzureClientID, &creds.AzureClientSecret,
			&creds.AzureSubscriptionID, &creds.AzureLocation,
		)
		if err != nil {
			return nil, errors.DatabaseError("Failed to scan provider", err)
		}

		p.Credentials = creds

		providers = append(providers, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.DatabaseError("Failed to iterate providers", err)
	}

	return providers, nil
}

// Delete deletes a provider account
func (r *ProviderRepository) Delete(ctx context.Context, userID int64, providerType string) error {
	query := `DELETE FROM provider_accounts WHERE user_id = $1 AND provider = $2`

	result, err := r.db.ExecContext(ctx, query, userID, providerType)
	if err != nil {
		return errors.DatabaseError("Failed to delete provider", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.DatabaseError("Failed to get affected rows", err)
	}

	if rows == 0 {
		return errors.NotFound("Provider")
	}

	return nil
}

// UpdateSyncStatus updates the sync status
func (r *ProviderRepository) UpdateSyncStatus(ctx context.Context, userID int64, providerType string, lastSynced time.Time) error {
	query := `UPDATE provider_accounts SET last_synced = $1 WHERE user_id = $2 AND provider = $3`

	result, err := r.db.ExecContext(ctx, query, lastSynced, userID, providerType)
	if err != nil {
		return errors.DatabaseError("Failed to update sync status", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.DatabaseError("Failed to get affected rows", err)
	}

	if rows == 0 {
		return errors.NotFound("Provider")
	}

	return nil
}

// UpdateConnectionStatus updates the connection status
func (r *ProviderRepository) UpdateConnectionStatus(ctx context.Context, userID int64, providerType string, isConnected bool) error {
	query := `UPDATE provider_accounts SET is_connected = $1 WHERE user_id = $2 AND provider = $3`

	result, err := r.db.ExecContext(ctx, query, isConnected, userID, providerType)
	if err != nil {
		return errors.DatabaseError("Failed to update connection status", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.DatabaseError("Failed to get affected rows", err)
	}

	if rows == 0 {
		return errors.NotFound("Provider")
	}

	return nil
}
