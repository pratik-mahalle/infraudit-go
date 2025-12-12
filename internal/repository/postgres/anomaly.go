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

	query := `INSERT INTO cost_anomalies (user_id, resource_id, anomaly_type, severity, percentage, previous_cost, current_cost, detected_at, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.ExecContext(ctx, query, a.UserID, a.ResourceID, a.AnomalyType, a.Severity, a.Percentage, a.PreviousCost, a.CurrentCost, now.Format(time.RFC3339), a.Status)
	if err != nil {
		return 0, errors.DatabaseError("Failed to create anomaly", err)
	}

	return result.LastInsertId()
}

func (r *AnomalyRepository) GetByID(ctx context.Context, userID int64, id int64) (*anomaly.Anomaly, error) {
	query := `SELECT id, user_id, resource_id, anomaly_type, severity, percentage, previous_cost, current_cost, detected_at, status FROM cost_anomalies WHERE user_id = ? AND id = ?`

	var a anomaly.Anomaly
	var detectedAt string
	err := r.db.QueryRowContext(ctx, query, userID, id).Scan(&a.ID, &a.UserID, &a.ResourceID, &a.AnomalyType, &a.Severity, &a.Percentage, &a.PreviousCost, &a.CurrentCost, &detectedAt, &a.Status)

	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Anomaly")
	}
	if err != nil {
		return nil, errors.DatabaseError("Failed to get anomaly", err)
	}

	a.DetectedAt, _ = time.Parse(time.RFC3339, detectedAt)
	return &a, nil
}

func (r *AnomalyRepository) Update(ctx context.Context, a *anomaly.Anomaly) error {
	a.UpdatedAt = time.Now()
	query := `UPDATE cost_anomalies SET resource_id = ?, anomaly_type = ?, severity = ?, percentage = ?, previous_cost = ?, current_cost = ?, status = ? WHERE user_id = ? AND id = ?`

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
	result, err := r.db.ExecContext(ctx, "DELETE FROM cost_anomalies WHERE user_id = ? AND id = ?", userID, id)
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
	where := []string{"user_id = ?"}
	args := []interface{}{userID}

	if filter.ResourceID != "" {
		where = append(where, "resource_id = ?")
		args = append(args, filter.ResourceID)
	}
	if filter.Type != "" {
		where = append(where, "anomaly_type = ?")
		args = append(args, filter.Type)
	}
	if filter.Severity != "" {
		where = append(where, "severity = ?")
		args = append(args, filter.Severity)
	}
	if filter.Status != "" {
		where = append(where, "status = ?")
		args = append(args, filter.Status)
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
		var detectedAt string
		err := rows.Scan(&a.ID, &a.UserID, &a.ResourceID, &a.AnomalyType, &a.Severity, &a.Percentage, &a.PreviousCost, &a.CurrentCost, &detectedAt, &a.Status)
		if err != nil {
			return nil, errors.DatabaseError("Failed to scan anomaly", err)
		}
		a.DetectedAt, _ = time.Parse(time.RFC3339, detectedAt)
		anomalies = append(anomalies, &a)
	}

	return anomalies, rows.Err()
}

func (r *AnomalyRepository) ListWithPagination(ctx context.Context, userID int64, filter anomaly.Filter, limit, offset int) ([]*anomaly.Anomaly, int64, error) {
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
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM cost_anomalies WHERE %s", whereClause), args...).Scan(&total)
	if err != nil {
		return nil, 0, errors.DatabaseError("Failed to count anomalies", err)
	}

	query := fmt.Sprintf(`SELECT id, user_id, resource_id, anomaly_type, severity, percentage, previous_cost, current_cost, detected_at, status FROM cost_anomalies WHERE %s ORDER BY id DESC LIMIT ? OFFSET ?`, whereClause)

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
		var detectedAt string
		err := rows.Scan(&a.ID, &a.UserID, &a.ResourceID, &a.AnomalyType, &a.Severity, &a.Percentage, &a.PreviousCost, &a.CurrentCost, &detectedAt, &a.Status)
		if err != nil {
			return nil, 0, errors.DatabaseError("Failed to scan anomaly", err)
		}
		a.DetectedAt, _ = time.Parse(time.RFC3339, detectedAt)
		anomalies = append(anomalies, &a)
	}

	return anomalies, total, rows.Err()
}

func (r *AnomalyRepository) CountBySeverity(ctx context.Context, userID int64) (map[string]int, error) {
	query := `SELECT severity, COUNT(*) FROM cost_anomalies WHERE user_id = ? GROUP BY severity`

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
