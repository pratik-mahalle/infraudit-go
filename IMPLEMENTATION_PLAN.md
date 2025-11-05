# InfraAudit - Complete Infrastructure Scanning & Drift Detection Implementation Plan

**Version**: 1.0
**Date**: 2025-11-04
**Status**: Planning Phase

---

## Executive Summary

This plan outlines the complete implementation of InfraAudit's infrastructure scanning and drift detection features, transforming it into a comprehensive Infrastructure as Code (IaC) auditing platform with automated drift detection, compliance scanning, cost optimization, and auto-remediation capabilities.

### Current State (85% Complete)
✅ Multi-cloud resource scanning (AWS, GCP, Azure)
✅ Configuration drift detection with security rules
✅ Vulnerability scanning (Trivy, NVD)
✅ AI-powered recommendations (Gemini)
✅ Complete REST API with authentication

### Target State (100% Complete)
🎯 IaC parsing and baseline comparison
🎯 Automated compliance scanning (CIS, NIST, SOC2)
🎯 Cloud cost analytics and optimization
🎯 Continuous monitoring with scheduled scans
🎯 Auto-remediation workflows
🎯 Multi-channel notifications

---

## Implementation Phases

### **Phase 1: IaC Parsing & Baseline Engine** (Critical - 2-3 weeks)
Build the foundation for comparing IaC definitions with deployed resources.

### **Phase 2: Security Scanner Integration** (High Priority - 1-2 weeks)
Integrate industry-standard IaC security scanners.

### **Phase 3: Cloud Cost Analytics** (High Priority - 2 weeks)
Implement real-time cost tracking and optimization.

### **Phase 4: Compliance Framework** (Medium Priority - 2 weeks)
Map security rules to compliance standards.

### **Phase 5: Automation & Orchestration** (Medium Priority - 2-3 weeks)
Build background job scheduler and auto-remediation.

### **Phase 6: Notifications & Integrations** (Low Priority - 1 week)
Complete notification system for alerts and reports.

---

## Phase 1: IaC Parsing & Baseline Engine

### Objective
Enable InfraAudit to parse IaC definitions (Terraform, CloudFormation, Kubernetes) and compare them with actual deployed resources to detect configuration drift.

### Components to Build

#### 1.1 Terraform HCL Parser
**Location**: `internal/iac/terraform/`

**Features**:
- Parse Terraform `.tf` files using `hashicorp/hcl/v2`
- Parse Terraform state files (`terraform.tfstate`)
- Extract resource definitions (aws_instance, google_compute_instance, azurerm_virtual_machine)
- Support for modules and variables
- Capture resource attributes (tags, security groups, encryption settings)

**Files to Create**:
```
internal/iac/terraform/
├── parser.go           - HCL parsing logic
├── state_parser.go     - Terraform state file parser
├── resource_mapper.go  - Map TF resources to InfraAudit resource types
└── types.go            - Terraform resource models
```

**Database Schema Addition**:
```sql
-- New table for IaC definitions
CREATE TABLE iac_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    iac_type VARCHAR(50) NOT NULL, -- terraform, cloudformation, kubernetes
    file_path TEXT,
    content TEXT NOT NULL,
    parsed_resources JSONB,
    last_parsed TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_iac_definitions_user_id ON iac_definitions(user_id);
CREATE INDEX idx_iac_definitions_iac_type ON iac_definitions(iac_type);
```

**API Endpoints**:
```
POST   /api/v1/iac/upload              - Upload IaC file
GET    /api/v1/iac/definitions         - List IaC definitions
GET    /api/v1/iac/definitions/{id}    - Get IaC definition
DELETE /api/v1/iac/definitions/{id}    - Delete IaC definition
POST   /api/v1/iac/parse               - Parse IaC file
POST   /api/v1/iac/compare             - Compare IaC with deployed resources
```

**Dependencies**:
```go
github.com/hashicorp/hcl/v2
github.com/hashicorp/terraform-config-inspect
```

**Success Criteria**:
- ✅ Parse Terraform files and extract resource definitions
- ✅ Parse Terraform state files
- ✅ Map Terraform resources to cloud provider resources
- ✅ Store parsed IaC definitions in database
- ✅ Compare IaC baseline with actual resources

---

#### 1.2 CloudFormation Parser
**Location**: `internal/iac/cloudformation/`

**Features**:
- Parse CloudFormation templates (JSON/YAML)
- Extract AWS resource definitions
- Support for nested stacks and parameters
- Intrinsic function resolution (Ref, GetAtt, Sub)

**Files to Create**:
```
internal/iac/cloudformation/
├── parser.go           - CloudFormation template parser
├── resource_mapper.go  - Map CFN resources to InfraAudit types
└── types.go            - CloudFormation models
```

**Dependencies**:
```go
gopkg.in/yaml.v3  (already available)
encoding/json     (standard library)
```

**Success Criteria**:
- ✅ Parse CloudFormation JSON/YAML templates
- ✅ Extract AWS resource definitions
- ✅ Map CFN resources to InfraAudit resource types
- ✅ Handle intrinsic functions

---

#### 1.3 Kubernetes Manifest Parser
**Location**: `internal/iac/kubernetes/`

**Features**:
- Parse Kubernetes YAML manifests
- Extract workload definitions (Deployments, StatefulSets, DaemonSets)
- Extract service and networking configs
- Parse Helm charts
- Support for Kustomize overlays

**Files to Create**:
```
internal/iac/kubernetes/
├── parser.go           - K8s manifest parser
├── helm_parser.go      - Helm chart parser
├── resource_mapper.go  - Map K8s resources to InfraAudit types
└── types.go            - Kubernetes models
```

