# Drift Detection Implementation

## Overview

The drift detection system identifies security configuration changes in cloud resources by comparing current states against stored baseline configurations. This document explains the implementation, architecture, and usage of the drift detection feature.

## Architecture

### Components

1. **Baseline Storage** (`resource_baselines` table)
   - Stores approved/expected configurations for each resource
   - Supports multiple baseline types (manual, automatic, approved)
   - Tracks configuration history

2. **Drift Detector** (`internal/detector/`)
   - Compares current vs baseline configurations
   - Identifies security-relevant changes
   - Calculates severity levels
   - Generates detailed drift reports

3. **Security Rules** (`internal/detector/security_rules.go`)
   - Defines what changes are security-relevant
   - Resource-type specific rules
   - Severity classification logic

4. **Drift Service** (`internal/services/drift_service.go`)
   - Orchestrates detection workflow
   - Creates automatic baselines
   - Records detected drifts

## How It Works

### 1. Baseline Creation

When resources are first scanned, automatic baselines are created:

```go
baseline := &baseline.Baseline{
    UserID:        userID,
    ResourceID:    resourceID,
    Provider:      "aws",
    ResourceType:  "s3-bucket",
    Configuration: `{"encryption": true, "public": false, ...}`,
    BaselineType:  "automatic",
}
```

Users can also create manual baselines via API:

```bash
POST /api/v1/baselines
{
  "resource_id": "i-1234567890",
  "provider": "aws",
  "resource_type": "ec2-instance",
  "configuration": "{...}",
  "baseline_type": "approved",
  "description": "Production-approved configuration"
}
```

### 2. Configuration Comparison

The drift detector performs deep comparison of JSON configurations:

```
Baseline Config          Current Config
{                        {
  "encryption": true  →    "encryption": false  ❌ DRIFT DETECTED
  "public": false     →    "public": true       ❌ CRITICAL DRIFT
  "versioning": true  →    "versioning": true   ✓ No change
}
```

### 3. Security Rule Evaluation

Each change is evaluated against security rules:

#### Critical Severity Rules
- **Encryption Disabled**: Encryption changed from enabled → disabled
- **Public Access Enabled**: Resource changed from private → public
- **Unrestricted Security Groups**: 0.0.0.0/0 or ::/0 access added

#### High Severity Rules
- **IAM Permission Escalation**: Wildcard (*) or admin permissions added
- **Security Group Changes**: Firewall rules modified
- **Bucket Policy Changes**: S3/storage access policies modified

#### Medium Severity Rules
- **Network Configuration**: VPC, subnet, or network settings changed
- **SSH Configuration**: SSH keys or access changed
- **Versioning Changes**: Object versioning disabled

#### Low Severity Rules
- **Tag Changes**: Resource tags modified
- **Minor Configuration**: Non-security config changes

### 4. Drift Recording

When drifts are detected, records are created:

```go
drift := &drift.Drift{
    UserID:     userID,
    ResourceID: "s3-bucket-123",
    DriftType:  "encryption",
    Severity:   "critical",
    Details:    "2 configuration change(s) detected...",
    Status:     "detected",
}
```

## Database Schema

### `resource_baselines` Table

```sql
CREATE TABLE resource_baselines (
    id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL,
    resource_id VARCHAR(255) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    configuration TEXT NOT NULL,
    baseline_type VARCHAR(50) DEFAULT 'manual',
    description TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    UNIQUE(user_id, resource_id, baseline_type)
);
```

### `resources` Table (Updated)

Added `configuration` field to store current resource configuration:

```sql
ALTER TABLE resources ADD COLUMN configuration TEXT;
```

## API Endpoints

### Baseline Management

#### Create Baseline
```http
POST /api/v1/baselines
Content-Type: application/json
Authorization: Bearer <token>

{
  "resource_id": "i-1234567890",
  "provider": "aws",
  "resource_type": "ec2-instance",
  "configuration": "{\"encryption\": true, \"public\": false}",
  "baseline_type": "approved",
  "description": "Production-approved configuration"
}
```

#### Get Baseline for Resource
```http
GET /api/v1/baselines/resource/{resourceId}?type=approved
Authorization: Bearer <token>
```

#### List All Baselines
```http
GET /api/v1/baselines
Authorization: Bearer <token>
```

#### Delete Baseline
```http
DELETE /api/v1/baselines/{id}
Authorization: Bearer <token>
```

### Drift Detection

#### Trigger Drift Detection
```http
POST /api/v1/drifts/detect
Authorization: Bearer <token>
```

**Response:**
```json
{
  "message": "Drift detection completed successfully"
}
```

#### List Detected Drifts
```http
GET /api/v1/drifts?severity=critical&status=detected
Authorization: Bearer <token>
```

**Response:**
```json
{
  "drifts": [
    {
      "id": 1,
      "resource_id": "s3-bucket-123",
      "drift_type": "encryption",
      "severity": "critical",
      "details": "2 configuration change(s) detected...",
      "status": "detected",
      "detected_at": "2025-11-01T10:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 20
}
```

#### Get Drift Summary
```http
GET /api/v1/drifts/summary
Authorization: Bearer <token>
```

**Response:**
```json
{
  "critical": 2,
  "high": 5,
  "medium": 10,
  "low": 3
}
```

#### Update Drift Status
```http
PUT /api/v1/drifts/{id}
Content-Type: application/json
Authorization: Bearer <token>

{
  "status": "acknowledged"
}
```

## Usage Workflow

### 1. Initial Setup

