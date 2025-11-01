# Drift Detection Implementation Summary

## Completed Tasks ✅

### 1. Database Migration
- ✅ Created `migrations/002_add_baselines_and_config.sql`
- ✅ Added `configuration` TEXT field to `resources` table
- ✅ Created `resource_baselines` table with indexes
- ✅ Migration executed successfully on `data.db`

### 2. Cloud Scanner Updates
All three cloud provider scanners updated to capture full resource configurations:

#### AWS Scanner (`internal/providers/aws.go`)
- ✅ **EC2 Instances**: Captures security groups, IAM roles, network config, encryption, SSH keys, monitoring, tags
- ✅ **S3 Buckets**: Captures encryption, versioning, public access settings, ACLs, bucket policies, tags

#### GCP Scanner (`internal/providers/gcp.go`)
- ✅ **Compute Engine**: Captures network interfaces, service accounts, disk encryption, shielded instance config, tags/labels
- ✅ **Cloud Storage**: Captures encryption (KMS), versioning, public access prevention, IAM config, lifecycle rules, logging

#### Azure Scanner (`internal/providers/azure.go`)
- ✅ **Virtual Machines**: Captures VM size, network interfaces, disk encryption, managed identity, security profile, tags
- ✅ **Storage Accounts**: Captures encryption, HTTPS-only, public network access, TLS version, blob public access settings

### 3. Drift Detection Components

#### Baseline Management
- ✅ Created `internal/domain/baseline/` package with model, repository, and service interfaces
- ✅ Implemented PostgreSQL repository (`internal/repository/postgres/baseline.go`)
- ✅ Implemented baseline service (`internal/services/baseline_service.go`)
- ✅ Added API handlers (`internal/api/handlers/baseline.go`)

#### Drift Detection Algorithm
- ✅ Created `internal/detector/drift_detector.go`:
  - Deep JSON configuration comparison
  - Recursive field-by-field analysis
  - Change tracking (added/removed/modified)
  - Human-readable diff generation

#### Security Rules Engine
- ✅ Created `internal/detector/security_rules.go`:
  - Resource-specific security rules
  - Automated severity classification
  - Condition-based rule matching
  - Support for encryption, public access, security groups, IAM policies

### 4. API Integration
- ✅ Updated `cmd/api/main.go` with baseline repository and services
- ✅ Added routes in `internal/api/router/router.go`:
  ```
  GET    /api/v1/baselines
  POST   /api/v1/baselines
  GET    /api/v1/baselines/resource/{resourceId}
  DELETE /api/v1/baselines/{id}
  POST   /api/v1/drifts/detect
  ```
- ✅ Enhanced drift service with actual detection logic

### 5. Build & Compilation
- ✅ Fixed all compilation errors
- ✅ Successfully built binary: `infraaudit`
- ✅ All dependencies resolved

---

## Implementation Details

### Configuration Capture
Each resource now stores a comprehensive JSON configuration capturing:

**EC2/GCE/Azure VM:**
```json
{
  "instance_id": "i-123",
  "security_groups": [...],
  "iam_instance_profile": {...},
  "network": {...},
  "encryption": {"ebs_encrypted": true},
  "key_name": "my-key"
}
```

**S3/GCS/Azure Storage:**
```json
{
  "encryption": {"enabled": true},
  "versioning": {"enabled": true},
  "public_access": {"public_access_blocked": true},
  "acl": [...]
}
```

### Drift Detection Workflow

```
1. User triggers scan → Resources scanned with full config
2. On first scan → Automatic baseline created
3. Subsequent scans → Compare current vs baseline
4. Security rules evaluate changes → Assign severity
5. Drifts recorded in database with details
```

### Security Severity Rules

| Severity | Examples |
|----------|----------|
| **Critical** | Encryption disabled, public access enabled, 0.0.0.0/0 security groups |
| **High** | IAM permission escalation, security group/ACL changes |
| **Medium** | Network config changes, SSH key changes, versioning disabled |
| **Low** | Tag changes, minor config updates |

---

## How to Use

### 1. Start the Server
```bash
./infraaudit
```

### 2. Create a Baseline
```bash
curl -X POST http://localhost:8080/api/v1/baselines \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "resource_id": "i-1234567890",
    "provider": "aws",
    "resource_type": "ec2-instance",
    "configuration": "{\"encryption\":{\"enabled\":true}}",
    "baseline_type": "approved",
    "description": "Production approved configuration"
  }'
```

### 3. Trigger Drift Detection
```bash
curl -X POST http://localhost:8080/api/v1/drifts/detect \
  -H "Authorization: Bearer <token>"
```

### 4. List Detected Drifts
```bash
curl http://localhost:8080/api/v1/drifts?severity=critical \
  -H "Authorization: Bearer <token>"
```

### 5. Get Drift Summary
```bash
curl http://localhost:8080/api/v1/drifts/summary \
  -H "Authorization: Bearer <token>"
```

---

## Files Created (10)

1. `migrations/002_add_baselines_and_config.sql`
2. `internal/domain/baseline/model.go`
3. `internal/domain/baseline/repository.go`
4. `internal/domain/baseline/service.go`
5. `internal/repository/postgres/baseline.go`
6. `internal/services/baseline_service.go`
7. `internal/detector/drift_detector.go`
8. `internal/detector/security_rules.go`
9. `internal/api/handlers/baseline.go`
10. `docs/DRIFT_DETECTION.md`

## Files Modified (8)

1. `internal/domain/resource/model.go` - Added Configuration field
2. `internal/services/types.go` - Added Configuration to CloudResource
3. `internal/services/drift_service.go` - Implemented DetectDrifts()
4. `internal/api/handlers/drift.go` - Added Detect() endpoint
5. `internal/api/router/router.go` - Added baseline and detect routes
6. `internal/providers/aws.go` - Enhanced with full config capture
7. `internal/providers/gcp.go` - Enhanced with full config capture
8. `internal/providers/azure.go` - Enhanced with full config capture
9. `cmd/api/main.go` - Wired up baseline services

---

## Next Steps (Optional Enhancements)

1. **Scheduled Detection**: Add cron job or background worker for periodic drift detection
2. **Notifications**: Implement email/Slack alerts for critical drifts
3. **Auto-Remediation**: Suggest or apply fixes for detected drifts
4. **Custom Rules**: Allow users to define their own security rules
5. **Compliance Frameworks**: Map drifts to CIS, NIST, SOC2 controls
6. **Drift History**: Track drift resolution timeline
7. **Resource Grouping**: Organize baselines by environment (prod/staging/dev)
8. **Webhook Integration**: POST drift events to external systems

---

## Testing Checklist

- [x] Database migration successful
- [x] Application builds without errors
- [ ] Cloud scanner captures configurations (requires cloud credentials)
- [ ] Baseline creation API works
- [ ] Drift detection algorithm compares configs
- [ ] Security rules classify severity correctly
- [ ] API endpoints return proper responses
- [ ] Drift detection workflow end-to-end

---

## Performance Notes

- Configuration comparison is performed in-memory (fast)
- Security rules use simple pattern matching (O(n) complexity)
- Database queries use indexes on user_id and resource_id
- Baseline retrieval is cached during detection loop

---

## Documentation

Full documentation available at: **`docs/DRIFT_DETECTION.md`**

Includes:
- Architecture diagrams
- API endpoint examples
- Security rule definitions
- Configuration format specifications
- Troubleshooting guide
