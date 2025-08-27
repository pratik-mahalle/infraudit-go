package db

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct {
	sql *sql.DB
}

func Open(path string) (*DB, error) {
	d, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err := d.Exec(`PRAGMA journal_mode=WAL;`); err != nil {
		return nil, err
	}
	repo := &DB{sql: d}
	if err := repo.migrate(); err != nil {
		_ = d.Close()
		return nil, err
	}
	return repo, nil
}

func (d *DB) Close() error { return d.sql.Close() }

func (d *DB) migrate() error {
	_, err := d.sql.Exec(`
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY,
    email TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS provider_accounts (
    user_id INTEGER NOT NULL,
    provider TEXT NOT NULL,
    is_connected INTEGER NOT NULL DEFAULT 0,
    last_synced INTEGER,

    aws_access_key_id TEXT,
    aws_secret_access_key TEXT,
    aws_region TEXT,

    gcp_project_id TEXT,
    gcp_service_account_json TEXT,
    gcp_region TEXT,

    azure_tenant_id TEXT,
    azure_client_id TEXT,
    azure_client_secret TEXT,
    azure_subscription_id TEXT,
    azure_location TEXT,

    PRIMARY KEY (user_id, provider)
);

CREATE TABLE IF NOT EXISTS resources (
    user_id INTEGER NOT NULL,
    provider TEXT NOT NULL,
    resource_id TEXT NOT NULL,
    name TEXT,
    type TEXT,
    region TEXT,
    status TEXT,
    PRIMARY KEY (user_id, provider, resource_id)
);
`)
	return err
}

type ProviderAccountRow struct {
	UserID              int64
	Provider            string
	IsConnected         bool
	LastSynced          sqlNullTime
	AWSAccessKeyID      string
	AWSSecretAccessKey  string
	AWSRegion           string
	GCPProjectID        string
	GCPServiceAccount   string
	GCPRegion           string
	AzureTenantID       string
	AzureClientID       string
	AzureClientSecret   string
	AzureSubscriptionID string
	AzureLocation       string
}

type sqlNullTime struct {
	Valid bool
	Time  time.Time
}

func (n *sqlNullTime) Scan(v interface{}) error {
	switch t := v.(type) {
	case int64:
		if t == 0 {
			n.Valid = false
			return nil
		}
		n.Valid = true
		n.Time = time.Unix(t, 0)
		return nil
	case nil:
		n.Valid = false
		return nil
	default:
		return errors.New("invalid time")
	}
}

func (d *DB) UpsertProviderAccount(ctx context.Context, p ProviderAccountRow) error {
	_, err := d.sql.ExecContext(ctx, `
INSERT INTO provider_accounts (
  user_id, provider, is_connected, last_synced,
  aws_access_key_id, aws_secret_access_key, aws_region,
  gcp_project_id, gcp_service_account_json, gcp_region,
  azure_tenant_id, azure_client_id, azure_client_secret, azure_subscription_id, azure_location
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(user_id, provider) DO UPDATE SET
  is_connected=excluded.is_connected,
  last_synced=excluded.last_synced,
  aws_access_key_id=excluded.aws_access_key_id,
  aws_secret_access_key=excluded.aws_secret_access_key,
  aws_region=excluded.aws_region,
  gcp_project_id=excluded.gcp_project_id,
  gcp_service_account_json=excluded.gcp_service_account_json,
  gcp_region=excluded.gcp_region,
  azure_tenant_id=excluded.azure_tenant_id,
  azure_client_id=excluded.azure_client_id,
  azure_client_secret=excluded.azure_client_secret,
  azure_subscription_id=excluded.azure_subscription_id,
  azure_location=excluded.azure_location
`,
		p.UserID, p.Provider, boolToInt(p.IsConnected), nullableUnix(p.LastSynced),
		p.AWSAccessKeyID, p.AWSSecretAccessKey, p.AWSRegion,
		p.GCPProjectID, p.GCPServiceAccount, p.GCPRegion,
		p.AzureTenantID, p.AzureClientID, p.AzureClientSecret, p.AzureSubscriptionID, p.AzureLocation,
	)
	return err
}

