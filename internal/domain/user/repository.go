package user

import "context"

// Repository defines the interface for user data access
type Repository interface {
	// Create creates a new user
	Create(ctx context.Context, user *User) error

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id int64) (*User, error)

	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*User, error)

	// Update updates a user
	Update(ctx context.Context, user *User) error

	// Delete deletes a user
	Delete(ctx context.Context, id int64) error

	// List retrieves all users with pagination
	List(ctx context.Context, limit, offset int) ([]*User, int64, error)
}
