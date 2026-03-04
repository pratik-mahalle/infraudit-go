package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/pratik-mahalle/infraudit/internal/domain/resource"
	"github.com/pratik-mahalle/infraudit/internal/pkg/errors"
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
		INSERT INTO resources (user_id, provider, resource_id, name, resource_type, region, status, configuration)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		res.UserID, res.Provider, res.ResourceID, res.Name, res.Type, res.Region, res.Status, res.Configuration,
	)
	if err != nil {
		return errors.DatabaseError("Failed to create resource", err)
	}

	return nil
}

// GetByID retrieves a resource by ID
func (r *ResourceRepository) GetByID(ctx context.Context, userID int64, resourceID string) (*resource.Resource, error) {
	query := `
		SELECT user_id, provider, resource_id, name, resource_type, region, status, COALESCE(configuration, '')
		FROM resources
		WHERE user_id = $1 AND resource_id = $2
	`

	var res resource.Resource
	err := r.db.QueryRowContext(ctx, query, userID, resourceID).Scan(
		&res.UserID, &res.Provider, &res.ResourceID, &res.Name, &res.Type, &res.Region, &res.Status, &res.Configuration,
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
		SET name = $1, resource_type = $2, region = $3, status = $4, configuration = $5
		WHERE user_id = $6 AND resource_id = $7
	`

	result, err := r.db.ExecContext(ctx, query,
		res.Name, res.Type, res.Region, res.Status, res.Configuration, res.UserID, res.ResourceID,
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
	query := `DELETE FROM resources WHERE user_id = $1 AND resource_id = $2`

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
	// Build WHERE clause with numbered placeholders
	paramN := 1
	where := []string{fmt.Sprintf("user_id = $%d", paramN)}
	args := []interface{}{userID}
	paramN++

	if filter.Provider != "" {
		where = append(where, fmt.Sprintf("provider = $%d", paramN))
		args = append(args, filter.Provider)
		paramN++
	}
	if filter.Type != "" {
		where = append(where, fmt.Sprintf("resource_type = $%d", paramN))
		args = append(args, filter.Type)
		paramN++
	}
	if filter.Region != "" {
		where = append(where, fmt.Sprintf("region = $%d", paramN))
		args = append(args, filter.Region)
		paramN++
	}
	if filter.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", paramN))
		args = append(args, filter.Status)
		paramN++
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
		SELECT user_id, provider, resource_id, name, resource_type, region, status, COALESCE(configuration, '')
		FROM resources
		WHERE %s
		ORDER BY provider, resource_id
		LIMIT $%d OFFSET $%d
	`, whereClause, paramN, paramN+1)

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, errors.DatabaseError("Failed to list resources", err)
	}
	defer rows.Close()

	// Pre-allocate slice with expected capacity to avoid repeated allocations
	resources := make([]*resource.Resource, 0, limit)
	for rows.Next() {
		var res resource.Resource
		err := rows.Scan(&res.UserID, &res.Provider, &res.ResourceID, &res.Name, &res.Type, &res.Region, &res.Status, &res.Configuration)
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
		SELECT user_id, provider, resource_id, name, resource_type, region, status, COALESCE(configuration, '')
		FROM resources
		WHERE user_id = $1 AND provider = $2
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
		err := rows.Scan(&res.UserID, &res.Provider, &res.ResourceID, &res.Name, &res.Type, &res.Region, &res.Status, &res.Configuration)
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
	_, err = tx.ExecContext(ctx, "DELETE FROM resources WHERE user_id = $1 AND provider = $2", userID, provider)
	if err != nil {
		return errors.DatabaseError("Failed to delete old resources", err)
	}

	// Insert new resources
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO resources (user_id, provider, resource_id, name, resource_type, region, status, configuration)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`)
	if err != nil {
		return errors.DatabaseError("Failed to prepare statement", err)
	}
	defer stmt.Close()

	for _, res := range resources {
		_, err := stmt.ExecContext(ctx, userID, provider, res.ResourceID, res.Name, res.Type, res.Region, res.Status, res.Configuration)
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
	query := `DELETE FROM resources WHERE user_id = $1 AND provider = $2`

	_, err := r.db.ExecContext(ctx, query, userID, provider)
	if err != nil {
		return errors.DatabaseError("Failed to delete resources by provider", err)
	}

	return nil
}
