-- Migration: Add IaC definitions table for Infrastructure as Code scanning
-- Description: This table stores uploaded IaC files (Terraform, CloudFormation, Kubernetes)
--              and their parsed resource definitions for drift detection

-- IaC Definitions Table
CREATE TABLE IF NOT EXISTS iac_definitions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    name VARCHAR(255) NOT NULL,
    iac_type VARCHAR(50) NOT NULL CHECK (iac_type IN ('terraform', 'cloudformation', 'kubernetes', 'helm')),
    file_path TEXT,
    content TEXT NOT NULL,
    parsed_resources TEXT, -- JSON string for SQLite
    last_parsed TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_iac_definitions_user_id ON iac_definitions(user_id);
CREATE INDEX IF NOT EXISTS idx_iac_definitions_iac_type ON iac_definitions(iac_type);
CREATE INDEX IF NOT EXISTS idx_iac_definitions_name ON iac_definitions(user_id, name);
CREATE INDEX IF NOT EXISTS idx_iac_definitions_created_at ON iac_definitions(created_at DESC);

-- IaC Resources Table (parsed resources from IaC files for comparison)
CREATE TABLE IF NOT EXISTS iac_resources (
    id TEXT PRIMARY KEY,
    iac_definition_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_name VARCHAR(255) NOT NULL,
    resource_address TEXT, -- Full resource address (e.g., module.vpc.aws_instance.web)
    provider VARCHAR(50) NOT NULL,
    configuration TEXT NOT NULL, -- JSON string
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (iac_definition_id) REFERENCES iac_definitions(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Indexes for IaC resources
CREATE INDEX IF NOT EXISTS idx_iac_resources_definition_id ON iac_resources(iac_definition_id);
CREATE INDEX IF NOT EXISTS idx_iac_resources_user_id ON iac_resources(user_id);
CREATE INDEX IF NOT EXISTS idx_iac_resources_type ON iac_resources(resource_type);
CREATE INDEX IF NOT EXISTS idx_iac_resources_provider ON iac_resources(provider);

-- IaC Drift Comparison Results (separate from regular drift detection)
CREATE TABLE IF NOT EXISTS iac_drift_results (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    iac_definition_id TEXT NOT NULL,
    iac_resource_id TEXT,
    actual_resource_id TEXT,
    drift_category VARCHAR(50) NOT NULL CHECK (drift_category IN ('missing', 'shadow', 'modified', 'compliant')),
    severity VARCHAR(20) CHECK (severity IN ('critical', 'high', 'medium', 'low', 'info')),
    details TEXT, -- JSON string with diff details
    detected_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(50) DEFAULT 'detected' CHECK (status IN ('detected', 'acknowledged', 'resolved', 'ignored')),
    resolved_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (iac_definition_id) REFERENCES iac_definitions(id) ON DELETE CASCADE,
    FOREIGN KEY (iac_resource_id) REFERENCES iac_resources(id) ON DELETE SET NULL,
    FOREIGN KEY (actual_resource_id) REFERENCES resources(id) ON DELETE SET NULL
);

-- Indexes for IaC drift results
CREATE INDEX IF NOT EXISTS idx_iac_drift_user_id ON iac_drift_results(user_id);
CREATE INDEX IF NOT EXISTS idx_iac_drift_definition_id ON iac_drift_results(iac_definition_id);
CREATE INDEX IF NOT EXISTS idx_iac_drift_category ON iac_drift_results(drift_category);
CREATE INDEX IF NOT EXISTS idx_iac_drift_severity ON iac_drift_results(severity);
CREATE INDEX IF NOT EXISTS idx_iac_drift_status ON iac_drift_results(status);
CREATE INDEX IF NOT EXISTS idx_iac_drift_detected_at ON iac_drift_results(detected_at DESC);
