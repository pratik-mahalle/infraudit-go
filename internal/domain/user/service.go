package user

import "context"

// Service defines the interface for user business logic
type Service interface {
	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id int64) (*User, error)

	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*User, error)

	// Create creates a new user
	Create(ctx context.Context, email string) (*User, error)

	// Update updates a user
	Update(ctx context.Context, user *User) error

	// GetTrialStatus gets the trial status for a user
	GetTrialStatus(ctx context.Context, userID int64) (*TrialStatus, error)

	// UpgradePlan upgrades a user's plan
	UpgradePlan(ctx context.Context, userID int64, planType string) error
}
