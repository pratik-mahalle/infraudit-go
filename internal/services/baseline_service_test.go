package services

import (
	"context"
	"testing"

	"github.com/pratik-mahalle/infraudit/internal/domain/baseline"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
	"github.com/pratik-mahalle/infraudit/internal/testutil"
)

func TestBaselineService_CreateBaseline(t *testing.T) {
	mockRepo := testutil.NewMockBaselineRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := NewBaselineService(mockRepo, log)

	tests := []struct {
		name     string
		baseline *baseline.Baseline
		wantErr  bool
	}{
		{
			name: "create manual baseline",
			baseline: &baseline.Baseline{
				UserID:        1,
				ResourceID:    "sg-12345",
				Provider:      "aws",
				ResourceType:  "security_group",
				Configuration: `{"ingress": [{"from_port": 443}]}`,
				BaselineType:  baseline.TypeManual,
				Description:   "Manual baseline for security group",
			},
			wantErr: false,
		},
		{
			name: "create approved baseline",
			baseline: &baseline.Baseline{
				UserID:        1,
				ResourceID:    "bucket-123",
				Provider:      "aws",
				ResourceType:  "s3_bucket",
				Configuration: `{"versioning": true, "encryption": true}`,
				BaselineType:  baseline.TypeApproved,
				Description:   "Approved secure bucket configuration",
			},
			wantErr: false,
		},
		{
			name: "create baseline with default type",
			baseline: &baseline.Baseline{
				UserID:        1,
				ResourceID:    "vm-123",
				Provider:      "gcp",
				ResourceType:  "compute_instance",
				Configuration: `{"machine_type": "n1-standard-1"}`,
				// BaselineType empty - should default to TypeManual
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			id, err := service.CreateBaseline(ctx, tt.baseline)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateBaseline() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && id == 0 {
				t.Error("CreateBaseline() returned 0 id")
			}

			// Verify default type was set if empty
			if tt.baseline.BaselineType == "" {
				// The function modifies the baseline in place, so check the passed baseline
				if tt.baseline.BaselineType != baseline.TypeManual {
					t.Errorf("CreateBaseline() did not set default type, got %v, want %v", tt.baseline.BaselineType, baseline.TypeManual)
				}
			}
		})
	}
}

func TestBaselineService_GetBaseline(t *testing.T) {
	mockRepo := testutil.NewMockBaselineRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := NewBaselineService(mockRepo, log)

	ctx := context.Background()
	b := &baseline.Baseline{
		UserID:        1,
		ResourceID:    "test-resource",
		Provider:      "aws",
		ResourceType:  "ec2_instance",
		Configuration: `{"instance_type": "t2.micro"}`,
		BaselineType:  baseline.TypeApproved,
	}
	service.CreateBaseline(ctx, b)

	tests := []struct {
		name         string
		userID       int64
		resourceID   string
		baselineType string
		wantErr      bool
	}{
		{
			name:         "get existing baseline",
			userID:       1,
			resourceID:   "test-resource",
			baselineType: baseline.TypeApproved,
			wantErr:      false,
		},
		{
			name:         "get non-existing baseline",
			userID:       1,
			resourceID:   "non-existent",
			baselineType: baseline.TypeApproved,
			wantErr:      true,
		},
		{
			name:         "get baseline for different user",
			userID:       2,
			resourceID:   "test-resource",
			baselineType: baseline.TypeApproved,
			wantErr:      true,
		},
		{
			name:         "get baseline with wrong type",
			userID:       1,
			resourceID:   "test-resource",
			baselineType: baseline.TypeManual, // Created with TypeApproved
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := service.GetBaseline(ctx, tt.userID, tt.resourceID, tt.baselineType)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetBaseline() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && b == nil {
				t.Error("GetBaseline() returned nil baseline")
			}
		})
	}
}

