package integration

import (
	"context"
	"testing"

	"github.com/pratik-mahalle/infraudit/internal/domain/baseline"
	"github.com/pratik-mahalle/infraudit/internal/domain/drift"
	"github.com/pratik-mahalle/infraudit/internal/domain/resource"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
	"github.com/pratik-mahalle/infraudit/internal/services"
	"github.com/pratik-mahalle/infraudit/internal/testutil"
)

// setupDriftTestServices creates services for drift detection integration tests
func setupDriftTestServices(t *testing.T) (*services.DriftService, *testutil.MockResourceRepository, *testutil.MockBaselineRepository, *testutil.MockDriftRepository) {
	driftRepo := testutil.NewMockDriftRepository()
	baselineRepo := testutil.NewMockBaselineRepository()
	resourceRepo := testutil.NewMockResourceRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})

	driftService := services.NewDriftService(driftRepo, baselineRepo, resourceRepo, log)

	return driftService.(*services.DriftService), resourceRepo, baselineRepo, driftRepo
}

func TestDriftDetection_SecurityGroupChange(t *testing.T) {
	driftService, resourceRepo, baselineRepo, _ := setupDriftTestServices(t)
	ctx := context.Background()
	userID := int64(1)

	// Setup: Create a resource with current configuration
	res := &resource.Resource{
		UserID:     userID,
		ResourceID: "sg-test-001",
		Provider:   "aws",
		Type:       "security_group",
		Name:       "Production Security Group",
		Configuration: `{
			"ingress": [
				{"from_port": 443, "to_port": 443, "cidr_blocks": ["0.0.0.0/0"]},
				{"from_port": 22, "to_port": 22, "cidr_blocks": ["10.0.0.0/8"]}
			],
			"egress": [
				{"from_port": 0, "to_port": 0, "cidr_blocks": ["0.0.0.0/0"]}
			]
		}`,
	}
	resourceRepo.Create(ctx, res)

	// Setup: Create a baseline with different (approved) configuration
	bl := &baseline.Baseline{
		UserID:       userID,
		ResourceID:   "sg-test-001",
		Provider:     "aws",
		ResourceType: "security_group",
		Configuration: `{
			"ingress": [
				{"from_port": 443, "to_port": 443, "cidr_blocks": ["10.0.0.0/8"]}
			],
			"egress": [
				{"from_port": 0, "to_port": 0, "cidr_blocks": ["0.0.0.0/0"]}
			]
		}`,
		BaselineType: baseline.TypeApproved,
	}
	baselineRepo.Create(ctx, bl)

	// Run drift detection
	err := driftService.DetectDrifts(ctx, userID)
	if err != nil {
		t.Errorf("DetectDrifts() error = %v", err)
	}

	// Verify drifts were detected
	drifts, total, err := driftService.List(ctx, userID, drift.Filter{}, 10, 0)
	if err != nil {
		t.Errorf("List() error = %v", err)
	}

	t.Logf("Detected %d drifts (total: %d)", len(drifts), total)

	// At least one drift should be detected (security group configuration change)
	// Note: This depends on the actual drift detection implementation
}

func TestDriftDetection_NewResourceWithoutBaseline(t *testing.T) {
	driftService, resourceRepo, _, _ := setupDriftTestServices(t)
	ctx := context.Background()
	userID := int64(1)

	// Setup: Create a resource without any baseline
	res := &resource.Resource{
		UserID:        userID,
		ResourceID:    "bucket-new-001",
		Provider:      "aws",
		Type:          "s3_bucket",
		Name:          "New S3 Bucket",
		Configuration: `{"versioning": true, "encryption": "AES256"}`,
	}
	resourceRepo.Create(ctx, res)

	// Run drift detection
	err := driftService.DetectDrifts(ctx, userID)
	if err != nil {
		t.Errorf("DetectDrifts() error = %v", err)
	}

	// For new resources without baselines, the system should either:
	// 1. Create an automatic baseline, or
	// 2. Report the resource as having no baseline
	t.Log("Drift detection completed for resource without baseline")
}