**Dependencies**:
```go
k8s.io/api
k8s.io/apimachinery
k8s.io/client-go
helm.sh/helm/v3
```

**Success Criteria**:
- ✅ Parse Kubernetes YAML manifests
- ✅ Extract workload and service definitions
- ✅ Parse Helm charts
- ✅ Map K8s resources to InfraAudit types

---

#### 1.4 IaC Drift Detector
**Location**: `internal/detector/iac_drift_detector.go`

**Features**:
- Compare IaC definitions with actual deployed resources
- Detect resources defined in IaC but not deployed
- Detect resources deployed but not in IaC (shadow resources)
- Detect configuration differences between IaC and actual state
- Generate drift reports with remediation suggestions

**Algorithm**:
```
1. Load IaC definition from database
2. Parse IaC file to extract resource definitions
3. Fetch actual deployed resources from cloud provider
4. For each IaC resource:
   a. Find matching deployed resource (by ID, name, or tags)
   b. Compare configurations (deep JSON diff)
   c. Classify drift severity using security rules
   d. Generate drift record
5. Identify shadow resources (deployed but not in IaC)
6. Identify missing resources (in IaC but not deployed)
7. Generate comprehensive drift report
```

**API Endpoints**:
```
POST   /api/v1/iac/drifts/detect       - Detect IaC drift
GET    /api/v1/iac/drifts              - List IaC drifts
GET    /api/v1/iac/drifts/summary      - Get IaC drift summary
```

**Success Criteria**:
- ✅ Compare IaC definitions with deployed resources
- ✅ Detect shadow resources
- ✅ Detect missing resources
- ✅ Generate actionable drift reports

---

### Phase 1 Deliverables
- [ ] Terraform HCL parser and state file parser
- [ ] CloudFormation template parser
- [ ] Kubernetes manifest parser
- [ ] IaC drift detection engine
- [ ] Database schema for IaC definitions
- [ ] API endpoints for IaC management
- [ ] Unit tests for all parsers
- [ ] Integration tests for drift detection

**Estimated Time**: 2-3 weeks
**Priority**: Critical
**Dependencies**: None (builds on existing drift detection)

---

## Phase 2: Security Scanner Integration

### Objective
Integrate industry-standard IaC security scanners (Checkov, tfsec, Trivy) to identify misconfigurations, security vulnerabilities, and compliance violations in IaC files before deployment.

### Components to Build

#### 2.1 Checkov Integration
**Location**: `internal/scanners/checkov.go`

**Features**:
- Scan Terraform files for security issues
- Scan CloudFormation templates
- Scan Kubernetes manifests
- CIS Benchmarks compliance checks
- OWASP Top 10 checks
- Parse Checkov JSON output
- Map Checkov findings to InfraAudit vulnerability model

**Installation**:
```bash
pip3 install checkov
```

**Execution**:
```bash
checkov -d /path/to/iac --framework terraform --output json
checkov -d /path/to/iac --framework cloudformation --output json
checkov -d /path/to/iac --framework kubernetes --output json
```

**Files to Create**:
```
internal/scanners/checkov.go    - Checkov scanner implementation
```

**Success Criteria**:
- ✅ Execute Checkov scans on IaC files
- ✅ Parse Checkov JSON output
- ✅ Map findings to InfraAudit vulnerability model
- ✅ Store findings in vulnerabilities table

---

#### 2.2 tfsec Integration
**Location**: `internal/scanners/tfsec.go`

**Features**:
- Terraform-specific security scanner
- Faster than Checkov for Terraform
- AWS, GCP, Azure provider-specific checks
- Custom rule support
- Parse tfsec JSON output

**Installation**:
```bash
brew install tfsec  # or download binary
```

**Execution**:
```bash
tfsec /path/to/terraform --format json
```

**Files to Create**:
```
internal/scanners/tfsec.go      - tfsec scanner implementation
```

**Success Criteria**:
- ✅ Execute tfsec scans on Terraform files
- ✅ Parse tfsec JSON output
- ✅ Map findings to InfraAudit vulnerability model
- ✅ Store findings with severity and remediation

---

#### 2.3 Trivy IaC Scanning (Extend Existing)
**Location**: `internal/scanners/trivy.go` (extend existing)

**Features**:
- Extend existing Trivy integration to support IaC scanning
- Scan Terraform, CloudFormation, Kubernetes
- Dockerfile misconfiguration detection

**Execution**:
```bash
trivy config /path/to/iac --format json
```

**Modifications**:
- Add `ScanIaC` method to existing Trivy scanner
- Add IaC-specific vulnerability parsing

**Success Criteria**:
- ✅ Extend Trivy to scan IaC files
- ✅ Parse IaC-specific findings
- ✅ Integrate with existing Trivy workflow

---

#### 2.4 Scanner Orchestration Service
**Location**: `internal/services/iac_scan_service.go`

**Features**:
- Orchestrate multiple scanners (Checkov, tfsec, Trivy)
- Run scanners in parallel for performance
- Aggregate findings from all scanners
- Deduplicate findings across scanners
- Prioritize findings by severity and confidence

**Algorithm**:
```
1. Receive IaC scan request (iac_definition_id)
2. Create vulnerability_scan record (status: running)
3. Run scanners in parallel:
   - Checkov scan
   - tfsec scan (if Terraform)
   - Trivy IaC scan
4. Collect all findings
5. Deduplicate findings (same resource + same issue)
6. Merge findings with highest severity
7. Store vulnerabilities in database
8. Update scan record (status: completed)
9. Trigger recommendation generation
10. Send notifications if critical findings
```

