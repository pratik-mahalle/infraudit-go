-- Fix missing provider_accounts table expected by the application
CREATE TABLE IF NOT EXISTS provider_accounts (
    user_id INTEGER NOT NULL,
    provider VARCHAR(50) NOT NULL,
    is_connected INTEGER DEFAULT 0,
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
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, provider),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_provider_accounts_user_id ON provider_accounts(user_id);
