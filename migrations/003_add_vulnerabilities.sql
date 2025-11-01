-- Migration 003: Add vulnerability scanning tables

-- Vulnerability scans table (tracks scan execution history)
CREATE TABLE vulnerability_scans (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    resource_id VARCHAR(255),
    scan_type VARCHAR(50) NOT NULL,  -- trivy, nvd, aws-inspector, gcp-scc, azure-sc
    status VARCHAR(50) NOT NULL,     -- pending, running, completed, failed
    scanner_version VARCHAR(50),
    total_vulnerabilities INTEGER DEFAULT 0,
    critical_count INTEGER DEFAULT 0,
    high_count INTEGER DEFAULT 0,
    medium_count INTEGER DEFAULT 0,
    low_count INTEGER DEFAULT 0,
    scan_duration INTEGER,           -- Duration in seconds
    error_message TEXT,
    metadata TEXT,                   -- JSON metadata
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Vulnerabilities table (individual vulnerability findings)
CREATE TABLE vulnerabilities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    scan_id INTEGER,
    resource_id VARCHAR(255),
    provider VARCHAR(50),            -- aws, gcp, azure, container, os
    resource_type VARCHAR(100),      -- ec2-instance, container-image, lambda, etc.

    -- Vulnerability identification
    cve_id VARCHAR(50),              -- CVE-2023-1234
    vulnerability_id VARCHAR(255),   -- Alternative ID if not CVE
    title VARCHAR(500) NOT NULL,
    description TEXT,

    -- Severity and scoring
    severity VARCHAR(50) NOT NULL,   -- critical, high, medium, low, info
    cvss_score DECIMAL(3, 1),        -- 0.0 - 10.0
    cvss_vector VARCHAR(255),        -- CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H

    -- Package/Component information
    package_name VARCHAR(255),
    package_version VARCHAR(100),
    fixed_version VARCHAR(100),
    package_type VARCHAR(50),        -- os, library, application

    -- Detection details
    scanner_type VARCHAR(50) NOT NULL, -- trivy, nvd, aws-inspector, etc.
    detection_method VARCHAR(100),

    -- Status and remediation
    status VARCHAR(50) DEFAULT 'open', -- open, patched, ignored, false_positive, accepted
    remediation TEXT,                   -- Remediation steps
    reference_urls TEXT,                -- JSON array of reference URLs

    -- Timestamps
    published_date TIMESTAMP,
    last_modified_date TIMESTAMP,
    detected_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (scan_id) REFERENCES vulnerability_scans(id) ON DELETE CASCADE
);

-- Indexes for efficient queries
CREATE INDEX idx_vulnerabilities_user_id ON vulnerabilities(user_id);
CREATE INDEX idx_vulnerabilities_scan_id ON vulnerabilities(scan_id);
CREATE INDEX idx_vulnerabilities_resource_id ON vulnerabilities(resource_id);
CREATE INDEX idx_vulnerabilities_severity ON vulnerabilities(severity);
CREATE INDEX idx_vulnerabilities_status ON vulnerabilities(status);
CREATE INDEX idx_vulnerabilities_cve_id ON vulnerabilities(cve_id);
CREATE INDEX idx_vulnerabilities_detected_at ON vulnerabilities(detected_at);
CREATE INDEX idx_vulnerability_scans_user_id ON vulnerability_scans(user_id);
CREATE INDEX idx_vulnerability_scans_status ON vulnerability_scans(status);
CREATE INDEX idx_vulnerability_scans_created_at ON vulnerability_scans(created_at);
