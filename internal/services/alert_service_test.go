package services

import (
	"context"
	"testing"

	"github.com/pratik-mahalle/infraudit/internal/domain/alert"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
	"github.com/pratik-mahalle/infraudit/internal/testutil"
)

func TestAlertService_Create(t *testing.T) {
	mockRepo := testutil.NewMockAlertRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := NewAlertService(mockRepo, log)

	tests := []struct {
		name    string
		alert   *alert.Alert
		wantErr bool
	}{
		{
			name: "create security alert",
			alert: &alert.Alert{
				UserID:      1,
				Type:        "security",
				Severity:    alert.SeverityCritical,
				Title:       "Unencrypted S3 Bucket",
				Description: "S3 bucket without encryption detected",
				Resource:    "my-bucket",
				Status:      alert.StatusOpen,
			},
			wantErr: false,
		},
		{
			name: "create cost alert",
			alert: &alert.Alert{
				UserID:      1,
				Type:        "cost",
				Severity:    alert.SeverityHigh,
				Title:       "Cost Spike Detected",
				Description: "Unusual cost increase in EC2",
				Resource:    "i-1234567890",
				Status:      alert.StatusOpen,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			id, err := service.Create(ctx, tt.alert)

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

func TestAlertService_GetByID(t *testing.T) {
	mockRepo := testutil.NewMockAlertRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := NewAlertService(mockRepo, log)

	ctx := context.Background()
	a := &alert.Alert{
		UserID:   1,
		Type:     "security",
		Severity: alert.SeverityCritical,
		Title:    "Test Alert",
	}
	id, _ := service.Create(ctx, a)

	tests := []struct {
		name    string
		userID  int64
		alertID int64
		wantErr bool
	}{
		{
			name:    "get existing alert",
			userID:  1,
			alertID: id,
			wantErr: false,
		},
		{
			name:    "get non-existing alert",
			userID:  1,
			alertID: 999,
			wantErr: true,
		},
		{
			name:    "get alert for different user",
			userID:  2,
			alertID: id,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := service.GetByID(ctx, tt.userID, tt.alertID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && a == nil {
				t.Error("GetByID() returned nil alert")
			}
		})
	}
}

func TestAlertService_UpdateStatus(t *testing.T) {
	mockRepo := testutil.NewMockAlertRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := NewAlertService(mockRepo, log)

	ctx := context.Background()
	a := &alert.Alert{
		UserID:   1,
		Type:     "security",
		Severity: alert.SeverityCritical,
		Title:    "Test Alert",
		Status:   alert.StatusOpen,
	}
	id, _ := service.Create(ctx, a)

	// Update status
	err := service.UpdateStatus(ctx, 1, id, alert.StatusResolved)
	if err != nil {
		t.Errorf("UpdateStatus() error = %v", err)
	}

	// Verify update
	updated, _ := service.GetByID(ctx, 1, id)
	if updated.Status != alert.StatusResolved {
		t.Errorf("UpdateStatus() status = %v, want %v", updated.Status, alert.StatusResolved)
	}
}

func TestAlertService_GetSummary(t *testing.T) {
	mockRepo := testutil.NewMockAlertRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := NewAlertService(mockRepo, log)

	ctx := context.Background()

	// Create alerts with different statuses
	alerts := []*alert.Alert{
		{UserID: 1, Type: "security", Severity: alert.SeverityCritical, Title: "Alert 1", Status: alert.StatusOpen},
		{UserID: 1, Type: "security", Severity: alert.SeverityHigh, Title: "Alert 2", Status: alert.StatusOpen},
		{UserID: 1, Type: "cost", Severity: alert.SeverityMedium, Title: "Alert 3", Status: alert.StatusResolved},
	}

	for _, a := range alerts {
		service.Create(ctx, a)
	}

	summary, err := service.GetSummary(ctx, 1)
	if err != nil {
		t.Errorf("GetSummary() error = %v", err)
		return
	}

	if summary[alert.StatusOpen] != 2 {
		t.Errorf("GetSummary() open alerts = %v, want %v", summary[alert.StatusOpen], 2)
	}

	if summary[alert.StatusResolved] != 1 {
		t.Errorf("GetSummary() resolved alerts = %v, want %v", summary[alert.StatusResolved], 1)
	}
}

func TestAlertService_Delete(t *testing.T) {
	mockRepo := testutil.NewMockAlertRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := NewAlertService(mockRepo, log)

	ctx := context.Background()
	a := &alert.Alert{
		UserID:   1,
		Type:     "security",
		Severity: alert.SeverityCritical,
		Title:    "To Delete",
	}
	id, _ := service.Create(ctx, a)

	// Delete alert
	err := service.Delete(ctx, 1, id)
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// Verify deletion
	_, err = service.GetByID(ctx, 1, id)
	if err == nil {
		t.Error("Delete() alert still exists after deletion")
	}
}
