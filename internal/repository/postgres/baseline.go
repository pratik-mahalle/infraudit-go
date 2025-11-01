package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/pratik-mahalle/infraudit/internal/domain/baseline"
	"github.com/pratik-mahalle/infraudit/internal/pkg/errors"
)

type BaselineRepository struct {
	db *sql.DB
}

func NewBaselineRepository(db *sql.DB) baseline.Repository {
	return &BaselineRepository{db: db}
}

func (r *BaselineRepository) Create(ctx context.Context, b *baseline.Baseline) (int64, error) {
	now := time.Now()
	b.CreatedAt = now
	b.UpdatedAt = now

	query := `INSERT INTO resource_baselines (user_id, resource_id, provider, resource_type, configuration, baseline_type, description, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.ExecContext(ctx, query,
		b.UserID, b.ResourceID, b.Provider, b.ResourceType,
		b.Configuration, b.BaselineType, b.Description,
		now.Format(time.RFC3339), now.Format(time.RFC3339))
	if err != nil {
		return 0, errors.DatabaseError("Failed to create baseline", err)
	}

	return result.LastInsertId()
}

func (r *BaselineRepository) GetByResourceID(ctx context.Context, userID int64, resourceID string, baselineType string) (*baseline.Baseline, error) {
	query := `SELECT id, user_id, resource_id, provider, resource_type, configuration, baseline_type, description, created_at, updated_at
	          FROM resource_baselines
	          WHERE user_id = ? AND resource_id = ? AND baseline_type = ?`

	var b baseline.Baseline
	var createdAt, updatedAt string
	err := r.db.QueryRowContext(ctx, query, userID, resourceID, baselineType).Scan(
		&b.ID, &b.UserID, &b.ResourceID, &b.Provider, &b.ResourceType,
		&b.Configuration, &b.BaselineType, &b.Description,
		&createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Baseline")
	}
	if err != nil {
		return nil, errors.DatabaseError("Failed to get baseline", err)
	}

	b.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	b.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return &b, nil
}

func (r *BaselineRepository) Update(ctx context.Context, b *baseline.Baseline) error {
	b.UpdatedAt = time.Now()

	query := `UPDATE resource_baselines
	          SET configuration = ?, description = ?, updated_at = ?
	          WHERE user_id = ? AND id = ?`

	_, err := r.db.ExecContext(ctx, query,
		b.Configuration, b.Description, b.UpdatedAt.Format(time.RFC3339),
		b.UserID, b.ID)
	if err != nil {
		return errors.DatabaseError("Failed to update baseline", err)
	}

	return nil
}

func (r *BaselineRepository) Delete(ctx context.Context, userID int64, id int64) error {
	query := `DELETE FROM resource_baselines WHERE user_id = ? AND id = ?`

	_, err := r.db.ExecContext(ctx, query, userID, id)
	if err != nil {
		return errors.DatabaseError("Failed to delete baseline", err)
	}

	return nil
}

func (r *BaselineRepository) List(ctx context.Context, userID int64) ([]*baseline.Baseline, error) {
	query := `SELECT id, user_id, resource_id, provider, resource_type, configuration, baseline_type, description, created_at, updated_at
	          FROM resource_baselines
	          WHERE user_id = ?
	          ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.DatabaseError("Failed to list baselines", err)
	}
	defer rows.Close()

	var baselines []*baseline.Baseline
	for rows.Next() {
		var b baseline.Baseline
		var createdAt, updatedAt string
		err := rows.Scan(&b.ID, &b.UserID, &b.ResourceID, &b.Provider, &b.ResourceType,
			&b.Configuration, &b.BaselineType, &b.Description,
			&createdAt, &updatedAt)
		if err != nil {
			return nil, errors.DatabaseError("Failed to scan baseline", err)
		}

		b.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		b.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

		baselines = append(baselines, &b)
	}

	return baselines, nil
}