**API Endpoints**:
```
POST   /api/v1/iac/scan                - Scan IaC file with all scanners
GET    /api/v1/iac/scan/{id}/results   - Get scan results
GET    /api/v1/iac/vulnerabilities     - List IaC vulnerabilities
```

**Success Criteria**:
- ✅ Orchestrate multiple scanners
- ✅ Aggregate and deduplicate findings
- ✅ Store findings with remediation steps
- ✅ Generate security score for IaC files

---

### Phase 2 Deliverables
- [ ] Checkov scanner integration
- [ ] tfsec scanner integration
- [ ] Trivy IaC scanning extension
- [ ] Scanner orchestration service
- [ ] API endpoints for IaC scanning
- [ ] Deduplication and aggregation logic
- [ ] Unit tests for all scanners
- [ ] Integration tests for orchestration

**Estimated Time**: 1-2 weeks
**Priority**: High
**Dependencies**: Phase 1 (IaC parsing)

---

## Phase 3: Cloud Cost Analytics

### Objective
Implement real-time cloud cost tracking and optimization using native cloud provider billing APIs to provide accurate cost insights and savings recommendations.

### Components to Build

#### 3.1 AWS Cost Explorer Integration
**Location**: `internal/providers/cost/aws_cost.go`

**Features**:
- Fetch cost and usage data using AWS Cost Explorer API
- Support for time range queries (daily, weekly, monthly)
- Group costs by service, region, resource
- Filter by tags for resource-specific costs
- Forecast future costs
- Identify unused resources (zero cost but existing)

**AWS SDK**:
```go
github.com/aws/aws-sdk-go-v2/service/costexplorer
```

**API Methods**:
```go
GetCostAndUsage()      - Get historical costs
GetCostForecast()      - Get cost forecasts
GetRightsizingRecommendation() - Get AWS recommendations
GetReservationPurchaseRecommendation() - Reserved instance recommendations
```

**Files to Create**:
```
internal/providers/cost/
├── aws_cost.go         - AWS Cost Explorer integration
├── types.go            - Cost data models
└── mapper.go           - Map costs to resources
```

**Database Schema Addition**:
```sql
CREATE TABLE resource_costs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    resource_id UUID REFERENCES resources(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    cost_date DATE NOT NULL,
    daily_cost DECIMAL(10, 4) NOT NULL,
    monthly_cost DECIMAL(10, 4),
    currency VARCHAR(3) DEFAULT 'USD',
    cost_details JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_resource_costs_user_id ON resource_costs(user_id);
CREATE INDEX idx_resource_costs_resource_id ON resource_costs(resource_id);
CREATE INDEX idx_resource_costs_cost_date ON resource_costs(cost_date);
CREATE INDEX idx_resource_costs_provider ON resource_costs(provider);
```

**Success Criteria**:
- ✅ Fetch AWS cost data for resources
- ✅ Store costs in database with daily granularity
- ✅ Calculate monthly costs and trends
- ✅ Identify unused resources

---

#### 3.2 GCP Billing API Integration
**Location**: `internal/providers/cost/gcp_cost.go`

**Features**:
- Fetch billing data using GCP Cloud Billing API
- Export billing data from BigQuery
- Group costs by service, region, SKU
- Filter by labels for resource-specific costs
- Cost forecasting using historical data

**GCP SDK**:
```go
cloud.google.com/go/billing
```

**API Methods**:
```go
ListBillingAccounts()
ExportBillingData()     - Export to BigQuery
QueryBillingData()      - Query BigQuery for costs
```

**Files to Create**:
```
internal/providers/cost/gcp_cost.go
```

**Success Criteria**:
- ✅ Fetch GCP billing data
- ✅ Map costs to GCP resources
- ✅ Store costs in unified format
- ✅ Calculate resource-level costs

---

#### 3.3 Azure Cost Management Integration
**Location**: `internal/providers/cost/azure_cost.go`

**Features**:
- Fetch cost data using Azure Cost Management API
- Support for usage and charges queries
- Group costs by resource group, resource type, location
- Filter by tags
- Budget and forecast support

**Azure SDK**:
```go
github.com/Azure/azure-sdk-for-go/services/costmanagement
```

**API Methods**:
```go
QueryUsage()           - Get usage and cost data
QueryForecast()        - Get cost forecasts
```

**Files to Create**:
```
internal/providers/cost/azure_cost.go
```

**Success Criteria**:
- ✅ Fetch Azure cost data
- ✅ Map costs to Azure resources
- ✅ Store costs in unified format
- ✅ Track cost trends

---

#### 3.4 Cost Analytics Service
**Location**: `internal/services/cost_service.go`

**Features**:
- Unified cost API across all cloud providers
- Cost trend analysis (daily, weekly, monthly)
- Cost anomaly detection (sudden spikes)
- Cost optimization recommendations
- Total spend tracking per user
- Cost allocation by tags/labels
- Budget alerts and forecasting

**API Endpoints**:
```
GET    /api/v1/costs                   - Get cost overview
GET    /api/v1/costs/resources         - Get resource-level costs
GET    /api/v1/costs/trends            - Get cost trends
GET    /api/v1/costs/forecast          - Get cost forecast
GET    /api/v1/costs/anomalies         - Detect cost anomalies
POST   /api/v1/costs/sync              - Sync costs from cloud providers
```

**Cost Optimization Recommendations**:
- Rightsizing recommendations (oversized instances)
- Unused resource identification (zero cost/usage)
- Reserved instance opportunities
- Storage class optimization
- Region cost comparison
- Idle resource detection

