package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/pratik-mahalle/infraudit/internal/domain/user"
	"github.com/pratik-mahalle/infraudit/internal/pkg/errors"
)

// UserRepository implements user.Repository
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) user.Repository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	now := time.Now()
	u.CreatedAt = now
	u.UpdatedAt = now

	query := `
		INSERT INTO users (email, username, full_name, role, plan_type, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		u.Email, u.Username, u.FullName, u.Role, u.PlanType, now.Unix(), now.Unix(),
	)
	if err != nil {
		return errors.DatabaseError("Failed to create user", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return errors.DatabaseError("Failed to get user ID", err)
	}

	u.ID = id
	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*user.User, error) {
	query := `
		SELECT id, email, username, full_name, role, plan_type, created_at, updated_at
		FROM users WHERE id = ?
	`

	var u user.User
	var fullName sql.NullString
	var createdAt, updatedAt int64

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.Username, &fullName, &u.Role, &u.PlanType, &createdAt, &updatedAt,
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
	u.CreatedAt = time.Unix(createdAt, 0)
	u.UpdatedAt = time.Unix(updatedAt, 0)

	return &u, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	query := `
		SELECT id, email, username, full_name, role, plan_type, created_at, updated_at
		FROM users WHERE email = ?
	`

	var u user.User
	var fullName sql.NullString
	var createdAt, updatedAt int64

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&u.ID, &u.Email, &u.Username, &fullName, &u.Role, &u.PlanType, &createdAt, &updatedAt,
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
	u.CreatedAt = time.Unix(createdAt, 0)
	u.UpdatedAt = time.Unix(updatedAt, 0)

	return &u, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	u.UpdatedAt = time.Now()

	query := `
		UPDATE users
		SET email = ?, username = ?, full_name = ?, role = ?, plan_type = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		u.Email, u.Username, u.FullName, u.Role, u.PlanType, u.UpdatedAt.Unix(), u.ID,
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

// Delete deletes a user
func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = ?`

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

// List retrieves all users with pagination
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*user.User, int64, error) {
	// Get total count
	var total int64
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&total)
	if err != nil {
		return nil, 0, errors.DatabaseError("Failed to count users", err)
	}

	// Get users
	query := `
		SELECT id, email, username, full_name, role, plan_type, created_at, updated_at
		FROM users
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, errors.DatabaseError("Failed to list users", err)
	}
	defer rows.Close()

	var users []*user.User
	for rows.Next() {
		var u user.User
		var fullName sql.NullString
		var createdAt, updatedAt int64

		err := rows.Scan(&u.ID, &u.Email, &u.Username, &fullName, &u.Role, &u.PlanType, &createdAt, &updatedAt)
		if err != nil {
			return nil, 0, errors.DatabaseError("Failed to scan user", err)
		}

		if fullName.Valid {
			u.FullName = &fullName.String
		}
		u.CreatedAt = time.Unix(createdAt, 0)
		u.UpdatedAt = time.Unix(updatedAt, 0)

		users = append(users, &u)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, errors.DatabaseError("Failed to iterate users", err)
	}

	return users, total, nil
}
