package services

import (
	"context"
	"testing"

	"github.com/pratik-mahalle/infraudit/internal/domain/resource"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
	"github.com/pratik-mahalle/infraudit/internal/testutil"
)

func TestResourceService_Create(t *testing.T) {
	mockRepo := testutil.NewMockResourceRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := NewResourceService(mockRepo, log)

	tests := []struct {
		name     string
		resource *resource.Resource
		wantErr  bool
	}{
		{
			name: "create EC2 instance",
			resource: &resource.Resource{
				UserID:     1,
				Provider:   "aws",
				ResourceID: "i-1234567890",
				Name:       "web-server",
				Type:       "ec2_instance",
				Region:     "us-east-1",
				Status:     "running",
			},
			wantErr: false,
		},
		{
			name: "create S3 bucket",
			resource: &resource.Resource{
				UserID:     1,
				Provider:   "aws",
				ResourceID: "my-bucket",
				Name:       "my-bucket",
				Type:       "s3_bucket",
				Region:     "us-west-2",
				Status:     "active",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := service.Create(ctx, tt.resource)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResourceService_GetByID(t *testing.T) {
	mockRepo := testutil.NewMockResourceRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := NewResourceService(mockRepo, log)

	ctx := context.Background()
	res := &resource.Resource{
		UserID:     1,
		Provider:   "aws",
		ResourceID: "i-test",
		Name:       "test-instance",
		Type:       "ec2_instance",
		Region:     "us-east-1",
		Status:     "running",
	}
	service.Create(ctx, res)

	tests := []struct {
		name       string
		userID     int64
		resourceID string
		wantErr    bool
	}{
		{
			name:       "get existing resource",
			userID:     1,
			resourceID: "i-test",
			wantErr:    false,
		},
		{
			name:       "get non-existing resource",
			userID:     1,
			resourceID: "i-nonexistent",
			wantErr:    true,
		},
		{
			name:       "get resource for different user",
			userID:     2,
			resourceID: "i-test",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := service.GetByID(ctx, tt.userID, tt.resourceID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && r == nil {
				t.Error("GetByID() returned nil resource")
			}
		})
	}
}

func TestResourceService_List(t *testing.T) {
	mockRepo := testutil.NewMockResourceRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := NewResourceService(mockRepo, log)

	ctx := context.Background()

	// Create multiple resources
	resources := []*resource.Resource{
		{
			UserID:     1,
			Provider:   "aws",
			ResourceID: "i-1",
			Name:       "instance-1",
			Type:       "ec2_instance",
			Region:     "us-east-1",
			Status:     "running",
		},
		{
			UserID:     1,
			Provider:   "aws",
			ResourceID: "i-2",
			Name:       "instance-2",
			Type:       "ec2_instance",
			Region:     "us-west-2",
			Status:     "stopped",
		},
		{
			UserID:     2,
			Provider:   "gcp",
			ResourceID: "vm-1",
			Name:       "vm-1",
			Type:       "gce_instance",
			Region:     "us-central1",
			Status:     "running",
		},
	}

	for _, r := range resources {
		service.Create(ctx, r)
	}

	tests := []struct {
		name      string
		userID    int64
		filter    resource.Filter
		wantCount int
	}{
		{
			name:      "list all resources for user 1",
			userID:    1,
			filter:    resource.Filter{},
			wantCount: 2,
		},
		{
			name:      "list all resources for user 2",
			userID:    2,
			filter:    resource.Filter{},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resources, total, err := service.List(ctx, tt.userID, tt.filter, 10, 0)

			if err != nil {
				t.Errorf("List() error = %v", err)
				return
			}

			if len(resources) != tt.wantCount {
				t.Errorf("List() count = %v, want %v", len(resources), tt.wantCount)
			}

			if total != int64(tt.wantCount) {
				t.Errorf("List() total = %v, want %v", total, tt.wantCount)
			}
		})
	}
}

func TestResourceService_Update(t *testing.T) {
	mockRepo := testutil.NewMockResourceRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := NewResourceService(mockRepo, log)

	ctx := context.Background()
	res := &resource.Resource{
		UserID:     1,
		Provider:   "aws",
		ResourceID: "i-test",
		Name:       "old-name",
		Type:       "ec2_instance",
		Region:     "us-east-1",
		Status:     "running",
	}
	service.Create(ctx, res)

	updates := map[string]interface{}{
		"name":   "new-name",
		"status": "stopped",
	}

	err := service.Update(ctx, 1, "i-test", updates)
	if err != nil {
		t.Errorf("Update() error = %v", err)
	}

	// Verify update
	updated, _ := service.GetByID(ctx, 1, "i-test")
	if updated.Name != "new-name" {
		t.Errorf("Update() name = %v, want %v", updated.Name, "new-name")
	}
	if updated.Status != "stopped" {
		t.Errorf("Update() status = %v, want %v", updated.Status, "stopped")
	}
}

func TestResourceService_Delete(t *testing.T) {
	mockRepo := testutil.NewMockResourceRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := NewResourceService(mockRepo, log)

	ctx := context.Background()
	res := &resource.Resource{
		UserID:     1,
		Provider:   "aws",
		ResourceID: "i-delete",
		Name:       "to-delete",
		Type:       "ec2_instance",
		Region:     "us-east-1",
		Status:     "running",
	}
	service.Create(ctx, res)

	// Delete resource
	err := service.Delete(ctx, 1, "i-delete")
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// Verify deletion
	_, err = service.GetByID(ctx, 1, "i-delete")
	if err == nil {
		t.Error("Delete() resource still exists after deletion")
	}
}