**Success Criteria**:
- ✅ Unified cost API for all providers
- ✅ Real-time cost tracking
- ✅ Cost anomaly detection
- ✅ Actionable optimization recommendations
- ✅ Integration with AI recommendation engine

---

### Phase 3 Deliverables
- [ ] AWS Cost Explorer integration
- [ ] GCP Billing API integration
- [ ] Azure Cost Management integration
- [ ] Cost analytics service
- [ ] Database schema for costs
- [ ] API endpoints for cost analytics
- [ ] Cost optimization recommendations
- [ ] Cost anomaly detection
- [ ] Unit tests and integration tests

**Estimated Time**: 2 weeks
**Priority**: High
**Dependencies**: None (extends existing resource management)

---

## Phase 4: Compliance Framework

### Objective
Map InfraAudit's security rules and findings to industry-standard compliance frameworks (CIS Benchmarks, NIST, SOC2, PCI-DSS, HIPAA) to provide compliance reporting and gap analysis.

### Components to Build

#### 4.1 Compliance Rule Mapping Engine
**Location**: `internal/compliance/`

**Features**:
- Map security rules to compliance controls
- Support multiple compliance frameworks
- Generate compliance reports
- Track compliance status over time
- Identify gaps and remediation steps

**Files to Create**:
```
internal/compliance/
├── frameworks.go       - Framework definitions (CIS, NIST, SOC2)
├── mapper.go           - Map security rules to controls
├── reporter.go         - Generate compliance reports
├── rules/
│   ├── cis_aws.go      - CIS AWS Foundations Benchmark
│   ├── cis_gcp.go      - CIS GCP Foundations Benchmark
│   ├── cis_azure.go    - CIS Azure Foundations Benchmark
│   ├── nist.go         - NIST 800-53 controls
│   ├── soc2.go         - SOC2 Trust Service Criteria
│   └── pci_dss.go      - PCI-DSS requirements
└── types.go
```

**Database Schema Addition**:
```sql
CREATE TABLE compliance_frameworks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    version VARCHAR(50),
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE compliance_controls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    framework_id UUID NOT NULL REFERENCES compliance_frameworks(id),
    control_id VARCHAR(50) NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    category VARCHAR(100),
    severity VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(framework_id, control_id)
);

CREATE TABLE compliance_mappings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    control_id UUID NOT NULL REFERENCES compliance_controls(id),
    security_rule_type VARCHAR(100),  -- Maps to drift_type
    resource_type VARCHAR(100),
    mapping_confidence VARCHAR(20),   -- high, medium, low
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE compliance_assessments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    framework_id UUID NOT NULL REFERENCES compliance_frameworks(id),
    assessment_date TIMESTAMP NOT NULL,
    total_controls INT NOT NULL,
    passed_controls INT NOT NULL,
    failed_controls INT NOT NULL,
    not_applicable_controls INT NOT NULL,
    compliance_percentage DECIMAL(5, 2),
    findings JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_compliance_assessments_user_id ON compliance_assessments(user_id);
CREATE INDEX idx_compliance_assessments_framework_id ON compliance_assessments(framework_id);
```

**Compliance Framework Examples**:

**CIS AWS Foundations Benchmark v1.5.0**:
- 1.1 - Maintain current contact details
- 1.2 - Ensure security contact information is registered
- 2.1.1 - Ensure S3 bucket has server-side encryption enabled
- 2.1.2 - Ensure S3 bucket has access logging enabled
- 4.1 - Ensure no security groups allow ingress from 0.0.0.0/0 to port 22
- 4.2 - Ensure no security groups allow ingress from 0.0.0.0/0 to port 3389

**NIST 800-53 Controls**:
- AC-2 - Account Management
- AC-3 - Access Enforcement
- AU-2 - Audit Events
- CM-6 - Configuration Settings
- SC-7 - Boundary Protection
- SC-28 - Protection of Information at Rest

**SOC2 Trust Service Criteria**:
- CC6.1 - Logical and physical access controls
- CC6.6 - Encryption of data at rest and in transit
- CC7.2 - System monitoring

**Mapping Logic**:
```go
// Example: Map drift detection to CIS control
drift.Type = "encryption"
drift.ResourceType = "s3_bucket"
-> CIS AWS 2.1.1: Ensure S3 bucket has server-side encryption enabled

drift.Type = "security_group"
drift.Details contains "0.0.0.0/0" and port 22
-> CIS AWS 4.1: Ensure no security groups allow ingress from 0.0.0.0/0 to port 22
```

**Success Criteria**:
- ✅ Define major compliance frameworks (CIS, NIST, SOC2)
- ✅ Map security rules to compliance controls
- ✅ Generate compliance assessment reports
- ✅ Track compliance over time

---

#### 4.2 Compliance Service
**Location**: `internal/services/compliance_service.go`

**Features**:
- Run compliance assessments on user's infrastructure
- Generate compliance reports (PDF, JSON)
- Identify failing controls
- Provide remediation guidance
- Track compliance score over time

**API Endpoints**:
```
GET    /api/v1/compliance/frameworks          - List available frameworks
POST   /api/v1/compliance/assess              - Run compliance assessment
GET    /api/v1/compliance/assessments         - List past assessments
GET    /api/v1/compliance/assessments/{id}    - Get assessment details
GET    /api/v1/compliance/report/{id}         - Download compliance report
GET    /api/v1/compliance/controls/failing    - Get failing controls
```

**Assessment Algorithm**:
```
1. Select compliance framework (e.g., CIS AWS)
2. Fetch all controls for framework
3. For each control:
   a. Find mapped security rules
   b. Check drifts table for violations
   c. Check vulnerabilities table for findings
   d. Determine control status (passed/failed/N/A)
4. Calculate compliance percentage
5. Generate findings with remediation steps
6. Store assessment in database
7. Return assessment report
```

