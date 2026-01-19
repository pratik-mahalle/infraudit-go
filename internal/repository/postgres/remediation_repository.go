package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pratik-mahalle/infraudit/internal/domain/remediation"
)

// RemediationRepository implements remediation.Repository for PostgreSQL/SQLite
type RemediationRepository struct {
	db *sql.DB
}

// NewRemediationRepository creates a new remediation repository
func NewRemediationRepository(db *sql.DB) *RemediationRepository {
	return &RemediationRepository{db: db}
}

// Create creates a new remediation action
func (r *RemediationRepository) Create(ctx context.Context, a *remediation.Action) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}

	strategyJSON, err := json.Marshal(a.Strategy)
	if err != nil {
		return fmt.Errorf("failed to marshal strategy: %w", err)
	}

	resultJSON, _ := json.Marshal(a.Result)
	rollbackJSON, _ := json.Marshal(a.RollbackData)

	query := `
		INSERT INTO remediation_actions (
			id, user_id, drift_id, vulnerability_id, remediation_type, status, 
			strategy, approval_required, approved_by, approved_at, started_at, 
			completed_at, result, rollback_data, error_message, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	_, err = r.db.ExecContext(ctx, query,
		a.ID,
		a.UserID,
		a.DriftID,
		a.VulnerabilityID,
		string(a.RemediationType),
		string(a.Status),
		string(strategyJSON),
		a.ApprovalRequired,
		a.ApprovedBy,
		a.ApprovedAt,
		a.StartedAt,
		a.CompletedAt,
		string(resultJSON),
		string(rollbackJSON),
		a.ErrorMessage,
		now,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to create remediation action: %w", err)
	}

	a.CreatedAt = now
	a.UpdatedAt = now
	return nil
}

// GetByID retrieves a remediation action by ID
func (r *RemediationRepository) GetByID(ctx context.Context, id string) (*remediation.Action, error) {
	query := `
		SELECT id, user_id, drift_id, vulnerability_id, remediation_type, status,
			   strategy, approval_required, approved_by, approved_at, started_at,
			   completed_at, result, rollback_data, error_message, created_at, updated_at
		FROM remediation_actions
		WHERE id = ?
	`

	return r.scanAction(r.db.QueryRowContext(ctx, query, id))
}

// Update updates a remediation action
func (r *RemediationRepository) Update(ctx context.Context, a *remediation.Action) error {
	strategyJSON, err := json.Marshal(a.Strategy)
	if err != nil {
		return fmt.Errorf("failed to marshal strategy: %w", err)
	}

	resultJSON, _ := json.Marshal(a.Result)
	rollbackJSON, _ := json.Marshal(a.RollbackData)

	query := `
		UPDATE remediation_actions
		SET status = ?, strategy = ?, approval_required = ?, approved_by = ?, 
			approved_at = ?, started_at = ?, completed_at = ?, result = ?, 
			rollback_data = ?, error_message = ?, updated_at = ?
		WHERE id = ?
	`

	a.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx, query,
		string(a.Status),
		string(strategyJSON),
		a.ApprovalRequired,
		a.ApprovedBy,
		a.ApprovedAt,
		a.StartedAt,
		a.CompletedAt,
		string(resultJSON),
		string(rollbackJSON),
		a.ErrorMessage,
		a.UpdatedAt,
		a.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update remediation action: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("remediation action not found")
	}

	return nil
}

// Delete deletes a remediation action
func (r *RemediationRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM remediation_actions WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete remediation action: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("remediation action not found")
	}

	return nil
}

// List lists remediation actions with filtering
func (r *RemediationRepository) List(ctx context.Context, filter remediation.Filter, limit, offset int) ([]*remediation.Action, int64, error) {
	query := `
		SELECT id, user_id, drift_id, vulnerability_id, remediation_type, status,
			   strategy, approval_required, approved_by, approved_at, started_at,
			   completed_at, result, rollback_data, error_message, created_at, updated_at
		FROM remediation_actions
		WHERE 1=1
	`
	countQuery := `SELECT COUNT(*) FROM remediation_actions WHERE 1=1`
	var args []interface{}

	if filter.UserID > 0 {
		query += " AND user_id = ?"
		countQuery += " AND user_id = ?"
		args = append(args, filter.UserID)
	}
	if filter.Status != "" {
		query += " AND status = ?"
		countQuery += " AND status = ?"
		args = append(args, string(filter.Status))
	}
	if filter.RemediationType != "" {
		query += " AND remediation_type = ?"
		countQuery += " AND remediation_type = ?"
		args = append(args, string(filter.RemediationType))
	}
	if filter.DriftID != nil {
		query += " AND drift_id = ?"
		countQuery += " AND drift_id = ?"
		args = append(args, *filter.DriftID)
	}
	if filter.VulnerabilityID != nil {
		query += " AND vulnerability_id = ?"
		countQuery += " AND vulnerability_id = ?"
		args = append(args, *filter.VulnerabilityID)
	}
	if filter.From != nil {
		query += " AND created_at >= ?"
		countQuery += " AND created_at >= ?"
		args = append(args, *filter.From)
	}
	if filter.To != nil {
		query += " AND created_at <= ?"
		countQuery += " AND created_at <= ?"
		args = append(args, *filter.To)
	}

	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count remediation actions: %w", err)
	}

	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list remediation actions: %w", err)
	}
	defer rows.Close()

	var actions []*remediation.Action
	for rows.Next() {
		action, err := r.scanActionFromRows(rows)
		if err != nil {
			return nil, 0, err
		}
		actions = append(actions, action)
	}

	return actions, total, nil
}

// GetByDriftID retrieves remediation actions for a drift
func (r *RemediationRepository) GetByDriftID(ctx context.Context, driftID string) ([]*remediation.Action, error) {
	query := `
		SELECT id, user_id, drift_id, vulnerability_id, remediation_type, status,
			   strategy, approval_required, approved_by, approved_at, started_at,
			   completed_at, result, rollback_data, error_message, created_at, updated_at
		FROM remediation_actions
		WHERE drift_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, driftID)
	if err != nil {
		return nil, fmt.Errorf("failed to query remediation actions: %w", err)
	}
	defer rows.Close()

	var actions []*remediation.Action
	for rows.Next() {
		action, err := r.scanActionFromRows(rows)
		if err != nil {
			return nil, err
		}
		actions = append(actions, action)
	}

	return actions, nil
}