func (d *DB) GetProviderAccounts(ctx context.Context, userID int64) ([]ProviderAccountRow, error) {
	rows, err := d.sql.QueryContext(ctx, `
SELECT user_id, provider, is_connected, last_synced,
  aws_access_key_id, aws_secret_access_key, aws_region,
  gcp_project_id, gcp_service_account_json, gcp_region,
  azure_tenant_id, azure_client_id, azure_client_secret, azure_subscription_id, azure_location
FROM provider_accounts WHERE user_id=?
`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ProviderAccountRow
	for rows.Next() {
		var p ProviderAccountRow
		var ts sqlNullTime
		var isConn int
		if err := rows.Scan(&p.UserID, &p.Provider, &isConn, &ts,
			&p.AWSAccessKeyID, &p.AWSSecretAccessKey, &p.AWSRegion,
			&p.GCPProjectID, &p.GCPServiceAccount, &p.GCPRegion,
			&p.AzureTenantID, &p.AzureClientID, &p.AzureClientSecret, &p.AzureSubscriptionID, &p.AzureLocation,
		); err != nil {
			return nil, err
		}
		p.IsConnected = isConn == 1
		p.LastSynced = ts
		out = append(out, p)
	}
	return out, rows.Err()
}

func (d *DB) DeleteProviderAccount(ctx context.Context, userID int64, provider string) error {
	_, err := d.sql.ExecContext(ctx, `DELETE FROM provider_accounts WHERE user_id=? AND provider=?`, userID, provider)
	return err
}

func (d *DB) SaveResources(ctx context.Context, userID int64, provider string, resources []ResourceRow) error {
	tx, err := d.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `DELETE FROM resources WHERE user_id=? AND provider=?`, userID, provider); err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO resources (user_id, provider, resource_id, name, type, region, status) VALUES (?,?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, r := range resources {
		if _, err := stmt.Exec(userID, provider, r.ResourceID, r.Name, r.Type, r.Region, r.Status); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (d *DB) ListResources(ctx context.Context, userID int64) ([]ResourceRow, error) {
	rows, err := d.sql.QueryContext(ctx, `SELECT provider, resource_id, name, type, region, status FROM resources WHERE user_id=?`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ResourceRow
	for rows.Next() {
		var r ResourceRow
		if err := rows.Scan(&r.Provider, &r.ResourceID, &r.Name, &r.Type, &r.Region, &r.Status); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ListResourcesPage returns a page of resources with optional provider filter and total count for pagination.
func (d *DB) ListResourcesPage(ctx context.Context, userID int64, provider string, limit, offset int) ([]ResourceRow, int, error) {
	// Count total
	var count int
	if provider == "" {
		if err := d.sql.QueryRowContext(ctx, `SELECT COUNT(1) FROM resources WHERE user_id=?`, userID).Scan(&count); err != nil {
			return nil, 0, err
		}
	} else {
		if err := d.sql.QueryRowContext(ctx, `SELECT COUNT(1) FROM resources WHERE user_id=? AND provider=?`, userID, provider).Scan(&count); err != nil {
			return nil, 0, err
		}
	}

	// Page query
	var rows *sql.Rows
	var err error
	if provider == "" {
		rows, err = d.sql.QueryContext(ctx, `SELECT provider, resource_id, name, type, region, status FROM resources WHERE user_id=? ORDER BY provider, resource_id LIMIT ? OFFSET ?`, userID, limit, offset)
	} else {
		rows, err = d.sql.QueryContext(ctx, `SELECT provider, resource_id, name, type, region, status FROM resources WHERE user_id=? AND provider=? ORDER BY resource_id LIMIT ? OFFSET ?`, userID, provider, limit, offset)
	}
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []ResourceRow
	for rows.Next() {
		var r ResourceRow
		if err := rows.Scan(&r.Provider, &r.ResourceID, &r.Name, &r.Type, &r.Region, &r.Status); err != nil {
			return nil, 0, err
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return out, count, nil
}

type ResourceRow struct {
	Provider   string
	ResourceID string
	Name       string
	Type       string
	Region     string
	Status     string
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
func nullableUnix(t sqlNullTime) any {
	if !t.Valid {
		return nil
	}
	return t.Time.Unix()
}
