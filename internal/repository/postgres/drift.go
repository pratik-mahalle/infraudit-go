package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"infraudit/backend/internal/domain/drift"
	"infraudit/backend/internal/pkg/errors"
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

	query := `INSERT INTO security_drifts (user_id, resource_id, drift_type, severity, details, detected_at, status) VALUES (?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.ExecContext(ctx, query, d.UserID, d.ResourceID, d.DriftType, d.Severity, d.Details, now.Format(time.RFC3339), d.Status)
	if err != nil {
		return 0, errors.DatabaseError("Failed to create drift", err)
	}

	return result.LastInsertId()
}

func (r *DriftRepository) GetByID(ctx context.Context, userID int64, id int64) (*drift.Drift, error) {
	query := `SELECT id, user_id, resource_id, drift_type, severity, details, detected_at, status FROM security_drifts WHERE user_id = ? AND id = ?`

	var d drift.Drift
	var detectedAt string
	err := r.db.QueryRowContext(ctx, query, userID, id).Scan(&d.ID, &d.UserID, &d.ResourceID, &d.DriftType, &d.Severity, &d.Details, &detectedAt, &d.Status)

	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Drift")
	}
	if err != nil {
		return nil, errors.DatabaseError("Failed to get drift", err)
	}

	d.DetectedAt, _ = time.Parse(time.RFC3339, detectedAt)
	return &d, nil
}

func (r *DriftRepository) Update(ctx context.Context, d *drift.Drift) error {
	d.UpdatedAt = time.Now()
	query := `UPDATE security_drifts SET resource_id = ?, drift_type = ?, severity = ?, details = ?, status = ? WHERE user_id = ? AND id = ?`

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
	result, err := r.db.ExecContext(ctx, "DELETE FROM security_drifts WHERE user_id = ? AND id = ?", userID, id)
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
	where := []string{"user_id = ?"}
	args := []interface{}{userID}

	if filter.ResourceID != "" {
		where = append(where, "resource_id = ?")
		args = append(args, filter.ResourceID)
	}
	if filter.DriftType != "" {
		where = append(where, "drift_type = ?")
		args = append(args, filter.DriftType)
	}
	if filter.Severity != "" {
		where = append(where, "severity = ?")
		args = append(args, filter.Severity)
	}
	if filter.Status != "" {
		where = append(where, "status = ?")
		args = append(args, filter.Status)
	}

	query := fmt.Sprintf(`SELECT id, user_id, resource_id, drift_type, severity, details, detected_at, status FROM security_drifts WHERE %s ORDER BY id DESC`, strings.Join(where, " AND "))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.DatabaseError("Failed to list drifts", err)
	}
	defer rows.Close()

	var drifts []*drift.Drift
	for rows.Next() {
		var d drift.Drift
		var detectedAt string
		err := rows.Scan(&d.ID, &d.UserID, &d.ResourceID, &d.DriftType, &d.Severity, &d.Details, &detectedAt, &d.Status)
		if err != nil {
			return nil, errors.DatabaseError("Failed to scan drift", err)
		}
		d.DetectedAt, _ = time.Parse(time.RFC3339, detectedAt)
		drifts = append(drifts, &d)
	}

	return drifts, rows.Err()
}

func (r *DriftRepository) ListWithPagination(ctx context.Context, userID int64, filter drift.Filter, limit, offset int) ([]*drift.Drift, int64, error) {
	where := []string{"user_id = ?"}
	args := []interface{}{userID}

	if filter.ResourceID != "" {
		where = append(where, "resource_id = ?")
		args = append(args, filter.ResourceID)
	}
	if filter.Severity != "" {
		where = append(where, "severity = ?")
		args = append(args, filter.Severity)
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM security_drifts WHERE %s", whereClause), args...).Scan(&total)
	if err != nil {
		return nil, 0, errors.DatabaseError("Failed to count drifts", err)
	}

	query := fmt.Sprintf(`SELECT id, user_id, resource_id, drift_type, severity, details, detected_at, status FROM security_drifts WHERE %s ORDER BY id DESC LIMIT ? OFFSET ?`, whereClause)

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, errors.DatabaseError("Failed to list drifts", err)
	}
	defer rows.Close()

	var drifts []*drift.Drift
	for rows.Next() {
		var d drift.Drift
		var detectedAt string
		err := rows.Scan(&d.ID, &d.UserID, &d.ResourceID, &d.DriftType, &d.Severity, &d.Details, &detectedAt, &d.Status)
		if err != nil {
			return nil, 0, errors.DatabaseError("Failed to scan drift", err)
		}
		d.DetectedAt, _ = time.Parse(time.RFC3339, detectedAt)
		drifts = append(drifts, &d)
	}

	return drifts, total, rows.Err()
}

func (r *DriftRepository) CountBySeverity(ctx context.Context, userID int64) (map[string]int, error) {
	query := `SELECT severity, COUNT(*) FROM security_drifts WHERE user_id = ? GROUP BY severity`

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
