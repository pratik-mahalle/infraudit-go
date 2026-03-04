package services

import (
	"context"
	"testing"

	"github.com/pratik-mahalle/infraudit/internal/domain/user"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
	"github.com/pratik-mahalle/infraudit/internal/testutil"
)

// helper to create a profile directly in the mock repo
func seedProfile(repo *testutil.MockUserRepository, authID, email string) *user.User {
	u := &user.User{
		AuthID:   authID,
		Email:    email,
		Username: email,
		Role:     user.RoleUser,
		PlanType: user.PlanTypeFree,
	}
	_ = repo.Create(context.Background(), u)
	return u
}

func TestUserService_EnsureProfile(t *testing.T) {
	mockRepo := testutil.NewMockUserRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := NewUserService(mockRepo, log)

	ctx := context.Background()

	t.Run("creates new profile when none exists", func(t *testing.T) {
		u, err := service.EnsureProfile(ctx, "auth-uuid-1", "new@example.com", "New User")
		if err != nil {
			t.Fatalf("EnsureProfile() error = %v", err)
		}
		if u == nil {
			t.Fatal("EnsureProfile() returned nil user")
		}
		if u.Email != "new@example.com" {
			t.Errorf("EnsureProfile() email = %v, want %v", u.Email, "new@example.com")
		}
		if u.AuthID != "auth-uuid-1" {
			t.Errorf("EnsureProfile() authID = %v, want %v", u.AuthID, "auth-uuid-1")
		}
		if u.Role != user.RoleUser {
			t.Errorf("EnsureProfile() role = %v, want %v", u.Role, user.RoleUser)
		}
	})

	t.Run("returns existing profile", func(t *testing.T) {
		u, err := service.EnsureProfile(ctx, "auth-uuid-1", "new@example.com", "New User")
		if err != nil {
			t.Fatalf("EnsureProfile() error = %v", err)
		}
		if u.Email != "new@example.com" {
			t.Errorf("EnsureProfile() should return existing profile")
		}
	})
}

func TestUserService_GetByID(t *testing.T) {
	mockRepo := testutil.NewMockUserRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := NewUserService(mockRepo, log)

	ctx := context.Background()
	createdUser := seedProfile(mockRepo, "auth-uuid-2", "test@example.com")

	t.Run("get existing user", func(t *testing.T) {
		u, err := service.GetByID(ctx, createdUser.ID)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}
		if u == nil {
			t.Fatal("GetByID() returned nil user")
		}
	})

	t.Run("get non-existing user", func(t *testing.T) {
		_, err := service.GetByID(ctx, 999)
		if err == nil {
			t.Error("GetByID() expected error for non-existing user")
		}
	})
}

func TestUserService_GetByAuthID(t *testing.T) {
	mockRepo := testutil.NewMockUserRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := NewUserService(mockRepo, log)

	ctx := context.Background()
	seedProfile(mockRepo, "auth-uuid-3", "auth@example.com")

	t.Run("get existing user by auth ID", func(t *testing.T) {
		u, err := service.GetByAuthID(ctx, "auth-uuid-3")
		if err != nil {
			t.Fatalf("GetByAuthID() error = %v", err)
		}
		if u.Email != "auth@example.com" {
			t.Errorf("GetByAuthID() email = %v, want %v", u.Email, "auth@example.com")
		}
	})

	t.Run("get non-existing user by auth ID", func(t *testing.T) {
		_, err := service.GetByAuthID(ctx, "nonexistent-uuid")
		if err == nil {
			t.Error("GetByAuthID() expected error for non-existing auth ID")
		}
	})
}

func TestUserService_GetByEmail(t *testing.T) {
	mockRepo := testutil.NewMockUserRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := NewUserService(mockRepo, log)

	ctx := context.Background()
	seedProfile(mockRepo, "auth-uuid-4", "email@example.com")

	t.Run("get existing user by email", func(t *testing.T) {
		u, err := service.GetByEmail(ctx, "email@example.com")
		if err != nil {
			t.Fatalf("GetByEmail() error = %v", err)
		}
		if u.Email != "email@example.com" {
			t.Errorf("GetByEmail() email = %v, want %v", u.Email, "email@example.com")
		}
	})

	t.Run("get non-existing user by email", func(t *testing.T) {
		_, err := service.GetByEmail(ctx, "nonexistent@example.com")
		if err == nil {
			t.Error("GetByEmail() expected error for non-existing email")
		}
	})
}

func TestUserService_Update(t *testing.T) {
	mockRepo := testutil.NewMockUserRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := NewUserService(mockRepo, log)

	ctx := context.Background()
	u := seedProfile(mockRepo, "auth-uuid-5", "update@example.com")

	u.PlanType = user.PlanTypePro
	err := service.Update(ctx, u)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

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
	u := seedProfile(mockRepo, "auth-uuid-6", "upgrade@example.com")

	tests := []struct {
		name     string
		userID   int64
		planType string
		wantErr  bool
	}{
		{"upgrade to pro", u.ID, user.PlanTypePro, false},
		{"upgrade to enterprise", u.ID, user.PlanTypeEnterprise, false},
		{"upgrade non-existing user", 999, user.PlanTypePro, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.UpgradePlan(ctx, tt.userID, tt.planType)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpgradePlan() error = %v, wantErr %v", err, tt.wantErr)
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