**Success Criteria**:
- ✅ Run automated compliance assessments
- ✅ Generate actionable compliance reports
- ✅ Track compliance trends
- ✅ Integrate with recommendation engine

---

### Phase 4 Deliverables
- [ ] Compliance framework definitions (CIS, NIST, SOC2)
- [ ] Rule mapping engine
- [ ] Compliance assessment service
- [ ] Database schema for compliance
- [ ] API endpoints for compliance
- [ ] Compliance report generation
- [ ] Unit tests and integration tests

**Estimated Time**: 2 weeks
**Priority**: Medium
**Dependencies**: Phase 1 (drift detection) and Phase 2 (security scanning)

---

## Phase 5: Automation & Orchestration

### Objective
Build a background job scheduler for continuous infrastructure scanning, automated drift detection, and self-healing workflows with auto-remediation capabilities.

### Components to Build

#### 5.1 Background Job Scheduler
**Location**: `internal/worker/scheduler.go`

**Features**:
- Cron-based job scheduling
- Distributed job execution (if multi-instance)
- Job retry with exponential backoff
- Job status tracking and monitoring
- Concurrent job execution with rate limiting

**Dependencies**:
```go
github.com/robfig/cron/v3
```

**Job Types**:
1. **Resource Sync Job**: Sync resources from cloud providers
2. **Drift Detection Job**: Periodic drift detection
3. **Vulnerability Scan Job**: Scheduled vulnerability scans
4. **Cost Sync Job**: Fetch cost data from cloud providers
5. **IaC Scan Job**: Scan IaC files on changes
6. **Compliance Assessment Job**: Periodic compliance checks
7. **Recommendation Generation Job**: Generate AI recommendations
8. **Anomaly Detection Job**: Detect cost/usage anomalies

**Files to Create**:
```
internal/worker/
├── scheduler.go        - Cron scheduler
├── jobs.go             - Job definitions
├── executor.go         - Job execution engine
├── registry.go         - Job registry
└── types.go
```

**Database Schema Addition**:
```sql
CREATE TABLE scheduled_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    job_type VARCHAR(100) NOT NULL,
    schedule VARCHAR(100) NOT NULL,  -- Cron expression
    is_enabled BOOLEAN DEFAULT true,
    config JSONB,
    last_run TIMESTAMP,
    next_run TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE job_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID NOT NULL REFERENCES scheduled_jobs(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    job_type VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL,  -- pending, running, completed, failed
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    duration_ms INT,
    result JSONB,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_scheduled_jobs_user_id ON scheduled_jobs(user_id);
CREATE INDEX idx_job_executions_job_id ON job_executions(job_id);
CREATE INDEX idx_job_executions_user_id ON job_executions(user_id);
CREATE INDEX idx_job_executions_status ON job_executions(status);
```

**Scheduler Configuration**:
```go
// Default schedules
ResourceSyncJob:          "0 */6 * * *"    // Every 6 hours
DriftDetectionJob:        "0 */4 * * *"    // Every 4 hours
VulnerabilityScanJob:     "0 2 * * *"      // Daily at 2 AM
CostSyncJob:              "0 1 * * *"      // Daily at 1 AM
ComplianceAssessmentJob:  "0 3 * * 0"      // Weekly on Sunday at 3 AM
RecommendationJob:        "0 4 * * *"      // Daily at 4 AM
AnomalyDetectionJob:      "0 */12 * * *"   // Every 12 hours
```

**API Endpoints**:
```
GET    /api/v1/jobs                    - List scheduled jobs
POST   /api/v1/jobs                    - Create scheduled job
GET    /api/v1/jobs/{id}               - Get job details
PUT    /api/v1/jobs/{id}               - Update job schedule
DELETE /api/v1/jobs/{id}               - Delete scheduled job
POST   /api/v1/jobs/{id}/run           - Trigger job manually
GET    /api/v1/jobs/{id}/executions    - List job executions
```

**Success Criteria**:
- ✅ Cron-based job scheduler
- ✅ Job registry and execution engine
- ✅ Job status tracking
- ✅ Retry mechanism with backoff
- ✅ API for job management

---

#### 5.2 Auto-Remediation Engine
**Location**: `internal/remediation/`

**Features**:
- Automated remediation workflows for common issues
- IaC-based remediation (generate Terraform PR)
- Cloud API-based remediation (apply fix directly)
- Policy-driven remediation (AWS Config, GCP Org Policies, Azure Policies)
- Human-in-the-loop approval for critical changes
- Rollback capability

**Remediation Strategies**:

**1. IaC Pull Request Generation**:
```
Issue: S3 bucket encryption disabled
Action: Generate Terraform PR to enable encryption
Steps:
  1. Parse Terraform file
  2. Add encryption block to s3_bucket resource
  3. Create git branch
  4. Commit changes
  5. Create pull request (GitHub/GitLab API)
  6. Notify user
```

**2. Direct Cloud API Remediation**:
```
Issue: Security group allows 0.0.0.0/0 on port 22
Action: Revoke security group rule
Steps:
  1. Identify security group ID
  2. Call AWS API to revoke ingress rule
  3. Verify rule removed
  4. Update resource configuration
  5. Create audit log
  6. Notify user
```

**3. Policy Enforcement**:
```
Issue: Unencrypted EBS volume created
Action: Enable AWS Config rule to auto-remediate
Steps:
  1. Create AWS Config rule
  2. Attach auto-remediation action (Lambda)
  3. Lambda encrypts new volumes automatically
  4. Monitor compliance
```

