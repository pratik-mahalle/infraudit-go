package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/pratik-mahalle/infraudit/internal/domain/drift"
	"github.com/pratik-mahalle/infraudit/internal/pkg/errors"
)

type DriftRepository struct {
	db *sql.DB
}

func NewDriftRepository(db *sql.DB) drift.Repository {
	return &DriftRepository{db: db}
}

func (r *DriftRepository) Create(ctx context.Context, d *drift.Drift) (int64, error) {
	now := time.Now()
	d.CreatedAt = now
	d.DetectedAt = now

	query := `INSERT INTO drifts (user_id, resource_id, type, severity, description, detected_at, status) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	var id int64
	err := r.db.QueryRowContext(ctx, query, d.UserID, d.ResourceID, d.DriftType, d.Severity, d.Details, now, d.Status).Scan(&id)
	if err != nil {
		return 0, errors.DatabaseError("Failed to create drift", err)
	}

	return id, nil
}

func (r *DriftRepository) GetByID(ctx context.Context, userID int64, id int64) (*drift.Drift, error) {
	query := `SELECT id, user_id, resource_id, type, severity, description, detected_at, status FROM drifts WHERE user_id = $1 AND id = $2`

	var d drift.Drift
	err := r.db.QueryRowContext(ctx, query, userID, id).Scan(&d.ID, &d.UserID, &d.ResourceID, &d.DriftType, &d.Severity, &d.Details, &d.DetectedAt, &d.Status)

	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Drift")
	}
	if err != nil {
		return nil, errors.DatabaseError("Failed to get drift", err)
	}

	return &d, nil
}

func (r *DriftRepository) Update(ctx context.Context, d *drift.Drift) error {
	d.UpdatedAt = time.Now()
	query := `UPDATE drifts SET resource_id = $1, type = $2, severity = $3, description = $4, status = $5 WHERE user_id = $6 AND id = $7`

	result, err := r.db.ExecContext(ctx, query, d.ResourceID, d.DriftType, d.Severity, d.Details, d.Status, d.UserID, d.ID)
	if err != nil {
		return errors.DatabaseError("Failed to update drift", err)
	}

	rows, err := result.RowsAffected()
	if err != nil || rows == 0 {
		return errors.NotFound("Drift")
	}

	return nil
}

func (r *DriftRepository) Delete(ctx context.Context, userID int64, id int64) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM drifts WHERE user_id = $1 AND id = $2", userID, id)
	if err != nil {
		return errors.DatabaseError("Failed to delete drift", err)
	}

	rows, err := result.RowsAffected()
	if err != nil || rows == 0 {
		return errors.NotFound("Drift")
	}

	return nil
}

func (r *DriftRepository) List(ctx context.Context, userID int64, filter drift.Filter) ([]*drift.Drift, error) {
	paramN := 1
	where := []string{fmt.Sprintf("user_id = $%d", paramN)}
	args := []interface{}{userID}
	paramN++

	if filter.ResourceID != "" {
		where = append(where, fmt.Sprintf("resource_id = $%d", paramN))
		args = append(args, filter.ResourceID)
		paramN++
	}
	if filter.DriftType != "" {
		where = append(where, fmt.Sprintf("type = $%d", paramN))
		args = append(args, filter.DriftType)
		paramN++
	}
	if filter.Severity != "" {
		where = append(where, fmt.Sprintf("severity = $%d", paramN))
		args = append(args, filter.Severity)
		paramN++
	}
	if filter.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", paramN))
		args = append(args, filter.Status)
		paramN++
	}

	query := fmt.Sprintf(`SELECT id, user_id, resource_id, type, severity, description, detected_at, status FROM drifts WHERE %s ORDER BY id DESC`, strings.Join(where, " AND "))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.DatabaseError("Failed to list drifts", err)
	}
	defer rows.Close()

	// Pre-allocate slice with reasonable capacity
	drifts := make([]*drift.Drift, 0, 100)
	for rows.Next() {
		var d drift.Drift
		err := rows.Scan(&d.ID, &d.UserID, &d.ResourceID, &d.DriftType, &d.Severity, &d.Details, &d.DetectedAt, &d.Status)
		if err != nil {
			return nil, errors.DatabaseError("Failed to scan drift", err)
		}
		drifts = append(drifts, &d)
	}

	return drifts, rows.Err()
}

func (r *DriftRepository) ListWithPagination(ctx context.Context, userID int64, filter drift.Filter, limit, offset int) ([]*drift.Drift, int64, error) {
	paramN := 1
	where := []string{fmt.Sprintf("user_id = $%d", paramN)}
	args := []interface{}{userID}
	paramN++

	if filter.ResourceID != "" {
		where = append(where, fmt.Sprintf("resource_id = $%d", paramN))
		args = append(args, filter.ResourceID)
		paramN++
	}
	if filter.Severity != "" {
		where = append(where, fmt.Sprintf("severity = $%d", paramN))
		args = append(args, filter.Severity)
		paramN++
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM drifts WHERE %s", whereClause), args...).Scan(&total)
	if err != nil {
		return nil, 0, errors.DatabaseError("Failed to count drifts", err)
	}

	query := fmt.Sprintf(`SELECT id, user_id, resource_id, type, severity, description, detected_at, status FROM drifts WHERE %s ORDER BY id DESC LIMIT $%d OFFSET $%d`, whereClause, paramN, paramN+1)

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, errors.DatabaseError("Failed to list drifts", err)
	}
	defer rows.Close()

	// Pre-allocate slice with expected capacity
	drifts := make([]*drift.Drift, 0, limit)
	for rows.Next() {
		var d drift.Drift
		err := rows.Scan(&d.ID, &d.UserID, &d.ResourceID, &d.DriftType, &d.Severity, &d.Details, &d.DetectedAt, &d.Status)
		if err != nil {
			return nil, 0, errors.DatabaseError("Failed to scan drift", err)
		}
		drifts = append(drifts, &d)
	}

	return drifts, total, rows.Err()
}

func (r *DriftRepository) CountBySeverity(ctx context.Context, userID int64) (map[string]int, error) {
	query := `SELECT severity, COUNT(*) FROM drifts WHERE user_id = $1 GROUP BY severity`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.DatabaseError("Failed to count drifts by severity", err)
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var severity string
		var count int
		if err := rows.Scan(&severity, &count); err != nil {
			return nil, errors.DatabaseError("Failed to scan count", err)
		}
		counts[severity] = count
	}

	return counts, rows.Err()
}
