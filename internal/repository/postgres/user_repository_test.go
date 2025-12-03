package postgres

import (
	"context"
	"testing"

	"github.com/pratik-mahalle/infraudit/internal/domain/user"
	"github.com/pratik-mahalle/infraudit/internal/testutil"
)

func TestUserRepository_Create(t *testing.T) {
	db := testutil.NewTestDB(t)
	defer testutil.CleanupDB(db)

	repo := NewUserRepository(db)

	tests := []struct {
		name    string
		user    *user.User
		wantErr bool
	}{
		{
			name: "create user successfully",
			user: &user.User{
				Email:    "test@example.com",
				Role:     user.RoleUser,
				PlanType: user.PlanTypeFree,
			},
			wantErr: false,
		},
		{
			name: "create another user",
			user: &user.User{
				Email:    "another@example.com",
				Role:     user.RoleUser,
				PlanType: user.PlanTypeFree,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := repo.Create(ctx, tt.user)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tt.user.ID == 0 {
					t.Error("Create() did not set user ID")
				}
			}
		})
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	db := testutil.NewTestDB(t)
	defer testutil.CleanupDB(db)

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create a user
	u := &user.User{
		Email:    "test@example.com",
		Role:     user.RoleUser,
		PlanType: user.PlanTypeFree,
	}
	repo.Create(ctx, u)

	tests := []struct {
		name    string
		userID  int64
		wantErr bool
	}{
		{
			name:    "get existing user",
			userID:  u.ID,
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
			got, err := repo.GetByID(ctx, tt.userID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got == nil {
					t.Error("GetByID() returned nil user")
					return
				}
				if got.ID != tt.userID {
					t.Errorf("GetByID() ID = %v, want %v", got.ID, tt.userID)
				}
				if got.Email != u.Email {
					t.Errorf("GetByID() Email = %v, want %v", got.Email, u.Email)
				}
			}
		})
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db := testutil.NewTestDB(t)
	defer testutil.CleanupDB(db)

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create a user
	email := "test@example.com"
	u := &user.User{
		Email:    email,
		Role:     user.RoleUser,
		PlanType: user.PlanTypeFree,
	}
	repo.Create(ctx, u)

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
			got, err := repo.GetByEmail(ctx, tt.email)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetByEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got == nil {
					t.Error("GetByEmail() returned nil user")
					return
				}
				if got.Email != tt.email {
					t.Errorf("GetByEmail() Email = %v, want %v", got.Email, tt.email)
				}
			}
		})
	}
}

func TestUserRepository_Update(t *testing.T) {
	db := testutil.NewTestDB(t)
	defer testutil.CleanupDB(db)

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create a user
	u := &user.User{
		Email:    "test@example.com",
		Role:     user.RoleUser,
		PlanType: user.PlanTypeFree,
	}
	repo.Create(ctx, u)

	// Update user
	u.PlanType = user.PlanTypePro
	err := repo.Update(ctx, u)
	if err != nil {
		t.Errorf("Update() error = %v", err)
	}

	// Verify update
	updated, err := repo.GetByID(ctx, u.ID)
	if err != nil {
		t.Errorf("GetByID() after update error = %v", err)
	}

	if updated.PlanType != user.PlanTypePro {
		t.Errorf("Update() PlanType = %v, want %v", updated.PlanType, user.PlanTypePro)
	}
}

func TestUserRepository_Delete(t *testing.T) {
	db := testutil.NewTestDB(t)
	defer testutil.CleanupDB(db)

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create a user
	u := &user.User{
		Email:    "test@example.com",
		Role:     user.RoleUser,
		PlanType: user.PlanTypeFree,
	}
	repo.Create(ctx, u)

	// Delete user
	err := repo.Delete(ctx, u.ID)
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// Verify deletion
	_, err = repo.GetByID(ctx, u.ID)
	if err == nil {
		t.Error("Delete() user still exists after deletion")
	}
}