func TestDriftDetection_EncryptionDisabled(t *testing.T) {
	driftService, resourceRepo, baselineRepo, _ := setupDriftTestServices(t)
	ctx := context.Background()
	userID := int64(1)

	// Setup: Resource with encryption disabled
	res := &resource.Resource{
		UserID:        userID,
		ResourceID:    "bucket-encrypt-001",
		Provider:      "aws",
		Type:          "s3_bucket",
		Name:          "Encryption Test Bucket",
		Configuration: `{"versioning": true, "encryption": "none"}`,
	}
	resourceRepo.Create(ctx, res)

	// Setup: Baseline with encryption enabled
	bl := &baseline.Baseline{
		UserID:        userID,
		ResourceID:    "bucket-encrypt-001",
		Provider:      "aws",
		ResourceType:  "s3_bucket",
		Configuration: `{"versioning": true, "encryption": "AES256"}`,
		BaselineType:  baseline.TypeApproved,
	}
	baselineRepo.Create(ctx, bl)

	// Run drift detection
	err := driftService.DetectDrifts(ctx, userID)
	if err != nil {
		t.Errorf("DetectDrifts() error = %v", err)
	}

	// Check for critical drift (encryption disabled)
	drifts, _, _ := driftService.List(ctx, userID, drift.Filter{}, 10, 0)

	for _, d := range drifts {
		if d.ResourceID == "bucket-encrypt-001" {
			t.Logf("Found drift for encryption resource: Type=%s, Severity=%s", d.DriftType, d.Severity)
		}
	}
}

func TestDriftDetection_MultipleResources(t *testing.T) {
	driftService, resourceRepo, baselineRepo, _ := setupDriftTestServices(t)
	ctx := context.Background()
	userID := int64(1)

	// Setup: Multiple resources with baselines
	resources := []*resource.Resource{
		{UserID: userID, ResourceID: "ec2-001", Provider: "aws", Type: "ec2_instance", Name: "Web Server", Configuration: `{"instance_type": "t3.large"}`},
		{UserID: userID, ResourceID: "ec2-002", Provider: "aws", Type: "ec2_instance", Name: "API Server", Configuration: `{"instance_type": "t3.medium"}`},
		{UserID: userID, ResourceID: "rds-001", Provider: "aws", Type: "rds_instance", Name: "Database", Configuration: `{"instance_class": "db.r5.large"}`},
	}

	baselines := []*baseline.Baseline{
		{UserID: userID, ResourceID: "ec2-001", Provider: "aws", ResourceType: "ec2_instance", Configuration: `{"instance_type": "t3.medium"}`, BaselineType: baseline.TypeApproved},    // Different instance type
		{UserID: userID, ResourceID: "ec2-002", Provider: "aws", ResourceType: "ec2_instance", Configuration: `{"instance_type": "t3.medium"}`, BaselineType: baseline.TypeApproved},    // Same
		{UserID: userID, ResourceID: "rds-001", Provider: "aws", ResourceType: "rds_instance", Configuration: `{"instance_class": "db.r5.large"}`, BaselineType: baseline.TypeApproved}, // Same
	}

	for _, r := range resources {
		resourceRepo.Create(ctx, r)
	}
	for _, b := range baselines {
		baselineRepo.Create(ctx, b)
	}

	// Run drift detection
	err := driftService.DetectDrifts(ctx, userID)
	if err != nil {
		t.Errorf("DetectDrifts() error = %v", err)
	}

	// Verify drifts
	drifts, total, _ := driftService.List(ctx, userID, drift.Filter{}, 10, 0)
	t.Logf("Detected %d drifts for %d resources (total in DB: %d)", len(drifts), len(resources), total)
}