**Files to Create**:
```
internal/remediation/
├── engine.go           - Remediation orchestrator
├── strategies/
│   ├── iac_pr.go       - IaC pull request generation
│   ├── cloud_api.go    - Direct cloud API fixes
│   ├── policy.go       - Policy enforcement
│   └── manual.go       - Manual remediation steps
├── git/
│   ├── github.go       - GitHub API integration
│   └── gitlab.go       - GitLab API integration
└── types.go
```

**Database Schema Addition**:
```sql
CREATE TABLE remediation_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    drift_id UUID REFERENCES drifts(id),
    vulnerability_id UUID REFERENCES vulnerabilities(id),
    remediation_type VARCHAR(50) NOT NULL,  -- iac_pr, cloud_api, policy, manual
    status VARCHAR(50) NOT NULL,            -- pending, approved, in_progress, completed, failed
    strategy JSONB NOT NULL,
    approval_required BOOLEAN DEFAULT false,
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    result JSONB,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_remediation_actions_user_id ON remediation_actions(user_id);
CREATE INDEX idx_remediation_actions_status ON remediation_actions(status);
```

**API Endpoints**:
```
POST   /api/v1/remediation/suggest             - Get remediation suggestions
POST   /api/v1/remediation/execute             - Execute remediation
GET    /api/v1/remediation/actions             - List remediation actions
GET    /api/v1/remediation/actions/{id}        - Get action details
POST   /api/v1/remediation/actions/{id}/approve - Approve remediation
POST   /api/v1/remediation/actions/{id}/rollback - Rollback remediation
```

**Success Criteria**:
- ✅ Automated remediation for common issues
- ✅ IaC PR generation for Terraform
- ✅ Cloud API-based fixes
- ✅ Human approval workflow
- ✅ Rollback capability
- ✅ Audit logging

---

#### 5.3 Event-Driven Remediation
**Location**: `internal/remediation/event_driven.go`

**Features**:
- Real-time remediation using cloud-native event systems
- AWS: CloudWatch Events → Lambda → Remediation
- GCP: Cloud Functions → Pub/Sub → Remediation
- Azure: Event Grid → Azure Functions → Remediation
- Webhook-based notifications

**Architecture**:
```
Cloud Provider Event → Webhook → InfraAudit → Remediation Action
```

**Example Flow**:
```
1. New EC2 instance created with public IP
2. CloudWatch Event triggered
3. Lambda forwards event to InfraAudit webhook
4. InfraAudit analyzes security posture
5. Detects security group allows 0.0.0.0/0
6. Auto-generates remediation action
7. Applies fix (if auto-approved)
8. Notifies user
```

**Files to Create**:
```
internal/remediation/event_driven.go
internal/api/handlers/webhook_handler.go
```

**API Endpoints**:
```
POST   /api/v1/webhooks/aws               - AWS event webhook
POST   /api/v1/webhooks/gcp               - GCP event webhook
POST   /api/v1/webhooks/azure             - Azure event webhook
```

**Success Criteria**:
- ✅ Event-driven remediation pipeline
- ✅ Webhook endpoints for cloud events
- ✅ Real-time remediation execution
- ✅ Event processing with deduplication

---

### Phase 5 Deliverables
- [ ] Cron-based job scheduler
- [ ] Job registry and execution engine
- [ ] Auto-remediation engine
- [ ] IaC PR generation
- [ ] Cloud API remediation
- [ ] Policy enforcement
- [ ] Event-driven remediation
- [ ] Database schema for jobs and remediation
- [ ] API endpoints for automation
- [ ] Unit tests and integration tests

**Estimated Time**: 2-3 weeks
**Priority**: Medium
**Dependencies**: All previous phases

---

## Phase 6: Notifications & Integrations

### Objective
Complete the notification system to deliver alerts, reports, and recommendations via multiple channels (Slack, email, webhooks).

### Components to Build

#### 6.1 Slack Integration
**Location**: `internal/integrations/slack/client.go`

**Features**:
- Send alerts to Slack channels
- Send compliance reports
- Send daily/weekly summary reports
- Interactive buttons for approvals
- Rich message formatting with attachments

**Dependencies**:
```go
github.com/slack-go/slack
```

**Message Types**:
1. **Critical Alert**: Drift detected with critical severity
2. **Vulnerability Alert**: High/critical vulnerabilities found
3. **Cost Alert**: Cost anomaly detected (spike > 20%)
4. **Compliance Alert**: Compliance assessment failed
5. **Remediation Approval**: Request approval for remediation
6. **Daily Summary**: Daily infrastructure health report

**Files to Create**:
```
internal/integrations/slack/
├── client.go           - Slack API client
├── messages.go         - Message templates
└── types.go
```

**API Endpoints**:
```
POST   /api/v1/integrations/slack/connect     - Connect Slack workspace
POST   /api/v1/integrations/slack/test        - Send test message
```

**Success Criteria**:
- ✅ Send formatted messages to Slack
- ✅ Support for attachments and buttons
- ✅ Configuration per user
- ✅ Test integration endpoint

---

#### 6.2 Email Notifications
**Location**: `internal/integrations/email/`

**Features**:
- Send email alerts
- HTML email templates
- Digest emails (daily/weekly)
- Compliance reports via email
- Email preferences management

**Dependencies**:
```go
github.com/sendgrid/sendgrid-go  (or SMTP)
```

**Email Templates**:
1. **Alert Email**: Security/compliance alert
2. **Digest Email**: Daily/weekly summary
3. **Report Email**: Compliance/vulnerability report
4. **Remediation Approval**: Approval request

