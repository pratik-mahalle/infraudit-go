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

const anomalySelectCols = `id, user_id, anomaly_type, service_name, region, severity, deviation_percentage, expected_cost, actual_cost, description, detected_at, status`

func scanAnomaly(scan func(dest ...any) error) (*anomaly.Anomaly, error) {
	var a anomaly.Anomaly
	var anomalyType, serviceName, region, description sql.NullString
	err := scan(&a.ID, &a.UserID, &anomalyType, &serviceName, &region, &a.Severity, &a.DeviationPercentage, &a.ExpectedCost, &a.ActualCost, &description, &a.DetectedAt, &a.Status)
	if err != nil {
		return nil, err
	}
	if anomalyType.Valid {
		a.AnomalyType = anomalyType.String
	}
	if serviceName.Valid {
		a.ServiceName = serviceName.String
	}
	if region.Valid {
		a.Region = region.String
	}
	if description.Valid {
		a.Description = description.String
	}
	return &a, nil
}

func (r *AnomalyRepository) Create(ctx context.Context, a *anomaly.Anomaly) (int64, error) {
	now := time.Now()
	a.CreatedAt = now
	a.DetectedAt = now

	query := `INSERT INTO cost_anomalies (user_id, anomaly_type, service_name, region, severity, deviation_percentage, expected_cost, actual_cost, description, detected_at, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id`

	var id int64
	err := r.db.QueryRowContext(ctx, query, a.UserID, a.AnomalyType, a.ServiceName, a.Region, a.Severity, a.DeviationPercentage, a.ExpectedCost, a.ActualCost, a.Description, now, a.Status).Scan(&id)
	if err != nil {
		return 0, errors.DatabaseError("Failed to create anomaly", err)
	}

	return id, nil
}

func (r *AnomalyRepository) GetByID(ctx context.Context, userID int64, id int64) (*anomaly.Anomaly, error) {
	query := fmt.Sprintf(`SELECT %s FROM cost_anomalies WHERE user_id = $1 AND id = $2`, anomalySelectCols)

	a, err := scanAnomaly(r.db.QueryRowContext(ctx, query, userID, id).Scan)
	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Anomaly")
	}
	if err != nil {
		return nil, errors.DatabaseError("Failed to get anomaly", err)
	}

	return a, nil
}

func (r *AnomalyRepository) Update(ctx context.Context, a *anomaly.Anomaly) error {
	a.UpdatedAt = time.Now()
	query := `UPDATE cost_anomalies SET anomaly_type = $1, severity = $2, deviation_percentage = $3, expected_cost = $4, actual_cost = $5, status = $6, description = $7 WHERE user_id = $8 AND id = $9`

	result, err := r.db.ExecContext(ctx, query, a.AnomalyType, a.Severity, a.DeviationPercentage, a.ExpectedCost, a.ActualCost, a.Status, a.Description, a.UserID, a.ID)
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

	query := fmt.Sprintf(`SELECT %s FROM cost_anomalies WHERE %s ORDER BY id DESC`, anomalySelectCols, strings.Join(where, " AND "))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.DatabaseError("Failed to list anomalies", err)
	}
	defer rows.Close()

	anomalies := make([]*anomaly.Anomaly, 0, 100)
	for rows.Next() {
		a, err := scanAnomaly(rows.Scan)
		if err != nil {
			return nil, errors.DatabaseError("Failed to scan anomaly", err)
		}
		anomalies = append(anomalies, a)
	}

	return anomalies, rows.Err()
}

func (r *AnomalyRepository) ListWithPagination(ctx context.Context, userID int64, filter anomaly.Filter, limit, offset int) ([]*anomaly.Anomaly, int64, error) {
	paramN := 1
	where := []string{fmt.Sprintf("user_id = $%d", paramN)}
	args := []interface{}{userID}
	paramN++

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

	query := fmt.Sprintf(`SELECT %s FROM cost_anomalies WHERE %s ORDER BY id DESC LIMIT $%d OFFSET $%d`, anomalySelectCols, whereClause, paramN, paramN+1)

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, errors.DatabaseError("Failed to list anomalies", err)
	}
	defer rows.Close()

	anomalies := make([]*anomaly.Anomaly, 0, limit)
	for rows.Next() {
		a, err := scanAnomaly(rows.Scan)
		if err != nil {
			return nil, 0, errors.DatabaseError("Failed to scan anomaly", err)
		}
		anomalies = append(anomalies, a)
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
