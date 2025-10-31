package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"infraaudit/backend/internal/domain/resource"
	"infraaudit/backend/internal/pkg/errors"
)

// ResourceRepository implements resource.Repository
type ResourceRepository struct {
	db *sql.DB
}

// NewResourceRepository creates a new resource repository
func NewResourceRepository(db *sql.DB) resource.Repository {
	return &ResourceRepository{db: db}
}

// Create creates a new resource
func (r *ResourceRepository) Create(ctx context.Context, res *resource.Resource) error {
	now := time.Now()
	res.CreatedAt = now
	res.UpdatedAt = now

	query := `
		INSERT INTO resources (user_id, provider, resource_id, name, type, region, status)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		res.UserID, res.Provider, res.ResourceID, res.Name, res.Type, res.Region, res.Status,
	)
	if err != nil {
		return errors.DatabaseError("Failed to create resource", err)
	}

	return nil
}

// GetByID retrieves a resource by ID
func (r *ResourceRepository) GetByID(ctx context.Context, userID int64, resourceID string) (*resource.Resource, error) {
	query := `
		SELECT user_id, provider, resource_id, name, type, region, status
		FROM resources
		WHERE user_id = ? AND resource_id = ?
	`

	var res resource.Resource
	err := r.db.QueryRowContext(ctx, query, userID, resourceID).Scan(
		&res.UserID, &res.Provider, &res.ResourceID, &res.Name, &res.Type, &res.Region, &res.Status,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Resource")
	}
	if err != nil {
		return nil, errors.DatabaseError("Failed to get resource", err)
	}

	return &res, nil
}

// Update updates a resource
func (r *ResourceRepository) Update(ctx context.Context, res *resource.Resource) error {
	res.UpdatedAt = time.Now()

	query := `
		UPDATE resources
		SET name = ?, type = ?, region = ?, status = ?
		WHERE user_id = ? AND resource_id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		res.Name, res.Type, res.Region, res.Status, res.UserID, res.ResourceID,
	)
	if err != nil {
		return errors.DatabaseError("Failed to update resource", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.DatabaseError("Failed to get affected rows", err)
	}

	if rows == 0 {
		return errors.NotFound("Resource")
	}

	return nil
}

// Delete deletes a resource
func (r *ResourceRepository) Delete(ctx context.Context, userID int64, resourceID string) error {
	query := `DELETE FROM resources WHERE user_id = ? AND resource_id = ?`

	result, err := r.db.ExecContext(ctx, query, userID, resourceID)
	if err != nil {
		return errors.DatabaseError("Failed to delete resource", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.DatabaseError("Failed to get affected rows", err)
	}

	if rows == 0 {
		return errors.NotFound("Resource")
	}

	return nil
}

// List retrieves resources with filters and pagination
func (r *ResourceRepository) List(ctx context.Context, userID int64, filter resource.Filter, limit, offset int) ([]*resource.Resource, int64, error) {
	// Build WHERE clause
	where := []string{"user_id = ?"}
	args := []interface{}{userID}

	if filter.Provider != "" {
		where = append(where, "provider = ?")
		args = append(args, filter.Provider)
	}
	if filter.Type != "" {
		where = append(where, "type = ?")
		args = append(args, filter.Type)
	}
	if filter.Region != "" {
		where = append(where, "region = ?")
		args = append(args, filter.Region)
	}
	if filter.Status != "" {
		where = append(where, "status = ?")
		args = append(args, filter.Status)
	}

	whereClause := strings.Join(where, " AND ")

	// Get total count
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM resources WHERE %s", whereClause)
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, errors.DatabaseError("Failed to count resources", err)
	}

	// Get resources
	query := fmt.Sprintf(`
		SELECT user_id, provider, resource_id, name, type, region, status
		FROM resources
		WHERE %s
		ORDER BY provider, resource_id
		LIMIT ? OFFSET ?
	`, whereClause)

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, errors.DatabaseError("Failed to list resources", err)
	}
	defer rows.Close()

	var resources []*resource.Resource
	for rows.Next() {
		var res resource.Resource
		err := rows.Scan(&res.UserID, &res.Provider, &res.ResourceID, &res.Name, &res.Type, &res.Region, &res.Status)
		if err != nil {
			return nil, 0, errors.DatabaseError("Failed to scan resource", err)
		}
		resources = append(resources, &res)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, errors.DatabaseError("Failed to iterate resources", err)
	}

	return resources, total, nil
}

// ListByProvider retrieves resources by provider
func (r *ResourceRepository) ListByProvider(ctx context.Context, userID int64, provider string) ([]*resource.Resource, error) {
	query := `
		SELECT user_id, provider, resource_id, name, type, region, status
		FROM resources
		WHERE user_id = ? AND provider = ?
		ORDER BY resource_id
	`

	rows, err := r.db.QueryContext(ctx, query, userID, provider)
	if err != nil {
		return nil, errors.DatabaseError("Failed to list resources by provider", err)
	}
	defer rows.Close()

	var resources []*resource.Resource
	for rows.Next() {
		var res resource.Resource
		err := rows.Scan(&res.UserID, &res.Provider, &res.ResourceID, &res.Name, &res.Type, &res.Region, &res.Status)
		if err != nil {
			return nil, errors.DatabaseError("Failed to scan resource", err)
		}
		resources = append(resources, &res)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.DatabaseError("Failed to iterate resources", err)
	}

	return resources, nil
}

// SaveBatch saves multiple resources (used for sync)
func (r *ResourceRepository) SaveBatch(ctx context.Context, userID int64, provider string, resources []*resource.Resource) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.DatabaseError("Failed to begin transaction", err)
	}
	defer tx.Rollback()

	// Delete existing resources for this provider
	_, err = tx.ExecContext(ctx, "DELETE FROM resources WHERE user_id = ? AND provider = ?", userID, provider)
	if err != nil {
		return errors.DatabaseError("Failed to delete old resources", err)
	}

	// Insert new resources
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO resources (user_id, provider, resource_id, name, type, region, status)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return errors.DatabaseError("Failed to prepare statement", err)
	}
	defer stmt.Close()

	for _, res := range resources {
		_, err := stmt.ExecContext(ctx, userID, provider, res.ResourceID, res.Name, res.Type, res.Region, res.Status)
		if err != nil {
			return errors.DatabaseError("Failed to insert resource", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.DatabaseError("Failed to commit transaction", err)
	}

	return nil
}

// DeleteByProvider deletes all resources for a provider
func (r *ResourceRepository) DeleteByProvider(ctx context.Context, userID int64, provider string) error {
	query := `DELETE FROM resources WHERE user_id = ? AND provider = ?`

	_, err := r.db.ExecContext(ctx, query, userID, provider)
	if err != nil {
		return errors.DatabaseError("Failed to delete resources by provider", err)
	}

	return nil
}
