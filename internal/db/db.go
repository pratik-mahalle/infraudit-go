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
	if err := RunMigrations(d); err != nil {
		_ = d.Close()
		return nil, err
	}
	return repo, nil
}

func (d *DB) Close() error { return d.sql.Close() }

// Ping verifies a connection to the database is still alive.
func (d *DB) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return d.sql.PingContext(ctx)

}

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

CREATE TABLE IF NOT EXISTS alerts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    type TEXT,
    severity TEXT,
    title TEXT,
    description TEXT,
    resource TEXT,
    status TEXT,
    timestamp TEXT
);

CREATE TABLE IF NOT EXISTS recommendations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    type TEXT,
    priority TEXT,
    title TEXT,
    description TEXT,
    savings REAL,
    effort TEXT,
    impact TEXT,
    category TEXT,
    resources TEXT
);

CREATE TABLE IF NOT EXISTS security_drifts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    resource_id TEXT,
    drift_type TEXT,
    severity TEXT,
    details TEXT,
    detected_at TEXT,
    status TEXT
);

CREATE TABLE IF NOT EXISTS cost_anomalies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    resource_id TEXT,
    anomaly_type TEXT,
    severity TEXT,
    percentage INTEGER,
    previous_cost INTEGER,
    current_cost INTEGER,
    detected_at TEXT,
    status TEXT
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

// UpsertResource inserts or updates a single resource row.
func (d *DB) UpsertResource(ctx context.Context, userID int64, r ResourceRow) error {
	_, err := d.sql.ExecContext(ctx, `
INSERT INTO resources (user_id, provider, resource_id, name, type, region, status)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(user_id, provider, resource_id) DO UPDATE SET
  name=excluded.name,
  type=excluded.type,
  region=excluded.region,
  status=excluded.status
`, userID, r.Provider, r.ResourceID, r.Name, r.Type, r.Region, r.Status)
	return err
}

// GetResourceByID returns a resource by its resource_id for a user.
func (d *DB) GetResourceByID(ctx context.Context, userID int64, resourceID string) (ResourceRow, bool, error) {
	var r ResourceRow
	err := d.sql.QueryRowContext(ctx, `
SELECT provider, resource_id, name, type, region, status
FROM resources WHERE user_id=? AND resource_id=?
`, userID, resourceID).Scan(&r.Provider, &r.ResourceID, &r.Name, &r.Type, &r.Region, &r.Status)
	if err == sql.ErrNoRows {
		return ResourceRow{}, false, nil
	}
	if err != nil {
		return ResourceRow{}, false, err
	}
	return r, true, nil
}

// UpdateResource updates selected fields for a resource by id.
func (d *DB) UpdateResource(ctx context.Context, userID int64, resourceID string, name, typ, region, status *string) error {
	_, err := d.sql.ExecContext(ctx, `
UPDATE resources
SET name = COALESCE(?, name),
    type = COALESCE(?, type),
    region = COALESCE(?, region),
    status = COALESCE(?, status)
WHERE user_id=? AND resource_id=?
`, name, typ, region, status, userID, resourceID)
	return err
}

// DeleteResource removes a resource by id for a user.
func (d *DB) DeleteResource(ctx context.Context, userID int64, resourceID string) error {
	_, err := d.sql.ExecContext(ctx, `DELETE FROM resources WHERE user_id=? AND resource_id=?`, userID, resourceID)
	return err
}

// ----- Alerts -----

type AlertRow struct {
	ID          int64
	UserID      int64
	Type        string
	Severity    string
	Title       string
	Description string
	Resource    string
	Status      string
	Timestamp   string
}

func (d *DB) CreateAlert(ctx context.Context, a AlertRow) (int64, error) {
	res, err := d.sql.ExecContext(ctx, `INSERT INTO alerts (user_id, type, severity, title, description, resource, status, timestamp) VALUES (?,?,?,?,?,?,?,?)`,
		a.UserID, a.Type, a.Severity, a.Title, a.Description, a.Resource, a.Status, a.Timestamp)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *DB) GetAlertByID(ctx context.Context, userID int64, id int64) (AlertRow, bool, error) {
	var a AlertRow
	err := d.sql.QueryRowContext(ctx, `SELECT id, user_id, type, severity, title, description, resource, status, timestamp FROM alerts WHERE user_id=? AND id=?`, userID, id).
		Scan(&a.ID, &a.UserID, &a.Type, &a.Severity, &a.Title, &a.Description, &a.Resource, &a.Status, &a.Timestamp)
	if err == sql.ErrNoRows {
		return AlertRow{}, false, nil
	}
	if err != nil {
		return AlertRow{}, false, err
	}
	return a, true, nil
}

