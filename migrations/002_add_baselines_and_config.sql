-- ALTER TABLE resources ADD COLUMN configuration TEXT;

-- Create resource_baselines table
CREATE TABLE IF NOT EXISTS resource_baselines (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    resource_id VARCHAR(255) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    configuration TEXT NOT NULL,
    baseline_type VARCHAR(50) NOT NULL DEFAULT 'manual',
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(user_id, resource_id, baseline_type)
);

-- Create index for efficient baseline lookups
CREATE INDEX IF NOT EXISTS idx_baselines_user_resource ON resource_baselines(user_id, resource_id);
CREATE INDEX IF NOT EXISTS idx_baselines_type ON resource_baselines(baseline_type);

-- Create index for resource configuration queries
CREATE INDEX IF NOT EXISTS idx_resources_last_scanned ON resources(last_scanned);