**Files to Create**:
```
internal/integrations/email/
├── client.go           - Email client
├── templates/          - HTML templates
│   ├── alert.html
│   ├── digest.html
│   └── report.html
└── types.go
```

**Database Schema Addition**:
```sql
CREATE TABLE notification_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    channel VARCHAR(50) NOT NULL,  -- slack, email, webhook
    is_enabled BOOLEAN DEFAULT true,
    config JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, channel)
);

CREATE TABLE notification_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    channel VARCHAR(50) NOT NULL,
    notification_type VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL,  -- sent, failed, pending
    payload JSONB,
    error_message TEXT,
    sent_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_notification_logs_user_id ON notification_logs(user_id);
CREATE INDEX idx_notification_logs_status ON notification_logs(status);
```

**API Endpoints**:
```
GET    /api/v1/notifications/preferences      - Get notification preferences
PUT    /api/v1/notifications/preferences      - Update preferences
GET    /api/v1/notifications/logs             - List notification logs
```

**Success Criteria**:
- ✅ Send email notifications
- ✅ HTML email templates
- ✅ User preferences management
- ✅ Notification logging

---

#### 6.3 Webhook Integration
**Location**: `internal/integrations/webhook/`

**Features**:
- Send events to external webhooks
- Retry with exponential backoff
- Signature verification (HMAC)
- Webhook management (create/test/delete)

**Webhook Events**:
- `drift.detected`
- `vulnerability.found`
- `compliance.failed`
- `cost.anomaly_detected`
- `remediation.completed`
- `scan.completed`

**Files to Create**:
```
internal/integrations/webhook/
├── client.go           - Webhook client
├── events.go           - Event definitions
└── types.go
```

**Database Schema Addition**:
```sql
CREATE TABLE webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    secret VARCHAR(255),  -- For HMAC signature
    events TEXT[] NOT NULL,  -- Array of subscribed events
    is_enabled BOOLEAN DEFAULT true,
    retry_config JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_webhooks_user_id ON webhooks(user_id);
```

**API Endpoints**:
```
GET    /api/v1/webhooks                - List webhooks
POST   /api/v1/webhooks                - Create webhook
GET    /api/v1/webhooks/{id}           - Get webhook details
PUT    /api/v1/webhooks/{id}           - Update webhook
DELETE /api/v1/webhooks/{id}           - Delete webhook
POST   /api/v1/webhooks/{id}/test      - Test webhook
```

**Success Criteria**:
- ✅ Send events to webhooks
- ✅ Retry mechanism
- ✅ HMAC signature verification
- ✅ Webhook CRUD operations

---

#### 6.4 Notification Service
**Location**: `internal/services/notification_service.go`

**Features**:
- Unified notification API
- Route notifications to appropriate channels
- Respect user preferences
- Batch notifications (digest mode)
- Priority-based notification routing

**Notification Routing Logic**:
```
Notification Priority:
- Critical: Slack + Email (immediate)
- High: Slack or Email (immediate)
- Medium: Email (digest)
- Low: Email (weekly digest)
```

**API Endpoints**:
```
POST   /api/v1/notifications/send      - Send notification
GET    /api/v1/notifications/history   - Notification history
```

**Success Criteria**:
- ✅ Unified notification service
- ✅ Multi-channel routing
- ✅ User preferences respected
- ✅ Digest mode support

---

### Phase 6 Deliverables
- [ ] Slack integration (complete)
- [ ] Email notification system
- [ ] Webhook integration
- [ ] Notification service
- [ ] Database schema for notifications
- [ ] API endpoints for integrations
- [ ] User preferences management
- [ ] Unit tests and integration tests

**Estimated Time**: 1 week
**Priority**: Low
**Dependencies**: Phase 5 (for alert generation)

---

## Implementation Timeline

### Sprint Breakdown (8-12 weeks)

**Weeks 1-3: Phase 1 - IaC Parsing & Baseline Engine**
- Week 1: Terraform parser + database schema
- Week 2: CloudFormation + Kubernetes parsers
- Week 3: IaC drift detector + API endpoints + testing

**Weeks 4-5: Phase 2 - Security Scanner Integration**
- Week 4: Checkov + tfsec integration
- Week 5: Trivy extension + orchestration + testing

**Weeks 6-7: Phase 3 - Cloud Cost Analytics**
- Week 6: AWS + GCP cost integration
- Week 7: Azure cost + analytics service + testing

**Weeks 8-9: Phase 4 - Compliance Framework**
- Week 8: Framework definitions + mapping engine
- Week 9: Compliance service + reports + testing

**Weeks 10-11: Phase 5 - Automation & Orchestration**
- Week 10: Job scheduler + auto-remediation engine
- Week 11: Event-driven remediation + testing

**Week 12: Phase 6 - Notifications & Integrations**
- Week 12: Slack + email + webhook integration + testing

---

## Testing Strategy

### Unit Tests
- All parsers (IaC, cost, compliance)
- All scanners (Checkov, tfsec, Trivy)
- Drift detection algorithms
- Remediation strategies
- Notification routing logic

**Target Coverage**: 80%+

### Integration Tests
- End-to-end IaC scanning workflow
- Multi-cloud resource sync
- Drift detection with real resources
- Remediation execution
- Notification delivery

### Performance Tests
- Parse large Terraform files (>1000 resources)
- Scan multiple cloud accounts concurrently
- Handle 10,000+ resources per user
- Job scheduler under load

---

## Deployment Strategy

### Infrastructure Requirements
- **Compute**: 2-4 CPU cores, 4-8 GB RAM
- **Storage**: 20 GB minimum (for scanner binaries and cache)
- **Database**: PostgreSQL 14+ with 10 GB storage
- **Network**: Outbound access to cloud APIs

