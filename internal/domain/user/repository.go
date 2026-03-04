package user

import "context"

// Repository defines the interface for user data access
type Repository interface {
	// Create creates a new user profile
	Create(ctx context.Context, user *User) error

	// GetByID retrieves a user by internal ID
	GetByID(ctx context.Context, id int64) (*User, error)

	// GetByAuthID retrieves a user by Supabase auth UUID
	GetByAuthID(ctx context.Context, authID string) (*User, error)

	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*User, error)

	// ResolveAuthID resolves a Supabase auth UUID to an internal user ID
	ResolveAuthID(ctx context.Context, authID string) (int64, error)

	// Update updates a user
	Update(ctx context.Context, user *User) error

	// Delete deletes a user
	Delete(ctx context.Context, id int64) error

	// List retrieves all users with pagination
	List(ctx context.Context, limit, offset int) ([]*User, int64, error)
}