```bash
# Run migration to create baselines table
sqlite3 infraaudit.db < migrations/002_add_baselines_and_config.sql

# Scan cloud resources (captures configuration)
POST /api/v1/providers/aws/sync
```

### 2. Create Approved Baselines

For production resources, create approved baselines:

```bash
# Get current resource config
GET /api/v1/resources/i-1234567890

# Create baseline from current config
POST /api/v1/baselines
{
  "resource_id": "i-1234567890",
  "configuration": "<current_config>",
  "baseline_type": "approved"
}
```

### 3. Run Drift Detection

```bash
# Manual detection
POST /api/v1/drifts/detect

# Check for drifts
GET /api/v1/drifts?severity=critical

# Get summary
GET /api/v1/drifts/summary
```

### 4. Handle Detected Drifts

```bash
# Review drift details
GET /api/v1/drifts/123

# Acknowledge drift
PUT /api/v1/drifts/123
{"status": "acknowledged"}

# Fix the issue in cloud provider

# Update baseline if change was intentional
POST /api/v1/baselines
{
  "resource_id": "i-1234567890",
  "configuration": "<new_approved_config>",
  "baseline_type": "approved"
}

# Mark drift as resolved
PUT /api/v1/drifts/123
{"status": "resolved"}
```

## Security Rules by Resource Type

### S3 Buckets / GCS Buckets / Azure Storage

- **Critical**: Encryption disabled, public access enabled
- **High**: ACL/policy changes
- **Medium**: Versioning changes
- **Low**: Tag changes

### EC2 / GCE / Azure VM Instances

- **Critical**: Security groups opened to 0.0.0.0/0
- **High**: IAM role changes, permission escalation
- **Medium**: SSH config changes, network changes
- **Low**: Tag changes

### RDS / Cloud SQL Instances

- **Critical**: Encryption disabled, public access enabled
- **High**: Security group changes
- **Medium**: Backup configuration changes
- **Low**: Parameter changes

## Example Drift Scenarios

### Scenario 1: Encryption Disabled

**Before (Baseline):**
```json
{
  "name": "prod-bucket",
  "encryption": {
    "enabled": true,
    "algorithm": "AES256"
  },
  "public": false
}
```

**After (Current):**
```json
{
  "name": "prod-bucket",
  "encryption": {
    "enabled": false
  },
  "public": false
}
```

**Result:**
- Drift Type: `encryption`
- Severity: `critical`
- Details: "Encryption was disabled for production bucket"

### Scenario 2: Security Group Opened

**Before:**
```json
{
  "security_groups": [
    {"cidr": "10.0.0.0/8", "port": 22}
  ]
}
```

**After:**
```json
{
  "security_groups": [
    {"cidr": "0.0.0.0/0", "port": 22}
  ]
}
```

**Result:**
- Drift Type: `security_group`
- Severity: `critical`
- Details: "Security group allows unrestricted SSH access from internet"

## Next Steps

To complete the drift detection implementation:

1. **Add Configuration Capture to Cloud Scans**
   - Update AWS/GCP/Azure scanners to capture full resource configurations
   - Store configurations in `resources.configuration` field

2. **Implement Scheduled Detection**
   - Add cron job or background worker
   - Run drift detection periodically (hourly/daily)

3. **Add Notification System**
   - Email/Slack alerts for critical drifts
   - Webhook support for integration

4. **Enhance Security Rules**
   - Add more resource-type specific rules
   - Support custom user-defined rules
   - Compliance framework mappings (CIS, SOC2, etc.)

5. **Add Drift Remediation**
   - Suggest fixes for detected drifts
   - Auto-remediation for approved changes
   - Terraform/CloudFormation generation

## Files Modified/Created

### Created
- `migrations/002_add_baselines_and_config.sql` - Database migration
- `internal/domain/baseline/model.go` - Baseline domain model
- `internal/domain/baseline/repository.go` - Baseline repository interface
- `internal/domain/baseline/service.go` - Baseline service interface
- `internal/repository/postgres/baseline.go` - Baseline repository implementation
- `internal/services/baseline_service.go` - Baseline service implementation
- `internal/detector/drift_detector.go` - Core drift detection algorithm
- `internal/detector/security_rules.go` - Security rule evaluation
- `internal/api/handlers/baseline.go` - Baseline API handlers
- `docs/DRIFT_DETECTION.md` - This documentation

### Modified
- `internal/domain/resource/model.go` - Added Configuration field
- `internal/services/drift_service.go` - Implemented DetectDrifts()
- `internal/api/handlers/drift.go` - Added Detect() endpoint
- `internal/api/router/router.go` - Added baseline and detect routes
- `cmd/api/main.go` - Wired up baseline repo and service

## Testing

To test the implementation:

1. **Run migration:**
   ```bash
   sqlite3 infraaudit.db < migrations/002_add_baselines_and_config.sql
   ```

2. **Build and run:**
   ```bash
   go build -o infraaudit cmd/api/main.go
   ./infraaudit
   ```

3. **Test API endpoints:**
   ```bash
   # Create a test baseline
   curl -X POST http://localhost:8080/api/v1/baselines \
     -H "Authorization: Bearer <token>" \
     -H "Content-Type: application/json" \
     -d '{
       "resource_id": "test-resource",
       "provider": "aws",
       "resource_type": "s3-bucket",
       "configuration": "{\"encryption\": true}",
       "baseline_type": "manual"
     }'

   # Trigger detection
   curl -X POST http://localhost:8080/api/v1/drifts/detect \
     -H "Authorization: Bearer <token>"

   # List drifts
   curl http://localhost:8080/api/v1/drifts \
     -H "Authorization: Bearer <token>"
   ```
