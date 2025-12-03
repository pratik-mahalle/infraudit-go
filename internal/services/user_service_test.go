package services

import (
	"context"
	"testing"

	"github.com/pratik-mahalle/infraudit/internal/domain/user"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
	"github.com/pratik-mahalle/infraudit/internal/testutil"
)

func TestUserService_Create(t *testing.T) {
	mockRepo := testutil.NewMockUserRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := NewUserService(mockRepo, log)

	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "successful user creation",
			email:   "test@example.com",
			wantErr: false,
		},
		{
			name:    "create user with valid email",
			email:   "user@domain.com",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			u, err := service.Create(ctx, tt.email)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if u == nil {
					t.Error("Create() returned nil user")
					return
				}
				if u.Email != tt.email {
					t.Errorf("Create() email = %v, want %v", u.Email, tt.email)
				}
				if u.Role != user.RoleUser {
					t.Errorf("Create() role = %v, want %v", u.Role, user.RoleUser)
				}
				if u.PlanType != user.PlanTypeFree {
					t.Errorf("Create() plan_type = %v, want %v", u.PlanType, user.PlanTypeFree)
				}
			}
		})
	}
}

func TestUserService_GetByID(t *testing.T) {
	mockRepo := testutil.NewMockUserRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := NewUserService(mockRepo, log)

	// Create a user first
	ctx := context.Background()
	createdUser, _ := service.Create(ctx, "test@example.com")

	tests := []struct {
		name    string
		userID  int64
		wantErr bool
	}{
		{
			name:    "get existing user",
			userID:  createdUser.ID,
			wantErr: false,
		},
		{
			name:    "get non-existing user",
			userID:  999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := service.GetByID(ctx, tt.userID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && u == nil {
				t.Error("GetByID() returned nil user")
			}
		})
	}
}

func TestUserService_GetByEmail(t *testing.T) {
	mockRepo := testutil.NewMockUserRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := NewUserService(mockRepo, log)

	// Create a user first
	ctx := context.Background()
	email := "test@example.com"
	service.Create(ctx, email)

	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "get existing user by email",
			email:   email,
			wantErr: false,
		},
		{
			name:    "get non-existing user by email",
			email:   "nonexistent@example.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := service.GetByEmail(ctx, tt.email)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetByEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if u == nil {
					t.Error("GetByEmail() returned nil user")
					return
				}
				if u.Email != tt.email {
					t.Errorf("GetByEmail() email = %v, want %v", u.Email, tt.email)
				}
			}
		})
	}
}

func TestUserService_Update(t *testing.T) {
	mockRepo := testutil.NewMockUserRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := NewUserService(mockRepo, log)

	ctx := context.Background()
	u, _ := service.Create(ctx, "test@example.com")

	// Update user
	u.PlanType = user.PlanTypePro
	err := service.Update(ctx, u)

	if err != nil {
		t.Errorf("Update() error = %v", err)
	}

	// Verify update
	updated, _ := service.GetByID(ctx, u.ID)
	if updated.PlanType != user.PlanTypePro {
		t.Errorf("Update() plan_type = %v, want %v", updated.PlanType, user.PlanTypePro)
	}
}

func TestUserService_UpgradePlan(t *testing.T) {
	mockRepo := testutil.NewMockUserRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := NewUserService(mockRepo, log)

	ctx := context.Background()
	u, _ := service.Create(ctx, "test@example.com")

	tests := []struct {
		name     string
		userID   int64
		planType string
		wantErr  bool
	}{
		{
			name:     "upgrade to premium",
			userID:   u.ID,
			planType: user.PlanTypePro,
			wantErr:  false,
		},
		{
			name:     "upgrade to enterprise",
			userID:   u.ID,
			planType: user.PlanTypeEnterprise,
			wantErr:  false,
		},
		{
			name:     "upgrade non-existing user",
			userID:   999,
			planType: user.PlanTypePro,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.UpgradePlan(ctx, tt.userID, tt.planType)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpgradePlan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				updated, _ := service.GetByID(ctx, tt.userID)
				if updated.PlanType != tt.planType {
					t.Errorf("UpgradePlan() plan_type = %v, want %v", updated.PlanType, tt.planType)
				}
			}
		})
	}
}

func TestUserService_GetTrialStatus(t *testing.T) {
	mockRepo := testutil.NewMockUserRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := NewUserService(mockRepo, log)

	ctx := context.Background()
	u, _ := service.Create(ctx, "test@example.com")

	// Set user to trial
	u.PlanType = user.PlanTypeTrial
	service.Update(ctx, u)

	status, err := service.GetTrialStatus(ctx, u.ID)

	if err != nil {
		t.Errorf("GetTrialStatus() error = %v", err)
	}

	if status == nil {
		t.Error("GetTrialStatus() returned nil status")
		return
	}

	if status.Status != "active" {
		t.Errorf("GetTrialStatus() status = %v, want %v", status.Status, "active")
	}
}