// GetByVulnerabilityID retrieves remediation actions for a vulnerability
func (r *RemediationRepository) GetByVulnerabilityID(ctx context.Context, vulnerabilityID string) ([]*remediation.Action, error) {
	query := `
		SELECT id, user_id, drift_id, vulnerability_id, remediation_type, status,
			   strategy, approval_required, approved_by, approved_at, started_at,
			   completed_at, result, rollback_data, error_message, created_at, updated_at
		FROM remediation_actions
		WHERE vulnerability_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, vulnerabilityID)
	if err != nil {
		return nil, fmt.Errorf("failed to query remediation actions: %w", err)
	}
	defer rows.Close()

	var actions []*remediation.Action
	for rows.Next() {
		action, err := r.scanActionFromRows(rows)
		if err != nil {
			return nil, err
		}
		actions = append(actions, action)
	}

	return actions, nil
}

// GetPendingApprovals retrieves pending approval actions for a user
func (r *RemediationRepository) GetPendingApprovals(ctx context.Context, userID int64) ([]*remediation.Action, error) {
	query := `
		SELECT id, user_id, drift_id, vulnerability_id, remediation_type, status,
			   strategy, approval_required, approved_by, approved_at, started_at,
			   completed_at, result, rollback_data, error_message, created_at, updated_at
		FROM remediation_actions
		WHERE user_id = ? AND status = 'pending' AND approval_required = true
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending approvals: %w", err)
	}
	defer rows.Close()

	var actions []*remediation.Action
	for rows.Next() {
		action, err := r.scanActionFromRows(rows)
		if err != nil {
			return nil, err
		}
		actions = append(actions, action)
	}

	return actions, nil
}