func TestBaselineService_DeleteBaseline(t *testing.T) {
	mockRepo := testutil.NewMockBaselineRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := NewBaselineService(mockRepo, log)

	ctx := context.Background()
	b := &baseline.Baseline{
		UserID:        1,
		ResourceID:    "to-delete",
		Provider:      "aws",
		ResourceType:  "security_group",
		Configuration: `{"rules": []}`,
		BaselineType:  baseline.TypeManual,
	}
	id, _ := service.CreateBaseline(ctx, b)

	// Delete baseline
	err := service.DeleteBaseline(ctx, 1, id)
	if err != nil {
		t.Errorf("DeleteBaseline() error = %v", err)
	}

	// Verify deletion
	_, err = service.GetBaseline(ctx, 1, "to-delete", baseline.TypeManual)
	if err == nil {
		t.Error("DeleteBaseline() baseline still exists after deletion")
	}
}

func TestBaselineService_ListBaselines(t *testing.T) {
	mockRepo := testutil.NewMockBaselineRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := NewBaselineService(mockRepo, log)

	ctx := context.Background()

	// Create baselines for user 1
	baselines := []*baseline.Baseline{
		{UserID: 1, ResourceID: "res-1", Provider: "aws", ResourceType: "ec2", Configuration: "{}", BaselineType: baseline.TypeManual},
		{UserID: 1, ResourceID: "res-2", Provider: "aws", ResourceType: "s3", Configuration: "{}", BaselineType: baseline.TypeApproved},
		{UserID: 1, ResourceID: "res-3", Provider: "gcp", ResourceType: "vm", Configuration: "{}", BaselineType: baseline.TypeAutomatic},
		{UserID: 2, ResourceID: "res-4", Provider: "azure", ResourceType: "vm", Configuration: "{}", BaselineType: baseline.TypeManual}, // Different user
	}

	for _, b := range baselines {
		service.CreateBaseline(ctx, b)
	}

	// List baselines for user 1
	result, err := service.ListBaselines(ctx, 1)
	if err != nil {
		t.Errorf("ListBaselines() error = %v", err)
		return
	}

	if len(result) != 3 {
		t.Errorf("ListBaselines() returned %v baselines, want %v", len(result), 3)
	}

	// List baselines for user 2
	result2, err := service.ListBaselines(ctx, 2)
	if err != nil {
		t.Errorf("ListBaselines() error = %v", err)
		return
	}

	if len(result2) != 1 {
		t.Errorf("ListBaselines() for user 2 returned %v baselines, want %v", len(result2), 1)
	}
}

func TestBaselineService_MultipleBaselineTypes(t *testing.T) {
	mockRepo := testutil.NewMockBaselineRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := NewBaselineService(mockRepo, log)

	ctx := context.Background()
	userID := int64(1)
	resourceID := "multi-type-resource"

	// Create baselines of different types for the same resource
	manualBaseline := &baseline.Baseline{
		UserID:        userID,
		ResourceID:    resourceID,
		Provider:      "aws",
		ResourceType:  "security_group",
		Configuration: `{"manual": true}`,
		BaselineType:  baseline.TypeManual,
	}
	service.CreateBaseline(ctx, manualBaseline)

	approvedBaseline := &baseline.Baseline{
		UserID:        userID,
		ResourceID:    resourceID,
		Provider:      "aws",
		ResourceType:  "security_group",
		Configuration: `{"approved": true}`,
		BaselineType:  baseline.TypeApproved,
	}
	service.CreateBaseline(ctx, approvedBaseline)

	// Retrieve each type
	manual, err := service.GetBaseline(ctx, userID, resourceID, baseline.TypeManual)
	if err != nil {
		t.Errorf("GetBaseline(manual) error = %v", err)
		return
	}
	if manual.Configuration != `{"manual": true}` {
		t.Errorf("GetBaseline(manual) got wrong config: %v", manual.Configuration)
	}

	approved, err := service.GetBaseline(ctx, userID, resourceID, baseline.TypeApproved)
	if err != nil {
		t.Errorf("GetBaseline(approved) error = %v", err)
		return
	}
	if approved.Configuration != `{"approved": true}` {
		t.Errorf("GetBaseline(approved) got wrong config: %v", approved.Configuration)
	}
}