### External Dependencies
- **Checkov**: Install via pip3
- **tfsec**: Install binary (Go)
- **Trivy**: Already integrated
- **Git**: For IaC repository cloning

### Configuration
```env
# IaC Scanner Paths
CHECKOV_PATH=/usr/local/bin/checkov
TFSEC_PATH=/usr/local/bin/tfsec
TRIVY_PATH=/usr/local/bin/trivy

# Git Integration
GIT_ENABLED=true
GITHUB_TOKEN=<github_personal_access_token>
GITLAB_TOKEN=<gitlab_personal_access_token>

# Notification Channels
SLACK_WEBHOOK_URL=<slack_webhook_url>
SENDGRID_API_KEY=<sendgrid_api_key>
EMAIL_FROM=noreply@infraaudit.com

# Job Scheduler
ENABLE_SCHEDULER=true
DEFAULT_RESOURCE_SYNC_CRON=0 */6 * * *
DEFAULT_DRIFT_DETECTION_CRON=0 */4 * * *
```

### Docker Deployment
```dockerfile
FROM golang:1.24-alpine AS builder
# Install scanner dependencies
RUN apk add --no-cache git python3 py3-pip
RUN pip3 install checkov
RUN wget -O /usr/local/bin/tfsec https://github.com/aquasecurity/tfsec/releases/latest/download/tfsec-linux-amd64
# Build Go application
COPY . /app
WORKDIR /app
RUN go build -o /bin/infraaudit ./cmd/api

FROM alpine:latest
RUN apk add --no-cache ca-certificates python3 py3-pip git
COPY --from=builder /usr/local/bin/checkov /usr/local/bin/checkov
COPY --from=builder /usr/local/bin/tfsec /usr/local/bin/tfsec
COPY --from=builder /bin/infraaudit /bin/infraaudit
ENTRYPOINT ["/bin/infraaudit"]
```

---

## Success Metrics

### Feature Adoption
- [ ] 90% of users connect at least one cloud provider
- [ ] 70% of users upload IaC definitions
- [ ] 50% of users run scheduled scans
- [ ] 30% of users enable auto-remediation

### Performance
- [ ] IaC parsing: <5 seconds for 500 resources
- [ ] Drift detection: <30 seconds for 1000 resources
- [ ] Scanner orchestration: <2 minutes per IaC file
- [ ] Cost sync: <1 minute per cloud account

### Quality
- [ ] 80%+ unit test coverage
- [ ] 50%+ integration test coverage
- [ ] <1% false positive rate for drift detection
- [ ] <5% false positive rate for vulnerability scanning

### User Satisfaction
- [ ] Average session time: >10 minutes
- [ ] Monthly active users: 80%+ of registered users
- [ ] Net Promoter Score (NPS): >40

---

## Risk Mitigation

### Technical Risks

**Risk**: IaC parsing failures for complex Terraform modules
**Mitigation**: Graceful error handling, partial parsing support, user feedback collection

**Risk**: Scanner binary dependencies (Checkov, tfsec) not available
**Mitigation**: Check binary availability at startup, provide clear installation instructions, support Docker-based deployment

**Risk**: Cloud API rate limiting
**Mitigation**: Implement exponential backoff, request throttling, multi-account cost distribution

**Risk**: Large infrastructure scales (10,000+ resources)
**Mitigation**: Pagination, concurrent processing, database query optimization, caching

### Security Risks

**Risk**: IaC files contain secrets
**Mitigation**: Secret scanning before parsing, warning messages, redaction in logs

**Risk**: Auto-remediation causes outages
**Mitigation**: Human-in-the-loop approval for critical changes, rollback capability, audit logging

**Risk**: Webhook secrets exposed
**Mitigation**: HMAC signature verification, encrypted secret storage, rate limiting

### Operational Risks

**Risk**: Scheduler job failures accumulate
**Mitigation**: Dead letter queue, alerting on failed jobs, manual retry capability

**Risk**: Notification spam
**Mitigation**: Rate limiting, digest mode, user preferences

---

## Next Steps

1. **Review this plan** with stakeholders
2. **Prioritize phases** based on business value
3. **Assign resources** (developers, timeline)
4. **Create GitHub issues** for each phase
5. **Set up project board** (Kanban/Scrum)
6. **Begin Phase 1 implementation** (IaC parsing)

---

## Appendix

### Related Documentation
- [IMPLEMENTATION_SUMMARY.md](./IMPLEMENTATION_SUMMARY.md) - Existing drift detection implementation
- [README.md](./README.md) - Project overview
- [Database Migrations](./migrations/) - Database schema

### External Resources
- [Checkov Documentation](https://www.checkov.io/)
- [tfsec Documentation](https://aquasecurity.github.io/tfsec/)
- [Trivy Documentation](https://aquasecurity.github.io/trivy/)
- [AWS Cost Explorer API](https://docs.aws.amazon.com/aws-cost-management/latest/APIReference/API_Operations_AWS_Cost_Explorer_Service.html)
- [GCP Cloud Billing API](https://cloud.google.com/billing/docs/reference/rest)
- [Azure Cost Management API](https://docs.microsoft.com/en-us/rest/api/cost-management/)
- [CIS Benchmarks](https://www.cisecurity.org/cis-benchmarks)
- [NIST 800-53](https://nvlpubs.nist.gov/nistpubs/SpecialPublications/NIST.SP.800-53r5.pdf)

---

**Document Version**: 1.0
**Last Updated**: 2025-11-04
**Next Review**: After Phase 1 completion
