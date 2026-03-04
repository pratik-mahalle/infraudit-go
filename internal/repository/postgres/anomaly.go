package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/pratik-mahalle/infraudit/internal/domain/anomaly"
	"github.com/pratik-mahalle/infraudit/internal/pkg/errors"
)

type AnomalyRepository struct {
	db *sql.DB
}

func NewAnomalyRepository(db *sql.DB) anomaly.Repository {
	return &AnomalyRepository{db: db}
}

func (r *AnomalyRepository) Create(ctx context.Context, a *anomaly.Anomaly) (int64, error) {
	now := time.Now()
	a.CreatedAt = now
	a.DetectedAt = now

	query := `INSERT INTO cost_anomalies (user_id, resource_id, anomaly_type, severity, percentage, previous_cost, current_cost, detected_at, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`

	var id int64
	err := r.db.QueryRowContext(ctx, query, a.UserID, a.ResourceID, a.AnomalyType, a.Severity, a.Percentage, a.PreviousCost, a.CurrentCost, now, a.Status).Scan(&id)
	if err != nil {
		return 0, errors.DatabaseError("Failed to create anomaly", err)
	}

	return id, nil
}

func (r *AnomalyRepository) GetByID(ctx context.Context, userID int64, id int64) (*anomaly.Anomaly, error) {
	query := `SELECT id, user_id, resource_id, anomaly_type, severity, percentage, previous_cost, current_cost, detected_at, status FROM cost_anomalies WHERE user_id = $1 AND id = $2`

	var a anomaly.Anomaly
	err := r.db.QueryRowContext(ctx, query, userID, id).Scan(&a.ID, &a.UserID, &a.ResourceID, &a.AnomalyType, &a.Severity, &a.Percentage, &a.PreviousCost, &a.CurrentCost, &a.DetectedAt, &a.Status)

	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Anomaly")
	}
	if err != nil {
		return nil, errors.DatabaseError("Failed to get anomaly", err)
	}

	return &a, nil
}

func (r *AnomalyRepository) Update(ctx context.Context, a *anomaly.Anomaly) error {
	a.UpdatedAt = time.Now()
	query := `UPDATE cost_anomalies SET resource_id = $1, anomaly_type = $2, severity = $3, percentage = $4, previous_cost = $5, current_cost = $6, status = $7 WHERE user_id = $8 AND id = $9`

	result, err := r.db.ExecContext(ctx, query, a.ResourceID, a.AnomalyType, a.Severity, a.Percentage, a.PreviousCost, a.CurrentCost, a.Status, a.UserID, a.ID)
	if err != nil {
		return errors.DatabaseError("Failed to update anomaly", err)
	}

	rows, err := result.RowsAffected()
	if err != nil || rows == 0 {
		return errors.NotFound("Anomaly")
	}

	return nil
}

func (r *AnomalyRepository) Delete(ctx context.Context, userID int64, id int64) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM cost_anomalies WHERE user_id = $1 AND id = $2", userID, id)
	if err != nil {
		return errors.DatabaseError("Failed to delete anomaly", err)
	}

	rows, err := result.RowsAffected()
	if err != nil || rows == 0 {
		return errors.NotFound("Anomaly")
	}

	return nil
}

func (r *AnomalyRepository) List(ctx context.Context, userID int64, filter anomaly.Filter) ([]*anomaly.Anomaly, error) {
	paramN := 1
	where := []string{fmt.Sprintf("user_id = $%d", paramN)}
	args := []interface{}{userID}
	paramN++

	if filter.ResourceID != "" {
		where = append(where, fmt.Sprintf("resource_id = $%d", paramN))
		args = append(args, filter.ResourceID)
		paramN++
	}
	if filter.Type != "" {
		where = append(where, fmt.Sprintf("anomaly_type = $%d", paramN))
		args = append(args, filter.Type)
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

	query := fmt.Sprintf(`SELECT id, user_id, resource_id, anomaly_type, severity, percentage, previous_cost, current_cost, detected_at, status FROM cost_anomalies WHERE %s ORDER BY id DESC`, strings.Join(where, " AND "))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.DatabaseError("Failed to list anomalies", err)
	}
	defer rows.Close()

	// Pre-allocate slice with reasonable capacity
	anomalies := make([]*anomaly.Anomaly, 0, 100)
	for rows.Next() {
		var a anomaly.Anomaly
		err := rows.Scan(&a.ID, &a.UserID, &a.ResourceID, &a.AnomalyType, &a.Severity, &a.Percentage, &a.PreviousCost, &a.CurrentCost, &a.DetectedAt, &a.Status)
		if err != nil {
			return nil, errors.DatabaseError("Failed to scan anomaly", err)
		}
		anomalies = append(anomalies, &a)
	}

	return anomalies, rows.Err()
}

func (r *AnomalyRepository) ListWithPagination(ctx context.Context, userID int64, filter anomaly.Filter, limit, offset int) ([]*anomaly.Anomaly, int64, error) {
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
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM cost_anomalies WHERE %s", whereClause), args...).Scan(&total)
	if err != nil {
		return nil, 0, errors.DatabaseError("Failed to count anomalies", err)
	}

	query := fmt.Sprintf(`SELECT id, user_id, resource_id, anomaly_type, severity, percentage, previous_cost, current_cost, detected_at, status FROM cost_anomalies WHERE %s ORDER BY id DESC LIMIT $%d OFFSET $%d`, whereClause, paramN, paramN+1)

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, errors.DatabaseError("Failed to list anomalies", err)
	}
	defer rows.Close()

	// Pre-allocate slice with expected capacity
	anomalies := make([]*anomaly.Anomaly, 0, limit)
	for rows.Next() {
		var a anomaly.Anomaly
		err := rows.Scan(&a.ID, &a.UserID, &a.ResourceID, &a.AnomalyType, &a.Severity, &a.Percentage, &a.PreviousCost, &a.CurrentCost, &a.DetectedAt, &a.Status)
		if err != nil {
			return nil, 0, errors.DatabaseError("Failed to scan anomaly", err)
		}
		anomalies = append(anomalies, &a)
	}

	return anomalies, total, rows.Err()
}

func (r *AnomalyRepository) CountBySeverity(ctx context.Context, userID int64) (map[string]int, error) {
	query := `SELECT severity, COUNT(*) FROM cost_anomalies WHERE user_id = $1 GROUP BY severity`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.DatabaseError("Failed to count anomalies by severity", err)
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
