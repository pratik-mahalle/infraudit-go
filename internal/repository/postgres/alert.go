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
		INSERT INTO alerts (user_id, type, severity, title, description, resource, status, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		a.UserID, a.Type, a.Severity, a.Title, a.Description, a.Resource, a.Status, now.Format(time.RFC3339),
	)
	if err != nil {
		return 0, errors.DatabaseError("Failed to create alert", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, errors.DatabaseError("Failed to get alert ID", err)
	}

	return id, nil
}

func (r *AlertRepository) GetByID(ctx context.Context, userID int64, id int64) (*alert.Alert, error) {
	query := `
		SELECT id, user_id, type, severity, title, description, resource, status, timestamp
		FROM alerts WHERE user_id = ? AND id = ?
	`

	var a alert.Alert
	var timestamp string
	err := r.db.QueryRowContext(ctx, query, userID, id).Scan(
		&a.ID, &a.UserID, &a.Type, &a.Severity, &a.Title, &a.Description, &a.Resource, &a.Status, &timestamp,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Alert")
	}
	if err != nil {
		return nil, errors.DatabaseError("Failed to get alert", err)
	}

	a.CreatedAt, _ = time.Parse(time.RFC3339, timestamp)
	return &a, nil
}

func (r *AlertRepository) Update(ctx context.Context, a *alert.Alert) error {
	a.UpdatedAt = time.Now()

	query := `
		UPDATE alerts SET type = ?, severity = ?, title = ?, description = ?, resource = ?, status = ?
		WHERE user_id = ? AND id = ?
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
	result, err := r.db.ExecContext(ctx, "DELETE FROM alerts WHERE user_id = ? AND id = ?", userID, id)
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
	where := []string{"user_id = ?"}
	args := []interface{}{userID}

	if filter.Type != "" {
		where = append(where, "type = ?")
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
	if filter.Resource != "" {
		where = append(where, "resource = ?")
		args = append(args, filter.Resource)
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, type, severity, title, description, resource, status, timestamp
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
		var timestamp string
		err := rows.Scan(&a.ID, &a.UserID, &a.Type, &a.Severity, &a.Title, &a.Description, &a.Resource, &a.Status, &timestamp)
		if err != nil {
			return nil, errors.DatabaseError("Failed to scan alert", err)
		}
		a.CreatedAt, _ = time.Parse(time.RFC3339, timestamp)
		alerts = append(alerts, &a)
	}

	return alerts, rows.Err()
}

func (r *AlertRepository) ListWithPagination(ctx context.Context, userID int64, filter alert.Filter, limit, offset int) ([]*alert.Alert, int64, error) {
	where := []string{"user_id = ?"}
	args := []interface{}{userID}

	if filter.Type != "" {
		where = append(where, "type = ?")
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

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM alerts WHERE %s", whereClause)
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, errors.DatabaseError("Failed to count alerts", err)
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, type, severity, title, description, resource, status, timestamp
		FROM alerts WHERE %s ORDER BY id DESC LIMIT ? OFFSET ?
	`, whereClause)

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
		var timestamp string
		err := rows.Scan(&a.ID, &a.UserID, &a.Type, &a.Severity, &a.Title, &a.Description, &a.Resource, &a.Status, &timestamp)
		if err != nil {
			return nil, 0, errors.DatabaseError("Failed to scan alert", err)
		}
		a.CreatedAt, _ = time.Parse(time.RFC3339, timestamp)
		alerts = append(alerts, &a)
	}

	return alerts, total, rows.Err()
}

func (r *AlertRepository) CountByStatus(ctx context.Context, userID int64) (map[string]int, error) {
	query := `SELECT status, COUNT(*) FROM alerts WHERE user_id = ? GROUP BY status`

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
