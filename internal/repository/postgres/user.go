package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/pratik-mahalle/infraudit/internal/domain/user"
	"github.com/pratik-mahalle/infraudit/internal/pkg/errors"
)

// UserRepository implements user.Repository using the profiles table
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) user.Repository {
	return &UserRepository{db: db}
}

// Create creates a new profile (fallback if the Supabase trigger hasn't fired)
func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	now := time.Now()
	u.CreatedAt = now
	u.UpdatedAt = now

	query := `
		INSERT INTO profiles (auth_id, email, username, full_name, role, plan_type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	var fullName interface{}
	if u.FullName != nil {
		fullName = *u.FullName
	}

	err := r.db.QueryRowContext(ctx, query,
		u.AuthID, u.Email, u.Username, fullName, u.Role, u.PlanType, now, now,
	).Scan(&u.ID)
	if err != nil {
		return errors.DatabaseError("Failed to create profile", err)
	}

	return nil
}

// GetByID retrieves a user profile by internal ID
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*user.User, error) {
	query := `
		SELECT id, auth_id, email, username, full_name, avatar_url, role, plan_type, created_at, updated_at
		FROM profiles WHERE id = $1
	`

	var u user.User
	var fullName, username, avatarURL sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.AuthID, &u.Email, &username, &fullName, &avatarURL, &u.Role, &u.PlanType, &u.CreatedAt, &u.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFound("User")
	}
	if err != nil {
		return nil, errors.DatabaseError("Failed to get user", err)
	}

	if fullName.Valid {
		u.FullName = &fullName.String
	}
	if username.Valid {
		u.Username = username.String
	}
	if avatarURL.Valid {
		u.AvatarURL = avatarURL.String
	}

	return &u, nil
}

// GetByAuthID retrieves a user profile by Supabase auth UUID
func (r *UserRepository) GetByAuthID(ctx context.Context, authID string) (*user.User, error) {
	query := `
		SELECT id, auth_id, email, username, full_name, avatar_url, role, plan_type, created_at, updated_at
		FROM profiles WHERE auth_id = $1
	`

	var u user.User
	var fullName, username, avatarURL sql.NullString

	err := r.db.QueryRowContext(ctx, query, authID).Scan(
		&u.ID, &u.AuthID, &u.Email, &username, &fullName, &avatarURL, &u.Role, &u.PlanType, &u.CreatedAt, &u.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFound("User")
	}
	if err != nil {
		return nil, errors.DatabaseError("Failed to get user by auth ID", err)
	}

	if fullName.Valid {
		u.FullName = &fullName.String
	}
	if username.Valid {
		u.Username = username.String
	}
	if avatarURL.Valid {
		u.AvatarURL = avatarURL.String
	}

	return &u, nil
}

// GetByEmail retrieves a user profile by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	query := `
		SELECT id, auth_id, email, username, full_name, avatar_url, role, plan_type, created_at, updated_at
		FROM profiles WHERE email = $1
	`

	var u user.User
	var fullName, username, avatarURL sql.NullString

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&u.ID, &u.AuthID, &u.Email, &username, &fullName, &avatarURL, &u.Role, &u.PlanType, &u.CreatedAt, &u.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFound("User")
	}
	if err != nil {
		return nil, errors.DatabaseError("Failed to get user", err)
	}

	if fullName.Valid {
		u.FullName = &fullName.String
	}
	if username.Valid {
		u.Username = username.String
	}
	if avatarURL.Valid {
		u.AvatarURL = avatarURL.String
	}

	return &u, nil
}

// ResolveAuthID resolves a Supabase auth UUID to an internal user ID
func (r *UserRepository) ResolveAuthID(ctx context.Context, authID string) (int64, error) {
	query := `SELECT id FROM profiles WHERE auth_id = $1`

	var id int64
	err := r.db.QueryRowContext(ctx, query, authID).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, errors.NotFound("User profile")
	}
	if err != nil {
		return 0, errors.DatabaseError("Failed to resolve auth ID", err)
	}

	return id, nil
}

// Update updates a user profile
func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	u.UpdatedAt = time.Now()

	query := `
		UPDATE profiles
		SET email = $1, username = $2, full_name = $3, role = $4, plan_type = $5, updated_at = $6
		WHERE id = $7
	`

	var fullName interface{}
	if u.FullName != nil {
		fullName = *u.FullName
	}

	result, err := r.db.ExecContext(ctx, query,
		u.Email, u.Username, fullName, u.Role, u.PlanType, u.UpdatedAt, u.ID,
	)
	if err != nil {
		return errors.DatabaseError("Failed to update user", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.DatabaseError("Failed to get affected rows", err)
	}

	if rows == 0 {
		return errors.NotFound("User")
	}

	return nil
}

// Delete deletes a user profile
func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM profiles WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.DatabaseError("Failed to delete user", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.DatabaseError("Failed to get affected rows", err)
	}

	if rows == 0 {
		return errors.NotFound("User")
	}

	return nil
}

// List retrieves all user profiles with pagination
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*user.User, int64, error) {
	// Get total count
	var total int64
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM profiles").Scan(&total)
	if err != nil {
		return nil, 0, errors.DatabaseError("Failed to count users", err)
	}

	// Get users
	query := `
		SELECT id, auth_id, email, username, full_name, avatar_url, role, plan_type, created_at, updated_at
		FROM profiles
		ORDER BY id DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, errors.DatabaseError("Failed to list users", err)
	}
	defer rows.Close()

	var users []*user.User
	for rows.Next() {
		var u user.User
		var fullName, username, avatarURL sql.NullString

		err := rows.Scan(&u.ID, &u.AuthID, &u.Email, &username, &fullName, &avatarURL, &u.Role, &u.PlanType, &u.CreatedAt, &u.UpdatedAt)
		if err != nil {
			return nil, 0, errors.DatabaseError("Failed to scan user", err)
		}

		if fullName.Valid {
			u.FullName = &fullName.String
		}
		if username.Valid {
			u.Username = username.String
		}
		if avatarURL.Valid {
			u.AvatarURL = avatarURL.String
		}

		users = append(users, &u)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, errors.DatabaseError("Failed to iterate users", err)
	}

	return users, total, nil
}