// CountByStatus counts remediation actions by status
func (r *RemediationRepository) CountByStatus(ctx context.Context, userID int64) (map[remediation.ActionStatus]int, error) {
	query := `
		SELECT status, COUNT(*) as count
		FROM remediation_actions
		WHERE user_id = ?
		GROUP BY status
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count by status: %w", err)
	}
	defer rows.Close()

	result := make(map[remediation.ActionStatus]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("failed to scan count: %w", err)
		}
		result[remediation.ActionStatus(status)] = count
	}

	return result, nil
}

// scanAction scans a single remediation action from a row
func (r *RemediationRepository) scanAction(row *sql.Row) (*remediation.Action, error) {
	var a remediation.Action
	var driftID, vulnID, approvedBy sql.NullString
	var approvedAt, startedAt, completedAt sql.NullTime
	var strategyStr, resultStr, rollbackStr sql.NullString
	var remType, status string

	err := row.Scan(
		&a.ID,
		&a.UserID,
		&driftID,
		&vulnID,
		&remType,
		&status,
		&strategyStr,
		&a.ApprovalRequired,
		&approvedBy,
		&approvedAt,
		&startedAt,
		&completedAt,
		&resultStr,
		&rollbackStr,
		&a.ErrorMessage,
		&a.CreatedAt,
		&a.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("remediation action not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan remediation action: %w", err)
	}

	a.RemediationType = remediation.RemediationType(remType)
	a.Status = remediation.ActionStatus(status)

	if driftID.Valid {
		a.DriftID = &driftID.String
	}
	if vulnID.Valid {
		a.VulnerabilityID = &vulnID.String
	}
	if approvedBy.Valid {
		// Parse approvedBy as int64
		var approverID int64
		fmt.Sscanf(approvedBy.String, "%d", &approverID)
		a.ApprovedBy = &approverID
	}
	if approvedAt.Valid {
		a.ApprovedAt = &approvedAt.Time
	}
	if startedAt.Valid {
		a.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		a.CompletedAt = &completedAt.Time
	}
	if strategyStr.Valid {
		var strategy remediation.Strategy
		if err := json.Unmarshal([]byte(strategyStr.String), &strategy); err == nil {
			a.Strategy = &strategy
		}
	}
	if resultStr.Valid {
		a.Result = json.RawMessage(resultStr.String)
	}
	if rollbackStr.Valid {
		a.RollbackData = json.RawMessage(rollbackStr.String)
	}

	return &a, nil
}

// scanActionFromRows scans a remediation action from rows
func (r *RemediationRepository) scanActionFromRows(rows *sql.Rows) (*remediation.Action, error) {
	var a remediation.Action
	var driftID, vulnID, approvedBy sql.NullString
	var approvedAt, startedAt, completedAt sql.NullTime
	var strategyStr, resultStr, rollbackStr sql.NullString
	var remType, status string

	err := rows.Scan(
		&a.ID,
		&a.UserID,
		&driftID,
		&vulnID,
		&remType,
		&status,
		&strategyStr,
		&a.ApprovalRequired,
		&approvedBy,
		&approvedAt,
		&startedAt,
		&completedAt,
		&resultStr,
		&rollbackStr,
		&a.ErrorMessage,
		&a.CreatedAt,
		&a.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to scan remediation action: %w", err)
	}

	a.RemediationType = remediation.RemediationType(remType)
	a.Status = remediation.ActionStatus(status)

	if driftID.Valid {
		a.DriftID = &driftID.String
	}
	if vulnID.Valid {
		a.VulnerabilityID = &vulnID.String
	}
	if approvedBy.Valid {
		var approverID int64
		fmt.Sscanf(approvedBy.String, "%d", &approverID)
		a.ApprovedBy = &approverID
	}
	if approvedAt.Valid {
		a.ApprovedAt = &approvedAt.Time
	}
	if startedAt.Valid {
		a.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		a.CompletedAt = &completedAt.Time
	}
	if strategyStr.Valid {
		var strategy remediation.Strategy
		if err := json.Unmarshal([]byte(strategyStr.String), &strategy); err == nil {
			a.Strategy = &strategy
		}
	}
	if resultStr.Valid {
		a.Result = json.RawMessage(resultStr.String)
	}
	if rollbackStr.Valid {
		a.RollbackData = json.RawMessage(rollbackStr.String)
	}

	return &a, nil
}
