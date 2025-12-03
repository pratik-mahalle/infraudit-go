package services

import (
	"context"
	"testing"

	"github.com/pratik-mahalle/infraudit/internal/domain/recommendation"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
	"github.com/pratik-mahalle/infraudit/internal/testutil"
)

func TestRecommendationService_Create(t *testing.T) {
	mockRepo := testutil.NewMockRecommendationRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := NewRecommendationService(mockRepo, nil, log)

	tests := []struct {
		name           string
		recommendation *recommendation.Recommendation
		wantErr        bool
	}{
		{
			name: "create cost optimization recommendation",
			recommendation: &recommendation.Recommendation{
				UserID:      1,
				Type:        "cost_optimization",
				Priority:    recommendation.PriorityHigh,
				Title:       "Resize Underutilized EC2",
				Description: "EC2 instance running at 10% CPU",
				Savings:     100.50,
				Effort:      recommendation.EffortLow,
				Impact:      recommendation.ImpactHigh,
				Category:    "compute",
				Resources:   []string{"i-1234567890"},
			},
			wantErr: false,
		},
		{
			name: "create security recommendation",
			recommendation: &recommendation.Recommendation{
				UserID:      1,
				Type:        "security",
				Priority:    recommendation.PriorityCritical,
				Title:       "Enable S3 Encryption",
				Description: "S3 bucket without encryption",
				Savings:     0,
				Effort:      recommendation.EffortLow,
				Impact:      recommendation.ImpactHigh,
				Category:    "storage",
				Resources:   []string{"my-bucket"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			id, err := service.Create(ctx, tt.recommendation)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && id == 0 {
				t.Error("Create() returned 0 id")
			}
		})
	}
}

func TestRecommendationService_GetByID(t *testing.T) {
	mockRepo := testutil.NewMockRecommendationRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := NewRecommendationService(mockRepo, nil, log)

	ctx := context.Background()
	rec := &recommendation.Recommendation{
		UserID:   1,
		Type:     "cost_optimization",
		Priority: recommendation.PriorityHigh,
		Title:    "Test Recommendation",
		Savings:  50.0,
	}
	id, _ := service.Create(ctx, rec)

	tests := []struct {
		name    string
		userID  int64
		recID   int64
		wantErr bool
	}{
		{
			name:    "get existing recommendation",
			userID:  1,
			recID:   id,
			wantErr: false,
		},
		{
			name:    "get non-existing recommendation",
			userID:  1,
			recID:   999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := service.GetByID(ctx, tt.userID, tt.recID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && r == nil {
				t.Error("GetByID() returned nil recommendation")
			}
		})
	}
}

func TestRecommendationService_GetTotalSavings(t *testing.T) {
	mockRepo := testutil.NewMockRecommendationRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := NewRecommendationService(mockRepo, nil, log)

	ctx := context.Background()

	// Create recommendations with different savings
	recommendations := []*recommendation.Recommendation{
		{UserID: 1, Type: "cost_optimization", Priority: recommendation.PriorityHigh, Title: "Rec 1", Savings: 100.0},
		{UserID: 1, Type: "cost_optimization", Priority: recommendation.PriorityMedium, Title: "Rec 2", Savings: 50.5},
		{UserID: 1, Type: "cost_optimization", Priority: recommendation.PriorityLow, Title: "Rec 3", Savings: 25.25},
		{UserID: 2, Type: "cost_optimization", Priority: recommendation.PriorityHigh, Title: "Rec 4", Savings: 200.0},
	}

	for _, r := range recommendations {
		service.Create(ctx, r)
	}

	totalSavings, err := service.GetTotalSavings(ctx, 1)
	if err != nil {
		t.Errorf("GetTotalSavings() error = %v", err)
		return
	}

	expectedSavings := 175.75
	if totalSavings != expectedSavings {
		t.Errorf("GetTotalSavings() = %v, want %v", totalSavings, expectedSavings)
	}
}

func TestRecommendationService_Update(t *testing.T) {
	mockRepo := testutil.NewMockRecommendationRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := NewRecommendationService(mockRepo, nil, log)

	ctx := context.Background()
	rec := &recommendation.Recommendation{
		UserID:   1,
		Type:     "cost_optimization",
		Priority: recommendation.PriorityMedium,
		Title:    "Old Title",
		Savings:  50.0,
	}
	id, _ := service.Create(ctx, rec)

	updates := map[string]interface{}{
		"title":    "New Title",
		"priority": recommendation.PriorityHigh,
		"savings":  100.0,
	}

	err := service.Update(ctx, 1, id, updates)
	if err != nil {
		t.Errorf("Update() error = %v", err)
	}

	// Verify update
	updated, _ := service.GetByID(ctx, 1, id)
	if updated.Title != "New Title" {
		t.Errorf("Update() title = %v, want %v", updated.Title, "New Title")
	}
	if updated.Priority != recommendation.PriorityHigh {
		t.Errorf("Update() priority = %v, want %v", updated.Priority, recommendation.PriorityHigh)
	}
	if updated.Savings != 100.0 {
		t.Errorf("Update() savings = %v, want %v", updated.Savings, 100.0)
	}
}

func TestRecommendationService_Delete(t *testing.T) {
	mockRepo := testutil.NewMockRecommendationRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := NewRecommendationService(mockRepo, nil, log)

	ctx := context.Background()
	rec := &recommendation.Recommendation{
		UserID:   1,
		Type:     "cost_optimization",
		Priority: recommendation.PriorityHigh,
		Title:    "To Delete",
		Savings:  50.0,
	}
	id, _ := service.Create(ctx, rec)

	// Delete recommendation
	err := service.Delete(ctx, 1, id)
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// Verify deletion
	_, err = service.GetByID(ctx, 1, id)
	if err == nil {
		t.Error("Delete() recommendation still exists after deletion")
	}
}
