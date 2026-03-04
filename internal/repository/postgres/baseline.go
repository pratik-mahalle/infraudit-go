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
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`

	var id int64
	err := r.db.QueryRowContext(ctx, query,
		b.UserID, b.ResourceID, b.Provider, b.ResourceType,
		b.Configuration, b.BaselineType, b.Description,
		now, now).Scan(&id)
	if err != nil {
		return 0, errors.DatabaseError("Failed to create baseline", err)
	}

	return id, nil
}

func (r *BaselineRepository) GetByResourceID(ctx context.Context, userID int64, resourceID string, baselineType string) (*baseline.Baseline, error) {
	query := `SELECT id, user_id, resource_id, provider, resource_type, configuration, baseline_type, description, created_at, updated_at
	          FROM resource_baselines
	          WHERE user_id = $1 AND resource_id = $2 AND baseline_type = $3`

	var b baseline.Baseline
	err := r.db.QueryRowContext(ctx, query, userID, resourceID, baselineType).Scan(
		&b.ID, &b.UserID, &b.ResourceID, &b.Provider, &b.ResourceType,
		&b.Configuration, &b.BaselineType, &b.Description,
		&b.CreatedAt, &b.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Baseline")
	}
	if err != nil {
		return nil, errors.DatabaseError("Failed to get baseline", err)
	}

	return &b, nil
}

func (r *BaselineRepository) Update(ctx context.Context, b *baseline.Baseline) error {
	b.UpdatedAt = time.Now()

	query := `UPDATE resource_baselines
	          SET configuration = $1, description = $2, updated_at = $3
	          WHERE user_id = $4 AND id = $5`

	_, err := r.db.ExecContext(ctx, query,
		b.Configuration, b.Description, b.UpdatedAt,
		b.UserID, b.ID)
	if err != nil {
		return errors.DatabaseError("Failed to update baseline", err)
	}

	return nil
}

func (r *BaselineRepository) Delete(ctx context.Context, userID int64, id int64) error {
	query := `DELETE FROM resource_baselines WHERE user_id = $1 AND id = $2`

	_, err := r.db.ExecContext(ctx, query, userID, id)
	if err != nil {
		return errors.DatabaseError("Failed to delete baseline", err)
	}

	return nil
}

func (r *BaselineRepository) List(ctx context.Context, userID int64) ([]*baseline.Baseline, error) {
	query := `SELECT id, user_id, resource_id, provider, resource_type, configuration, baseline_type, description, created_at, updated_at
	          FROM resource_baselines
	          WHERE user_id = $1
	          ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.DatabaseError("Failed to list baselines", err)
	}
	defer rows.Close()

	var baselines []*baseline.Baseline
	for rows.Next() {
		var b baseline.Baseline
		err := rows.Scan(&b.ID, &b.UserID, &b.ResourceID, &b.Provider, &b.ResourceType,
			&b.Configuration, &b.BaselineType, &b.Description,
			&b.CreatedAt, &b.UpdatedAt)
		if err != nil {
			return nil, errors.DatabaseError("Failed to scan baseline", err)
		}

		baselines = append(baselines, &b)
	}

	return baselines, nil
}
