package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/pratik-mahalle/infraudit/internal/domain/alert"
	"github.com/pratik-mahalle/infraudit/internal/pkg/errors"
)

type AlertRepository struct {
	db *sql.DB
}

func NewAlertRepository(db *sql.DB) alert.Repository {
	return &AlertRepository{db: db}
}

func (r *AlertRepository) Create(ctx context.Context, a *alert.Alert) (int64, error) {
	now := time.Now()
	a.CreatedAt = now
	a.UpdatedAt = now

	query := `
		INSERT INTO alerts (user_id, type, severity, title, description, resource_name, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRowContext(ctx, query,
		a.UserID, a.Type, a.Severity, a.Title, a.Description, a.Resource, a.Status, now,
	).Scan(&id)
	if err != nil {
		return 0, errors.DatabaseError("Failed to create alert", err)
	}

	return id, nil
}

func (r *AlertRepository) GetByID(ctx context.Context, userID int64, id int64) (*alert.Alert, error) {
	query := `
		SELECT id, user_id, type, severity, title, description, resource_name, status, created_at
		FROM alerts WHERE user_id = $1 AND id = $2
	`

	var a alert.Alert
	err := r.db.QueryRowContext(ctx, query, userID, id).Scan(
		&a.ID, &a.UserID, &a.Type, &a.Severity, &a.Title, &a.Description, &a.Resource, &a.Status, &a.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Alert")
	}
	if err != nil {
		return nil, errors.DatabaseError("Failed to get alert", err)
	}

	return &a, nil
}

func (r *AlertRepository) Update(ctx context.Context, a *alert.Alert) error {
	a.UpdatedAt = time.Now()

	query := `
		UPDATE alerts SET type = $1, severity = $2, title = $3, description = $4, resource_name = $5, status = $6
		WHERE user_id = $7 AND id = $8
	`

	result, err := r.db.ExecContext(ctx, query,
		a.Type, a.Severity, a.Title, a.Description, a.Resource, a.Status, a.UserID, a.ID,
	)
	if err != nil {
		return errors.DatabaseError("Failed to update alert", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.DatabaseError("Failed to get affected rows", err)
	}
	if rows == 0 {
		return errors.NotFound("Alert")
	}

	return nil
}

func (r *AlertRepository) Delete(ctx context.Context, userID int64, id int64) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM alerts WHERE user_id = $1 AND id = $2", userID, id)
	if err != nil {
		return errors.DatabaseError("Failed to delete alert", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.DatabaseError("Failed to get affected rows", err)
	}
	if rows == 0 {
		return errors.NotFound("Alert")
	}

	return nil
}

func (r *AlertRepository) List(ctx context.Context, userID int64, filter alert.Filter) ([]*alert.Alert, error) {
	paramN := 1
	where := []string{fmt.Sprintf("user_id = $%d", paramN)}
	args := []interface{}{userID}
	paramN++

	if filter.Type != "" {
		where = append(where, fmt.Sprintf("type = $%d", paramN))
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
	if filter.Resource != "" {
		where = append(where, fmt.Sprintf("resource_name = $%d", paramN))
		args = append(args, filter.Resource)
		paramN++
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, type, severity, title, description, resource_name, status, created_at
		FROM alerts WHERE %s ORDER BY id DESC
	`, strings.Join(where, " AND "))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.DatabaseError("Failed to list alerts", err)
	}
	defer rows.Close()

	// Pre-allocate slice with reasonable capacity
	alerts := make([]*alert.Alert, 0, 100)
	for rows.Next() {
		var a alert.Alert
		err := rows.Scan(&a.ID, &a.UserID, &a.Type, &a.Severity, &a.Title, &a.Description, &a.Resource, &a.Status, &a.CreatedAt)
		if err != nil {
			return nil, errors.DatabaseError("Failed to scan alert", err)
		}
		alerts = append(alerts, &a)
	}

	return alerts, rows.Err()
}

func (r *AlertRepository) ListWithPagination(ctx context.Context, userID int64, filter alert.Filter, limit, offset int) ([]*alert.Alert, int64, error) {
	paramN := 1
	where := []string{fmt.Sprintf("user_id = $%d", paramN)}
	args := []interface{}{userID}
	paramN++

	if filter.Type != "" {
		where = append(where, fmt.Sprintf("type = $%d", paramN))
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

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM alerts WHERE %s", whereClause)
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, errors.DatabaseError("Failed to count alerts", err)
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, type, severity, title, description, resource_name, status, created_at
		FROM alerts WHERE %s ORDER BY id DESC LIMIT $%d OFFSET $%d
	`, whereClause, paramN, paramN+1)

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, errors.DatabaseError("Failed to list alerts", err)
	}
	defer rows.Close()

	// Pre-allocate slice with expected capacity
	alerts := make([]*alert.Alert, 0, limit)
	for rows.Next() {
		var a alert.Alert
		err := rows.Scan(&a.ID, &a.UserID, &a.Type, &a.Severity, &a.Title, &a.Description, &a.Resource, &a.Status, &a.CreatedAt)
		if err != nil {
			return nil, 0, errors.DatabaseError("Failed to scan alert", err)
		}
		alerts = append(alerts, &a)
	}

	return alerts, total, rows.Err()
}

func (r *AlertRepository) CountByStatus(ctx context.Context, userID int64) (map[string]int, error) {
	query := `SELECT status, COUNT(*) FROM alerts WHERE user_id = $1 GROUP BY status`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.DatabaseError("Failed to count alerts by status", err)
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, errors.DatabaseError("Failed to scan count", err)
		}
		counts[status] = count
	}

	return counts, rows.Err()
}
