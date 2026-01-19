-- Migration: Add Cost and Compliance tables
-- Phase 3: Cloud Cost Analytics
-- Phase 4: Compliance Framework

-- ========================================
-- PHASE 3: CLOUD COST ANALYTICS
-- ========================================

-- Resource costs table
CREATE TABLE IF NOT EXISTS resource_costs (
    id VARCHAR(36) PRIMARY KEY,
    user_id BIGINT NOT NULL,
    resource_id VARCHAR(36),
    provider VARCHAR(50) NOT NULL,
    region VARCHAR(100),
    service_name VARCHAR(255) NOT NULL,
    resource_type VARCHAR(100),
    cost_date DATE NOT NULL,
    daily_cost DECIMAL(15, 4) NOT NULL,
    monthly_cost DECIMAL(15, 4),
    currency VARCHAR(3) DEFAULT 'USD',
    cost_details JSON,
    tags JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_resource_costs_user_id ON resource_costs(user_id);
CREATE INDEX idx_resource_costs_resource_id ON resource_costs(resource_id);
CREATE INDEX idx_resource_costs_cost_date ON resource_costs(cost_date);
CREATE INDEX idx_resource_costs_provider ON resource_costs(provider);
CREATE INDEX idx_resource_costs_service_name ON resource_costs(service_name);

-- Cost anomalies table
CREATE TABLE IF NOT EXISTS cost_anomalies (
    id VARCHAR(36) PRIMARY KEY,
    user_id BIGINT NOT NULL,
    provider VARCHAR(50) NOT NULL,
    service_name VARCHAR(255) NOT NULL,
    resource_id VARCHAR(36),
    anomaly_type VARCHAR(50) NOT NULL,
    expected_cost DECIMAL(15, 4) NOT NULL,
    actual_cost DECIMAL(15, 4) NOT NULL,
    deviation DECIMAL(10, 2) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    status VARCHAR(50) DEFAULT 'open',
    notes TEXT,
    detected_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_cost_anomalies_user_id ON cost_anomalies(user_id);
CREATE INDEX idx_cost_anomalies_status ON cost_anomalies(status);
CREATE INDEX idx_cost_anomalies_detected_at ON cost_anomalies(detected_at);

-- Cost optimizations table
CREATE TABLE IF NOT EXISTS cost_optimizations (
    id VARCHAR(36) PRIMARY KEY,
    user_id BIGINT NOT NULL,
    provider VARCHAR(50) NOT NULL,
    resource_id VARCHAR(36),
    resource_type VARCHAR(100),
    optimization_type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    current_cost DECIMAL(15, 4) NOT NULL,
    estimated_savings DECIMAL(15, 4) NOT NULL,
    savings_percent DECIMAL(5, 2),
    implementation VARCHAR(50),
    status VARCHAR(50) DEFAULT 'pending',
    details JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_cost_optimizations_user_id ON cost_optimizations(user_id);
CREATE INDEX idx_cost_optimizations_status ON cost_optimizations(status);
CREATE INDEX idx_cost_optimizations_optimization_type ON cost_optimizations(optimization_type);

-- ========================================
-- PHASE 4: COMPLIANCE FRAMEWORK
-- ========================================

-- Compliance frameworks table
CREATE TABLE IF NOT EXISTS compliance_frameworks (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    version VARCHAR(50),
    description TEXT,
    provider VARCHAR(50),
    is_enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Compliance controls table
CREATE TABLE IF NOT EXISTS compliance_controls (
    id VARCHAR(36) PRIMARY KEY,
    framework_id VARCHAR(36) NOT NULL,
    control_id VARCHAR(50) NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    category VARCHAR(100),
    severity VARCHAR(20),
    remediation TEXT,
    reference_url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(framework_id, control_id),
    FOREIGN KEY (framework_id) REFERENCES compliance_frameworks(id) ON DELETE CASCADE
);

CREATE INDEX idx_compliance_controls_framework_id ON compliance_controls(framework_id);
CREATE INDEX idx_compliance_controls_category ON compliance_controls(category);
CREATE INDEX idx_compliance_controls_severity ON compliance_controls(severity);

-- Control mappings table (maps security rules to compliance controls)
CREATE TABLE IF NOT EXISTS compliance_mappings (
    id VARCHAR(36) PRIMARY KEY,
    control_id VARCHAR(36) NOT NULL,
    security_rule_type VARCHAR(100),
    resource_type VARCHAR(100),
    provider VARCHAR(50),
    mapping_confidence VARCHAR(20),
    check_query TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (control_id) REFERENCES compliance_controls(id) ON DELETE CASCADE
);

CREATE INDEX idx_compliance_mappings_control_id ON compliance_mappings(control_id);
CREATE INDEX idx_compliance_mappings_security_rule_type ON compliance_mappings(security_rule_type);

-- Compliance assessments table
CREATE TABLE IF NOT EXISTS compliance_assessments (
    id VARCHAR(36) PRIMARY KEY,
    user_id BIGINT NOT NULL,
    framework_id VARCHAR(36) NOT NULL,
    framework_name VARCHAR(100) NOT NULL,
    assessment_date TIMESTAMP NOT NULL,
    total_controls INT NOT NULL,
    passed_controls INT NOT NULL,
    failed_controls INT NOT NULL,
    not_applicable_controls INT DEFAULT 0,
    compliance_percent DECIMAL(5, 2),
    findings JSON,
    status VARCHAR(50) DEFAULT 'running',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (framework_id) REFERENCES compliance_frameworks(id) ON DELETE CASCADE
);

CREATE INDEX idx_compliance_assessments_user_id ON compliance_assessments(user_id);
CREATE INDEX idx_compliance_assessments_framework_id ON compliance_assessments(framework_id);
CREATE INDEX idx_compliance_assessments_assessment_date ON compliance_assessments(assessment_date);
CREATE INDEX idx_compliance_assessments_status ON compliance_assessments(status);

-- ========================================
-- SEED DATA: Compliance Frameworks
-- ========================================

INSERT INTO compliance_frameworks (id, name, version, description, provider, is_enabled) VALUES
    ('cis-aws-v1.5', 'CIS AWS Foundations Benchmark', '1.5.0', 'CIS Amazon Web Services Foundations Benchmark provides prescriptive guidance for configuring security options for a subset of Amazon Web Services.', 'aws', true),
    ('cis-gcp-v1.3', 'CIS GCP Foundations Benchmark', '1.3.0', 'CIS Google Cloud Platform Foundation Benchmark provides prescriptive guidance for establishing a secure baseline configuration for GCP.', 'gcp', true),
    ('cis-azure-v1.4', 'CIS Azure Foundations Benchmark', '1.4.0', 'CIS Microsoft Azure Foundations Benchmark provides prescriptive guidance for establishing a secure baseline configuration for Azure.', 'azure', true),
    ('nist-800-53-r5', 'NIST 800-53 Rev 5', '5.0', 'NIST Special Publication 800-53 provides a catalog of security and privacy controls for federal information systems.', NULL, true),
    ('soc2-2017', 'SOC 2 Type II', '2017', 'SOC 2 examines controls at a service organization relevant to security, availability, processing integrity, confidentiality, and privacy.', NULL, true),
    ('pci-dss-v4', 'PCI DSS', '4.0', 'Payment Card Industry Data Security Standard - security standards for organizations that handle credit card data.', NULL, false),
    ('hipaa', 'HIPAA', '2013', 'Health Insurance Portability and Accountability Act - US legislation for data privacy and security for safeguarding medical information.', NULL, false),
    ('iso-27001', 'ISO 27001', '2022', 'International standard for managing information security.', NULL, false)
ON CONFLICT (name) DO NOTHING;
