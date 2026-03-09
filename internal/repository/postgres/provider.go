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

	// Compute the region based on provider type (DB uses single generic column)
	var regionValue string
	switch p.Provider {
	case "aws":
		regionValue = p.Credentials.AWSRegion
	case "gcp":
		regionValue = p.Credentials.GCPRegion
	case "azure":
		regionValue = p.Credentials.AzureLocation
	}

	query := `
		INSERT INTO provider_accounts (
			user_id, provider, is_connected, last_synced,
			access_key_id, secret_access_key, region,
			project_id, service_account_json,
			tenant_id, client_id, client_secret, subscription_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT(user_id, provider) DO UPDATE SET
			is_connected = excluded.is_connected,
			last_synced = excluded.last_synced,
			access_key_id = excluded.access_key_id,
			secret_access_key = excluded.secret_access_key,
			region = excluded.region,
			project_id = excluded.project_id,
			service_account_json = excluded.service_account_json,
			tenant_id = excluded.tenant_id,
			client_id = excluded.client_id,
			client_secret = excluded.client_secret,
			subscription_id = excluded.subscription_id
	`

	_, err := r.db.ExecContext(ctx, query,
		p.UserID, p.Provider, p.IsConnected, p.LastSynced,
		p.Credentials.AWSAccessKeyID, p.Credentials.AWSSecretAccessKey, regionValue,
		p.Credentials.GCPProjectID, p.Credentials.GCPServiceAccountJSON,
		p.Credentials.AzureTenantID, p.Credentials.AzureClientID, p.Credentials.AzureClientSecret,
		p.Credentials.AzureSubscriptionID,
	)
	if err != nil {
		return errors.DatabaseError("Failed to upsert provider", err)
	}

	return nil
}

// scanProvider scans a provider row from the DB generic columns into the Go struct
func scanProvider(scan func(dest ...any) error) (*provider.Provider, error) {
	var p provider.Provider
	var creds provider.Credentials
	var region string

	err := scan(
		&p.UserID, &p.Provider, &p.IsConnected, &p.LastSynced,
		&creds.AWSAccessKeyID, &creds.AWSSecretAccessKey, &region,
		&creds.GCPProjectID, &creds.GCPServiceAccountJSON,
		&creds.AzureTenantID, &creds.AzureClientID, &creds.AzureClientSecret,
		&creds.AzureSubscriptionID,
	)
	if err != nil {
		return nil, err
	}

	// Map the generic region column to the provider-specific field
	switch p.Provider {
	case "aws":
		creds.AWSRegion = region
	case "gcp":
		creds.GCPRegion = region
	case "azure":
		creds.AzureLocation = region
	}

	p.Credentials = creds
	return &p, nil
}

const providerSelectCols = `user_id, provider, is_connected, last_synced,
	access_key_id, secret_access_key, region,
	project_id, service_account_json,
	tenant_id, client_id, client_secret, subscription_id`

// GetByProvider retrieves a provider account by provider type
func (r *ProviderRepository) GetByProvider(ctx context.Context, userID int64, providerType string) (*provider.Provider, error) {
	query := `SELECT ` + providerSelectCols + ` FROM provider_accounts WHERE user_id = $1 AND provider = $2`

	row := r.db.QueryRowContext(ctx, query, userID, providerType)
	p, err := scanProvider(row.Scan)

	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Provider")
	}
	if err != nil {
		return nil, errors.DatabaseError("Failed to get provider", err)
	}

	return p, nil
}

// List retrieves all provider accounts for a user
func (r *ProviderRepository) List(ctx context.Context, userID int64) ([]*provider.Provider, error) {
	query := `SELECT ` + providerSelectCols + ` FROM provider_accounts WHERE user_id = $1`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.DatabaseError("Failed to list providers", err)
	}
	defer rows.Close()

	var providers []*provider.Provider
	for rows.Next() {
		p, err := scanProvider(rows.Scan)
		if err != nil {
			return nil, errors.DatabaseError("Failed to scan provider", err)
		}
		providers = append(providers, p)
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
