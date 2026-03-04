package user

import "context"

// Service defines the interface for user business logic
type Service interface {
	// GetByID retrieves a user by internal ID
	GetByID(ctx context.Context, id int64) (*User, error)

	// GetByAuthID retrieves a user by Supabase auth UUID
	GetByAuthID(ctx context.Context, authID string) (*User, error)

	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*User, error)

	// EnsureProfile creates a profile if it doesn't exist (fallback for trigger)
	EnsureProfile(ctx context.Context, authID, email, fullName string) (*User, error)

	// Update updates a user
	Update(ctx context.Context, user *User) error

	// GetTrialStatus gets the trial status for a user
	GetTrialStatus(ctx context.Context, userID int64) (*TrialStatus, error)

	// UpgradePlan upgrades a user's plan
	UpgradePlan(ctx context.Context, userID int64, planType string) error
}