func (d *DB) ListAlerts(ctx context.Context, userID int64, typ, severity string) ([]AlertRow, error) {
	var rows *sql.Rows
	var err error
	if typ != "" && severity != "" {
		rows, err = d.sql.QueryContext(ctx, `SELECT id, user_id, type, severity, title, description, resource, status, timestamp FROM alerts WHERE user_id=? AND type=? AND severity=? ORDER BY id DESC`, userID, typ, severity)
	} else if typ != "" {
		rows, err = d.sql.QueryContext(ctx, `SELECT id, user_id, type, severity, title, description, resource, status, timestamp FROM alerts WHERE user_id=? AND type=? ORDER BY id DESC`, userID, typ)
	} else if severity != "" {
		rows, err = d.sql.QueryContext(ctx, `SELECT id, user_id, type, severity, title, description, resource, status, timestamp FROM alerts WHERE user_id=? AND severity=? ORDER BY id DESC`, userID, severity)
	} else {
		rows, err = d.sql.QueryContext(ctx, `SELECT id, user_id, type, severity, title, description, resource, status, timestamp FROM alerts WHERE user_id=? ORDER BY id DESC`, userID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AlertRow
	for rows.Next() {
		var a AlertRow
		if err := rows.Scan(&a.ID, &a.UserID, &a.Type, &a.Severity, &a.Title, &a.Description, &a.Resource, &a.Status, &a.Timestamp); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (d *DB) UpdateAlert(ctx context.Context, userID int64, id int64, typ, severity, title, description, resource, status, timestamp *string) error {
	_, err := d.sql.ExecContext(ctx, `UPDATE alerts SET type=COALESCE(?, type), severity=COALESCE(?, severity), title=COALESCE(?, title), description=COALESCE(?, description), resource=COALESCE(?, resource), status=COALESCE(?, status), timestamp=COALESCE(?, timestamp) WHERE user_id=? AND id=?`,
		typ, severity, title, description, resource, status, timestamp, userID, id)
	return err
}

func (d *DB) DeleteAlert(ctx context.Context, userID int64, id int64) error {
	_, err := d.sql.ExecContext(ctx, `DELETE FROM alerts WHERE user_id=? AND id=?`, userID, id)
	return err
}

// ----- Recommendations -----

type RecommendationRow struct {
	ID          int64
	UserID      int64
	Type        string
	Priority    string
	Title       string
	Description string
	Savings     float64
	Effort      string
	Impact      string
	Category    string
	Resources   string
}

func (d *DB) CreateRecommendation(ctx context.Context, r RecommendationRow) (int64, error) {
	res, err := d.sql.ExecContext(ctx, `INSERT INTO recommendations (user_id, type, priority, title, description, savings, effort, impact, category, resources) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		r.UserID, r.Type, r.Priority, r.Title, r.Description, r.Savings, r.Effort, r.Impact, r.Category, r.Resources)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *DB) GetRecommendationByID(ctx context.Context, userID int64, id int64) (RecommendationRow, bool, error) {
	var r RecommendationRow
	err := d.sql.QueryRowContext(ctx, `SELECT id, user_id, type, priority, title, description, savings, effort, impact, category, resources FROM recommendations WHERE user_id=? AND id=?`, userID, id).
		Scan(&r.ID, &r.UserID, &r.Type, &r.Priority, &r.Title, &r.Description, &r.Savings, &r.Effort, &r.Impact, &r.Category, &r.Resources)
	if err == sql.ErrNoRows {
		return RecommendationRow{}, false, nil
	}
	if err != nil {
		return RecommendationRow{}, false, err
	}
	return r, true, nil
}

func (d *DB) ListRecommendations(ctx context.Context, userID int64, typ string) ([]RecommendationRow, error) {
	var rows *sql.Rows
	var err error
	if typ != "" {
		rows, err = d.sql.QueryContext(ctx, `SELECT id, user_id, type, priority, title, description, savings, effort, impact, category, resources FROM recommendations WHERE user_id=? AND type=? ORDER BY id DESC`, userID, typ)
	} else {
		rows, err = d.sql.QueryContext(ctx, `SELECT id, user_id, type, priority, title, description, savings, effort, impact, category, resources FROM recommendations WHERE user_id=? ORDER BY id DESC`, userID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []RecommendationRow
	for rows.Next() {
		var r RecommendationRow
		if err := rows.Scan(&r.ID, &r.UserID, &r.Type, &r.Priority, &r.Title, &r.Description, &r.Savings, &r.Effort, &r.Impact, &r.Category, &r.Resources); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (d *DB) UpdateRecommendation(ctx context.Context, userID int64, id int64, typ, priority, title, description, effort, impact, category, resources *string, savings *float64) error {
	_, err := d.sql.ExecContext(ctx, `UPDATE recommendations SET type=COALESCE(?, type), priority=COALESCE(?, priority), title=COALESCE(?, title), description=COALESCE(?, description), effort=COALESCE(?, effort), impact=COALESCE(?, impact), category=COALESCE(?, category), resources=COALESCE(?, resources), savings=COALESCE(?, savings) WHERE user_id=? AND id=?`,
		typ, priority, title, description, effort, impact, category, resources, savings, userID, id)
	return err
}

func (d *DB) DeleteRecommendation(ctx context.Context, userID int64, id int64) error {
	_, err := d.sql.ExecContext(ctx, `DELETE FROM recommendations WHERE user_id=? AND id=?`, userID, id)
	return err
}

// ----- Security Drifts -----

type SecurityDriftRow struct {
	ID         int64
	UserID     int64
	ResourceID string
	DriftType  string
	Severity   string
	Details    string
	DetectedAt string
	Status     string
}

func (d *DB) CreateSecurityDrift(ctx context.Context, s SecurityDriftRow) (int64, error) {
	res, err := d.sql.ExecContext(ctx, `INSERT INTO security_drifts (user_id, resource_id, drift_type, severity, details, detected_at, status) VALUES (?,?,?,?,?,?,?)`,
		s.UserID, s.ResourceID, s.DriftType, s.Severity, s.Details, s.DetectedAt, s.Status)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *DB) GetSecurityDriftByID(ctx context.Context, userID int64, id int64) (SecurityDriftRow, bool, error) {
	var s SecurityDriftRow
	err := d.sql.QueryRowContext(ctx, `SELECT id, user_id, resource_id, drift_type, severity, details, detected_at, status FROM security_drifts WHERE user_id=? AND id=?`, userID, id).
		Scan(&s.ID, &s.UserID, &s.ResourceID, &s.DriftType, &s.Severity, &s.Details, &s.DetectedAt, &s.Status)
	if err == sql.ErrNoRows {
		return SecurityDriftRow{}, false, nil
	}
	if err != nil {
		return SecurityDriftRow{}, false, err
	}
	return s, true, nil
}

func (d *DB) ListSecurityDrifts(ctx context.Context, userID int64, resourceID string, severity string) ([]SecurityDriftRow, error) {
	var rows *sql.Rows
	var err error
	if resourceID != "" {
		rows, err = d.sql.QueryContext(ctx, `SELECT id, user_id, resource_id, drift_type, severity, details, detected_at, status FROM security_drifts WHERE user_id=? AND resource_id=? ORDER BY id DESC`, userID, resourceID)
	} else if severity != "" {
		rows, err = d.sql.QueryContext(ctx, `SELECT id, user_id, resource_id, drift_type, severity, details, detected_at, status FROM security_drifts WHERE user_id=? AND severity=? ORDER BY id DESC`, userID, severity)
	} else {
		rows, err = d.sql.QueryContext(ctx, `SELECT id, user_id, resource_id, drift_type, severity, details, detected_at, status FROM security_drifts WHERE user_id=? ORDER BY id DESC`, userID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []SecurityDriftRow
	for rows.Next() {
		var s SecurityDriftRow
		if err := rows.Scan(&s.ID, &s.UserID, &s.ResourceID, &s.DriftType, &s.Severity, &s.Details, &s.DetectedAt, &s.Status); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (d *DB) UpdateSecurityDrift(ctx context.Context, userID int64, id int64, resourceID, driftType, severity, details, detectedAt, status *string) error {
	_, err := d.sql.ExecContext(ctx, `UPDATE security_drifts SET resource_id=COALESCE(?, resource_id), drift_type=COALESCE(?, drift_type), severity=COALESCE(?, severity), details=COALESCE(?, details), detected_at=COALESCE(?, detected_at), status=COALESCE(?, status) WHERE user_id=? AND id=?`,
		resourceID, driftType, severity, details, detectedAt, status, userID, id)
	return err
}

func (d *DB) DeleteSecurityDrift(ctx context.Context, userID int64, id int64) error {
	_, err := d.sql.ExecContext(ctx, `DELETE FROM security_drifts WHERE user_id=? AND id=?`, userID, id)
	return err
}

// ----- Cost Anomalies -----

type CostAnomalyRow struct {
	ID           int64
	UserID       int64
	ResourceID   string
	AnomalyType  string
	Severity     string
	Percentage   int
	PreviousCost int
	CurrentCost  int
	DetectedAt   string
	Status       string
}

func (d *DB) CreateCostAnomaly(ctx context.Context, c CostAnomalyRow) (int64, error) {
	res, err := d.sql.ExecContext(ctx, `INSERT INTO cost_anomalies (user_id, resource_id, anomaly_type, severity, percentage, previous_cost, current_cost, detected_at, status) VALUES (?,?,?,?,?,?,?,?,?)`,
		c.UserID, c.ResourceID, c.AnomalyType, c.Severity, c.Percentage, c.PreviousCost, c.CurrentCost, c.DetectedAt, c.Status)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *DB) GetCostAnomalyByID(ctx context.Context, userID int64, id int64) (CostAnomalyRow, bool, error) {
	var c CostAnomalyRow
	err := d.sql.QueryRowContext(ctx, `SELECT id, user_id, resource_id, anomaly_type, severity, percentage, previous_cost, current_cost, detected_at, status FROM cost_anomalies WHERE user_id=? AND id=?`, userID, id).
		Scan(&c.ID, &c.UserID, &c.ResourceID, &c.AnomalyType, &c.Severity, &c.Percentage, &c.PreviousCost, &c.CurrentCost, &c.DetectedAt, &c.Status)
	if err == sql.ErrNoRows {
		return CostAnomalyRow{}, false, nil
	}
	if err != nil {
		return CostAnomalyRow{}, false, err
	}
	return c, true, nil
}

func (d *DB) ListCostAnomalies(ctx context.Context, userID int64, resourceID, severity string) ([]CostAnomalyRow, error) {
	var rows *sql.Rows
	var err error
	if resourceID != "" {
		rows, err = d.sql.QueryContext(ctx, `SELECT id, user_id, resource_id, anomaly_type, severity, percentage, previous_cost, current_cost, detected_at, status FROM cost_anomalies WHERE user_id=? AND resource_id=? ORDER BY id DESC`, userID, resourceID)
	} else if severity != "" {
		rows, err = d.sql.QueryContext(ctx, `SELECT id, user_id, resource_id, anomaly_type, severity, percentage, previous_cost, current_cost, detected_at, status FROM cost_anomalies WHERE user_id=? AND severity=? ORDER BY id DESC`, userID, severity)
	} else {
		rows, err = d.sql.QueryContext(ctx, `SELECT id, user_id, resource_id, anomaly_type, severity, percentage, previous_cost, current_cost, detected_at, status FROM cost_anomalies WHERE user_id=? ORDER BY id DESC`, userID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CostAnomalyRow
	for rows.Next() {
		var c CostAnomalyRow
		if err := rows.Scan(&c.ID, &c.UserID, &c.ResourceID, &c.AnomalyType, &c.Severity, &c.Percentage, &c.PreviousCost, &c.CurrentCost, &c.DetectedAt, &c.Status); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (d *DB) UpdateCostAnomaly(ctx context.Context, userID int64, id int64, resourceID, anomalyType, severity, detectedAt, status *string, percentage, previousCost, currentCost *int) error {
	_, err := d.sql.ExecContext(ctx, `UPDATE cost_anomalies SET resource_id=COALESCE(?, resource_id), anomaly_type=COALESCE(?, anomaly_type), severity=COALESCE(?, severity), detected_at=COALESCE(?, detected_at), status=COALESCE(?, status), percentage=COALESCE(?, percentage), previous_cost=COALESCE(?, previous_cost), current_cost=COALESCE(?, current_cost) WHERE user_id=? AND id=?`,
		resourceID, anomalyType, severity, detectedAt, status, percentage, previousCost, currentCost, userID, id)
	return err
}

func (d *DB) DeleteCostAnomaly(ctx context.Context, userID int64, id int64) error {
	_, err := d.sql.ExecContext(ctx, `DELETE FROM cost_anomalies WHERE user_id=? AND id=?`, userID, id)
	return err
}
