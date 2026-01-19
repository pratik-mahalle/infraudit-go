-- Migration: Add scheduled jobs, remediation actions, and notifications tables
-- Description: This migration adds support for Phase 5 (Automation & Orchestration) 
--              and Phase 6 (Notifications & Integrations)

-- ================================
-- Phase 5: Automation & Orchestration
-- ================================

-- Scheduled Jobs Table
CREATE TABLE IF NOT EXISTS scheduled_jobs (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    job_type VARCHAR(100) NOT NULL CHECK (job_type IN (
        'resource_sync', 'drift_detection', 'vulnerability_scan', 
        'cost_sync', 'iac_scan', 'compliance_assessment', 
        'recommendation', 'anomaly_detection'
    )),
    schedule VARCHAR(100) NOT NULL,  -- Cron expression
    is_enabled BOOLEAN DEFAULT true,
    config TEXT,  -- JSON string
    last_run TIMESTAMP,
    next_run TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_scheduled_jobs_user_id ON scheduled_jobs(user_id);
CREATE INDEX IF NOT EXISTS idx_scheduled_jobs_job_type ON scheduled_jobs(job_type);
CREATE INDEX IF NOT EXISTS idx_scheduled_jobs_is_enabled ON scheduled_jobs(is_enabled);
CREATE INDEX IF NOT EXISTS idx_scheduled_jobs_next_run ON scheduled_jobs(next_run);

-- Job Executions Table
CREATE TABLE IF NOT EXISTS job_executions (
    id TEXT PRIMARY KEY,
    job_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    job_type VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL CHECK (status IN ('pending', 'running', 'completed', 'failed', 'cancelled')),
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    duration_ms INTEGER,
    result TEXT,  -- JSON string
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (job_id) REFERENCES scheduled_jobs(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_job_executions_job_id ON job_executions(job_id);
CREATE INDEX IF NOT EXISTS idx_job_executions_user_id ON job_executions(user_id);
CREATE INDEX IF NOT EXISTS idx_job_executions_status ON job_executions(status);
CREATE INDEX IF NOT EXISTS idx_job_executions_started_at ON job_executions(started_at DESC);

-- Remediation Actions Table
CREATE TABLE IF NOT EXISTS remediation_actions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    drift_id TEXT,
    vulnerability_id TEXT,
    remediation_type VARCHAR(50) NOT NULL CHECK (remediation_type IN ('iac_pr', 'cloud_api', 'policy', 'manual')),
    status VARCHAR(50) NOT NULL CHECK (status IN ('pending', 'approved', 'in_progress', 'completed', 'failed', 'rejected', 'rolled_back')),
    strategy TEXT NOT NULL,  -- JSON string with remediation strategy details
    approval_required BOOLEAN DEFAULT false,
    approved_by TEXT,
    approved_at TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    result TEXT,  -- JSON string
    rollback_data TEXT,  -- JSON string for rollback capability
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (drift_id) REFERENCES drifts(id) ON DELETE SET NULL,
    FOREIGN KEY (vulnerability_id) REFERENCES vulnerabilities(id) ON DELETE SET NULL,
    FOREIGN KEY (approved_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_remediation_actions_user_id ON remediation_actions(user_id);
CREATE INDEX IF NOT EXISTS idx_remediation_actions_status ON remediation_actions(status);
CREATE INDEX IF NOT EXISTS idx_remediation_actions_drift_id ON remediation_actions(drift_id);
CREATE INDEX IF NOT EXISTS idx_remediation_actions_vulnerability_id ON remediation_actions(vulnerability_id);
CREATE INDEX IF NOT EXISTS idx_remediation_actions_type ON remediation_actions(remediation_type);

-- ================================
-- Phase 6: Notifications & Integrations
-- ================================

-- Notification Preferences Table
CREATE TABLE IF NOT EXISTS notification_preferences (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    channel VARCHAR(50) NOT NULL CHECK (channel IN ('slack', 'email', 'webhook')),
    is_enabled BOOLEAN DEFAULT true,
    config TEXT,  -- JSON string with channel-specific config
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, channel),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_notification_preferences_user_id ON notification_preferences(user_id);
CREATE INDEX IF NOT EXISTS idx_notification_preferences_channel ON notification_preferences(channel);

-- Notification Logs Table
CREATE TABLE IF NOT EXISTS notification_logs (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    channel VARCHAR(50) NOT NULL,
    notification_type VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL CHECK (status IN ('pending', 'sent', 'failed', 'retrying')),
    priority VARCHAR(20) CHECK (priority IN ('critical', 'high', 'medium', 'low')),
    payload TEXT,  -- JSON string
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    sent_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_notification_logs_user_id ON notification_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_notification_logs_status ON notification_logs(status);
CREATE INDEX IF NOT EXISTS idx_notification_logs_channel ON notification_logs(channel);
CREATE INDEX IF NOT EXISTS idx_notification_logs_sent_at ON notification_logs(sent_at DESC);
CREATE INDEX IF NOT EXISTS idx_notification_logs_type ON notification_logs(notification_type);

-- Webhooks Table
CREATE TABLE IF NOT EXISTS webhooks (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    name VARCHAR(255) NOT NULL,
    url TEXT NOT NULL,
    secret VARCHAR(255),  -- For HMAC signature
    events TEXT NOT NULL,  -- JSON array of subscribed events
    is_enabled BOOLEAN DEFAULT true,
    retry_config TEXT,  -- JSON string
    last_triggered TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_webhooks_user_id ON webhooks(user_id);
CREATE INDEX IF NOT EXISTS idx_webhooks_is_enabled ON webhooks(is_enabled);

-- Webhook Deliveries Table (for tracking webhook delivery attempts)
CREATE TABLE IF NOT EXISTS webhook_deliveries (
    id TEXT PRIMARY KEY,
    webhook_id TEXT NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload TEXT NOT NULL,  -- JSON string
    status VARCHAR(50) NOT NULL CHECK (status IN ('pending', 'sent', 'failed')),
    response_status INTEGER,
    response_body TEXT,
    retry_count INTEGER DEFAULT 0,
    delivered_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (webhook_id) REFERENCES webhooks(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_webhook_id ON webhook_deliveries(webhook_id);
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_status ON webhook_deliveries(status);
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_event_type ON webhook_deliveries(event_type);