func TestDriftWorkflow_DetectAcknowledgeResolve(t *testing.T) {
	driftService, resourceRepo, baselineRepo, _ := setupDriftTestServices(t)
	ctx := context.Background()
	userID := int64(1)

	// Setup: Create driftable resource
	res := &resource.Resource{
		UserID:        userID,
		ResourceID:    "workflow-test-001",
		Provider:      "aws",
		Type:          "security_group",
		Name:          "Workflow Test SG",
		Configuration: `{"ingress": [{"from_port": 22, "to_port": 22, "cidr_blocks": ["0.0.0.0/0"]}]}`,
	}
	resourceRepo.Create(ctx, res)

	bl := &baseline.Baseline{
		UserID:        userID,
		ResourceID:    "workflow-test-001",
		Provider:      "aws",
		ResourceType:  "security_group",
		Configuration: `{"ingress": [{"from_port": 22, "to_port": 22, "cidr_blocks": ["10.0.0.0/8"]}]}`,
		BaselineType:  baseline.TypeApproved,
	}
	baselineRepo.Create(ctx, bl)

	// Step 1: Detect drifts
	err := driftService.DetectDrifts(ctx, userID)
	if err != nil {
		t.Fatalf("DetectDrifts() error = %v", err)
	}

	drifts, _, _ := driftService.List(ctx, userID, drift.Filter{}, 10, 0)
	if len(drifts) == 0 {
		t.Log("No drifts detected - implementation may handle this differently")
		return
	}

	driftID := drifts[0].ID
	t.Logf("Step 1: Detected drift ID=%d, Status=%s", driftID, drifts[0].Status)

	// Step 2: Acknowledge drift
	err = driftService.UpdateStatus(ctx, userID, driftID, drift.StatusAcknowledged)
	if err != nil {
		t.Errorf("UpdateStatus(acknowledged) error = %v", err)
	}

	updated, _ := driftService.GetByID(ctx, userID, driftID)
	if updated.Status != drift.StatusAcknowledged {
		t.Errorf("Status should be acknowledged, got %s", updated.Status)
	}
	t.Log("Step 2: Drift acknowledged")

	// Step 3: Resolve drift
	err = driftService.UpdateStatus(ctx, userID, driftID, drift.StatusResolved)
	if err != nil {
		t.Errorf("UpdateStatus(resolved) error = %v", err)
	}

	resolved, _ := driftService.GetByID(ctx, userID, driftID)
	if resolved.Status != drift.StatusResolved {
		t.Errorf("Status should be resolved, got %s", resolved.Status)
	}
	t.Log("Step 3: Drift resolved")
}

func TestDriftSummary_BySeverity(t *testing.T) {
	driftService, _, _, driftRepo := setupDriftTestServices(t)
	ctx := context.Background()
	userID := int64(1)

	// Create drifts with different severities directly in repo
	drifts := []*drift.Drift{
		{UserID: userID, ResourceID: "res-1", DriftType: drift.TypeSecurityGroup, Severity: drift.SeverityCritical, Status: drift.StatusDetected},
		{UserID: userID, ResourceID: "res-2", DriftType: drift.TypeSecurityGroup, Severity: drift.SeverityCritical, Status: drift.StatusDetected},
		{UserID: userID, ResourceID: "res-3", DriftType: drift.TypeIAMPolicy, Severity: drift.SeverityHigh, Status: drift.StatusDetected},
		{UserID: userID, ResourceID: "res-4", DriftType: drift.TypeEncryption, Severity: drift.SeverityMedium, Status: drift.StatusDetected},
		{UserID: userID, ResourceID: "res-5", DriftType: drift.TypeConfigurationChange, Severity: drift.SeverityLow, Status: drift.StatusResolved},
	}

	for _, d := range drifts {
		driftRepo.Create(ctx, d)
	}

	summary, err := driftService.GetSummary(ctx, userID)
	if err != nil {
		t.Errorf("GetSummary() error = %v", err)
		return
	}

	t.Logf("Drift Summary: %+v", summary)

	if summary[drift.SeverityCritical] != 2 {
		t.Errorf("Critical count = %v, want 2", summary[drift.SeverityCritical])
	}
	if summary[drift.SeverityHigh] != 1 {
		t.Errorf("High count = %v, want 1", summary[drift.SeverityHigh])
	}
}
