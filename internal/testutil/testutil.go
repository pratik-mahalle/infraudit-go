package testutil

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

// NewTestDB creates an in-memory SQLite database for testing
func NewTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create schema
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email VARCHAR(255) NOT NULL UNIQUE,
		username VARCHAR(255),
		password_hash VARCHAR(255) NOT NULL,
		full_name VARCHAR(255),
		role VARCHAR(50) NOT NULL DEFAULT 'user',
		plan_type VARCHAR(50) NOT NULL DEFAULT 'free',
		trial_ends_at TIMESTAMP,
		trial_extended BOOLEAN DEFAULT FALSE,
		is_active BOOLEAN DEFAULT TRUE,
		email_verified BOOLEAN DEFAULT FALSE,
		verification_token VARCHAR(255),
		reset_token VARCHAR(255),
		reset_token_expires_at TIMESTAMP,
		last_login_at TIMESTAMP,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS provider_accounts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		provider VARCHAR(50) NOT NULL,
		is_connected BOOLEAN DEFAULT FALSE,
		last_synced TIMESTAMP,
		aws_access_key_id VARCHAR(255),
		aws_secret_access_key VARCHAR(255),
		aws_region VARCHAR(50),
		gcp_project_id VARCHAR(255),
		gcp_service_account_json TEXT,
		gcp_region VARCHAR(50),
		azure_tenant_id VARCHAR(255),
		azure_client_id VARCHAR(255),
		azure_client_secret VARCHAR(255),
		azure_subscription_id VARCHAR(255),
		azure_location VARCHAR(50),
		sync_status VARCHAR(50),
		sync_message TEXT,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		UNIQUE(user_id, provider)
	);

	CREATE TABLE IF NOT EXISTS resources (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		provider VARCHAR(50) NOT NULL,
		type VARCHAR(100) NOT NULL,
		resource_id VARCHAR(255) NOT NULL,
		name VARCHAR(255) NOT NULL,
		region VARCHAR(100),
		status VARCHAR(50),
		cost DECIMAL(10, 2) DEFAULT 0,
		tags TEXT,
		metadata TEXT,
		configuration TEXT,
		last_scanned TIMESTAMP,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		UNIQUE(user_id, provider, resource_id)
	);

	CREATE TABLE IF NOT EXISTS alerts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		type VARCHAR(50) NOT NULL,
		severity VARCHAR(50) NOT NULL,
		title VARCHAR(255) NOT NULL,
		description TEXT,
		resource VARCHAR(255),
		status VARCHAR(50) NOT NULL DEFAULT 'active',
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS recommendations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		type VARCHAR(50) NOT NULL,
		priority VARCHAR(50) NOT NULL,
		title VARCHAR(255) NOT NULL,
		description TEXT,
		savings DECIMAL(10, 2) DEFAULT 0,
		effort VARCHAR(50),
		impact VARCHAR(50),
		category VARCHAR(100),
		resources TEXT,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS drifts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		resource_id VARCHAR(255) NOT NULL,
		drift_type VARCHAR(50) NOT NULL,
		severity VARCHAR(50) NOT NULL,
		details TEXT,
		status VARCHAR(50) NOT NULL DEFAULT 'active',
		detected_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		resolved_at TIMESTAMP,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS anomalies (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		resource_id VARCHAR(255),
		anomaly_type VARCHAR(50) NOT NULL,
		severity VARCHAR(50) NOT NULL,
		percentage INTEGER,
		previous_cost INTEGER,
		current_cost INTEGER,
		detected_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		status VARCHAR(50) NOT NULL DEFAULT 'active',
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS iac_definitions (
		id VARCHAR(36) PRIMARY KEY,
		user_id VARCHAR(36) NOT NULL,
		name VARCHAR(255) NOT NULL,
		iac_type VARCHAR(50) NOT NULL,
		file_path TEXT,
		content TEXT NOT NULL,
		parsed_resources TEXT,
		last_parsed TIMESTAMP,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS iac_resources (
		id VARCHAR(36) PRIMARY KEY,
		iac_definition_id VARCHAR(36) NOT NULL,
		user_id VARCHAR(36) NOT NULL,
		resource_type VARCHAR(100) NOT NULL,
		resource_name VARCHAR(255) NOT NULL,
		resource_address VARCHAR(255),
		provider VARCHAR(50),
		configuration TEXT,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (iac_definition_id) REFERENCES iac_definitions(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS iac_drift_results (
		id VARCHAR(36) PRIMARY KEY,
		user_id VARCHAR(36) NOT NULL,
		iac_definition_id VARCHAR(36) NOT NULL,
		iac_resource_id VARCHAR(36),
		actual_resource_id VARCHAR(36),
		drift_category VARCHAR(100) NOT NULL,
		severity VARCHAR(20),
		details TEXT,
		status VARCHAR(50) NOT NULL,
		detected_at TIMESTAMP NOT NULL,
		resolved_at TIMESTAMP,
		FOREIGN KEY (iac_definition_id) REFERENCES iac_definitions(id) ON DELETE CASCADE
	);
	`

	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}

	return db
}

// CleanupDB closes the test database
func CleanupDB(db *sql.DB) {
	if db != nil {
		db.Close()
	}
}
