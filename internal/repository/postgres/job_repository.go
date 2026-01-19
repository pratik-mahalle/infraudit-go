package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pratik-mahalle/infraudit/internal/domain/job"
)

// JobRepository implements job.Repository for PostgreSQL/SQLite
type JobRepository struct {
	db *sql.DB
}

// NewJobRepository creates a new job repository
func NewJobRepository(db *sql.DB) *JobRepository {
	return &JobRepository{db: db}
}

// CreateJob creates a new scheduled job
func (r *JobRepository) CreateJob(ctx context.Context, j *job.ScheduledJob) error {
	if j.ID == "" {
		j.ID = uuid.New().String()
	}

	configJSON, err := json.Marshal(j.Config)
	if err != nil {
		configJSON = []byte("{}")
	}

	query := `
		INSERT INTO scheduled_jobs (id, user_id, job_type, schedule, is_enabled, config, last_run, next_run, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	_, err = r.db.ExecContext(ctx, query,
		j.ID,
		j.UserID,
		string(j.JobType),
		j.Schedule,
		j.IsEnabled,
		string(configJSON),
		j.LastRun,
		j.NextRun,
		now,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to create scheduled job: %w", err)
	}

	j.CreatedAt = now
	j.UpdatedAt = now
	return nil
}

// GetJob retrieves a scheduled job by ID
func (r *JobRepository) GetJob(ctx context.Context, id string) (*job.ScheduledJob, error) {
	query := `
		SELECT id, user_id, job_type, schedule, is_enabled, config, last_run, next_run, created_at, updated_at
		FROM scheduled_jobs
		WHERE id = ?
	`

	var j job.ScheduledJob
	var configStr sql.NullString
	var lastRun, nextRun sql.NullTime
	var jobType string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&j.ID,
		&j.UserID,
		&jobType,
		&j.Schedule,
		&j.IsEnabled,
		&configStr,
		&lastRun,
		&nextRun,
		&j.CreatedAt,
		&j.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("scheduled job not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get scheduled job: %w", err)
	}

	j.JobType = job.JobType(jobType)
	if configStr.Valid {
		j.Config = json.RawMessage(configStr.String)
	}
	if lastRun.Valid {
		j.LastRun = &lastRun.Time
	}
	if nextRun.Valid {
		j.NextRun = &nextRun.Time
	}

	return &j, nil
}

// GetJobByUserAndType retrieves a job by user ID and job type
func (r *JobRepository) GetJobByUserAndType(ctx context.Context, userID int64, jobType job.JobType) (*job.ScheduledJob, error) {
	query := `
		SELECT id, user_id, job_type, schedule, is_enabled, config, last_run, next_run, created_at, updated_at
		FROM scheduled_jobs
		WHERE user_id = ? AND job_type = ?
	`

	var j job.ScheduledJob
	var configStr sql.NullString
	var lastRun, nextRun sql.NullTime
	var jt string

	err := r.db.QueryRowContext(ctx, query, userID, string(jobType)).Scan(
		&j.ID,
		&j.UserID,
		&jt,
		&j.Schedule,
		&j.IsEnabled,
		&configStr,
		&lastRun,
		&nextRun,
		&j.CreatedAt,
		&j.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get scheduled job: %w", err)
	}

	j.JobType = job.JobType(jt)
	if configStr.Valid {
		j.Config = json.RawMessage(configStr.String)
	}
	if lastRun.Valid {
		j.LastRun = &lastRun.Time
	}
	if nextRun.Valid {
		j.NextRun = &nextRun.Time
	}

	return &j, nil
}

// UpdateJob updates a scheduled job
func (r *JobRepository) UpdateJob(ctx context.Context, j *job.ScheduledJob) error {
	configJSON, err := json.Marshal(j.Config)
	if err != nil {
		configJSON = []byte("{}")
	}

	query := `
		UPDATE scheduled_jobs
		SET schedule = ?, is_enabled = ?, config = ?, last_run = ?, next_run = ?, updated_at = ?
		WHERE id = ?
	`

	j.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx, query,
		j.Schedule,
		j.IsEnabled,
		string(configJSON),
		j.LastRun,
		j.NextRun,
		j.UpdatedAt,
		j.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update scheduled job: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("scheduled job not found")
	}

	return nil
}

// DeleteJob deletes a scheduled job
func (r *JobRepository) DeleteJob(ctx context.Context, id string) error {
	query := `DELETE FROM scheduled_jobs WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete scheduled job: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("scheduled job not found")
	}

	return nil
}

// ListJobs lists scheduled jobs with filtering
func (r *JobRepository) ListJobs(ctx context.Context, userID int64, filter job.Filter, limit, offset int) ([]*job.ScheduledJob, int64, error) {
	query := `
		SELECT id, user_id, job_type, schedule, is_enabled, config, last_run, next_run, created_at, updated_at
		FROM scheduled_jobs
		WHERE user_id = ?
	`
	countQuery := `SELECT COUNT(*) FROM scheduled_jobs WHERE user_id = ?`
	args := []interface{}{userID}

	if filter.JobType != "" {
		query += " AND job_type = ?"
		countQuery += " AND job_type = ?"
		args = append(args, string(filter.JobType))
	}
	if filter.IsEnabled != nil {
		query += " AND is_enabled = ?"
		countQuery += " AND is_enabled = ?"
		args = append(args, *filter.IsEnabled)
	}

	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count scheduled jobs: %w", err)
	}

	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list scheduled jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*job.ScheduledJob
	for rows.Next() {
		var j job.ScheduledJob
		var configStr sql.NullString
		var lastRun, nextRun sql.NullTime
		var jobType string

		err := rows.Scan(
			&j.ID,
			&j.UserID,
			&jobType,
			&j.Schedule,
			&j.IsEnabled,
			&configStr,
			&lastRun,
			&nextRun,
			&j.CreatedAt,
			&j.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan scheduled job: %w", err)
		}

		j.JobType = job.JobType(jobType)
		if configStr.Valid {
			j.Config = json.RawMessage(configStr.String)
		}
		if lastRun.Valid {
			j.LastRun = &lastRun.Time
		}
		if nextRun.Valid {
			j.NextRun = &nextRun.Time
		}

		jobs = append(jobs, &j)
	}

	return jobs, total, nil
}

// GetEnabledJobs retrieves all enabled jobs
func (r *JobRepository) GetEnabledJobs(ctx context.Context) ([]*job.ScheduledJob, error) {
	query := `
		SELECT id, user_id, job_type, schedule, is_enabled, config, last_run, next_run, created_at, updated_at
		FROM scheduled_jobs
		WHERE is_enabled = true
		ORDER BY next_run ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*job.ScheduledJob
	for rows.Next() {
		var j job.ScheduledJob
		var configStr sql.NullString
		var lastRun, nextRun sql.NullTime
		var jobType string

		err := rows.Scan(
			&j.ID,
			&j.UserID,
			&jobType,
			&j.Schedule,
			&j.IsEnabled,
			&configStr,
			&lastRun,
			&nextRun,
			&j.CreatedAt,
			&j.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan scheduled job: %w", err)
		}

		j.JobType = job.JobType(jobType)
		if configStr.Valid {
			j.Config = json.RawMessage(configStr.String)
		}
		if lastRun.Valid {
			j.LastRun = &lastRun.Time
		}
		if nextRun.Valid {
			j.NextRun = &nextRun.Time
		}

		jobs = append(jobs, &j)
	}

	return jobs, nil
}

// UpdateLastRun updates the last run and next run times
func (r *JobRepository) UpdateLastRun(ctx context.Context, id string, lastRun, nextRun interface{}) error {
	query := `
		UPDATE scheduled_jobs
		SET last_run = ?, next_run = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query, lastRun, nextRun, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update last run: %w", err)
	}

	return nil
}

// CreateExecution creates a new job execution
func (r *JobRepository) CreateExecution(ctx context.Context, e *job.JobExecution) error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}

	resultJSON, err := json.Marshal(e.Result)
	if err != nil {
		resultJSON = []byte("{}")
	}

	query := `
		INSERT INTO job_executions (id, job_id, user_id, job_type, status, started_at, completed_at, duration_ms, result, error_message, retry_count, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	_, err = r.db.ExecContext(ctx, query,
		e.ID,
		e.JobID,
		e.UserID,
		string(e.JobType),
		string(e.Status),
		e.StartedAt,
		e.CompletedAt,
		e.DurationMs,
		string(resultJSON),
		e.ErrorMessage,
		e.RetryCount,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to create job execution: %w", err)
	}

	e.CreatedAt = now
	return nil
}

// GetExecution retrieves a job execution by ID
func (r *JobRepository) GetExecution(ctx context.Context, id string) (*job.JobExecution, error) {
	query := `
		SELECT id, job_id, user_id, job_type, status, started_at, completed_at, duration_ms, result, error_message, retry_count, created_at
		FROM job_executions
		WHERE id = ?
	`

	var e job.JobExecution
	var resultStr sql.NullString
	var startedAt, completedAt sql.NullTime
	var jobType, status string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&e.ID,
		&e.JobID,
		&e.UserID,
		&jobType,
		&status,
		&startedAt,
		&completedAt,
		&e.DurationMs,
		&resultStr,
		&e.ErrorMessage,
		&e.RetryCount,
		&e.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("job execution not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get job execution: %w", err)
	}

	e.JobType = job.JobType(jobType)
	e.Status = job.ExecutionStatus(status)
	if resultStr.Valid {
		e.Result = json.RawMessage(resultStr.String)
	}
	if startedAt.Valid {
		e.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		e.CompletedAt = &completedAt.Time
	}

	return &e, nil
}

// UpdateExecution updates a job execution
func (r *JobRepository) UpdateExecution(ctx context.Context, e *job.JobExecution) error {
	resultJSON, err := json.Marshal(e.Result)
	if err != nil {
		resultJSON = []byte("{}")
	}

	query := `
		UPDATE job_executions
		SET status = ?, started_at = ?, completed_at = ?, duration_ms = ?, result = ?, error_message = ?, retry_count = ?
		WHERE id = ?
	`

	_, err = r.db.ExecContext(ctx, query,
		string(e.Status),
		e.StartedAt,
		e.CompletedAt,
		e.DurationMs,
		string(resultJSON),
		e.ErrorMessage,
		e.RetryCount,
		e.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update job execution: %w", err)
	}

	return nil
}

// ListExecutions lists job executions with filtering
func (r *JobRepository) ListExecutions(ctx context.Context, filter job.ExecutionFilter, limit, offset int) ([]*job.JobExecution, int64, error) {
	query := `
		SELECT id, job_id, user_id, job_type, status, started_at, completed_at, duration_ms, result, error_message, retry_count, created_at
		FROM job_executions
		WHERE 1=1
	`
	countQuery := `SELECT COUNT(*) FROM job_executions WHERE 1=1`
	var args []interface{}

	if filter.JobID != "" {
		query += " AND job_id = ?"
		countQuery += " AND job_id = ?"
		args = append(args, filter.JobID)
	}
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
	if filter.JobType != "" {
		query += " AND job_type = ?"
		countQuery += " AND job_type = ?"
		args = append(args, string(filter.JobType))
	}
	if filter.From != nil {
		query += " AND started_at >= ?"
		countQuery += " AND started_at >= ?"
		args = append(args, *filter.From)
	}
	if filter.To != nil {
		query += " AND started_at <= ?"
		countQuery += " AND started_at <= ?"
		args = append(args, *filter.To)
	}

	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count job executions: %w", err)
	}

	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list job executions: %w", err)
	}
	defer rows.Close()

	var executions []*job.JobExecution
	for rows.Next() {
		var e job.JobExecution
		var resultStr sql.NullString
		var startedAt, completedAt sql.NullTime
		var jobType, status string

		err := rows.Scan(
			&e.ID,
			&e.JobID,
			&e.UserID,
			&jobType,
			&status,
			&startedAt,
			&completedAt,
			&e.DurationMs,
			&resultStr,
			&e.ErrorMessage,
			&e.RetryCount,
			&e.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan job execution: %w", err)
		}

		e.JobType = job.JobType(jobType)
		e.Status = job.ExecutionStatus(status)
		if resultStr.Valid {
			e.Result = json.RawMessage(resultStr.String)
		}
		if startedAt.Valid {
			e.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			e.CompletedAt = &completedAt.Time
		}

		executions = append(executions, &e)
	}

	return executions, total, nil
}

// GetLatestExecution retrieves the latest execution for a job
func (r *JobRepository) GetLatestExecution(ctx context.Context, jobID string) (*job.JobExecution, error) {
	query := `
		SELECT id, job_id, user_id, job_type, status, started_at, completed_at, duration_ms, result, error_message, retry_count, created_at
		FROM job_executions
		WHERE job_id = ?
		ORDER BY created_at DESC
		LIMIT 1
	`

	var e job.JobExecution
	var resultStr sql.NullString
	var startedAt, completedAt sql.NullTime
	var jobType, status string

	err := r.db.QueryRowContext(ctx, query, jobID).Scan(
		&e.ID,
		&e.JobID,
		&e.UserID,
		&jobType,
		&status,
		&startedAt,
		&completedAt,
		&e.DurationMs,
		&resultStr,
		&e.ErrorMessage,
		&e.RetryCount,
		&e.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest execution: %w", err)
	}

	e.JobType = job.JobType(jobType)
	e.Status = job.ExecutionStatus(status)
	if resultStr.Valid {
		e.Result = json.RawMessage(resultStr.String)
	}
	if startedAt.Valid {
		e.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		e.CompletedAt = &completedAt.Time
	}

	return &e, nil
}

// GetRunningExecutions retrieves all running executions
func (r *JobRepository) GetRunningExecutions(ctx context.Context) ([]*job.JobExecution, error) {
	query := `
		SELECT id, job_id, user_id, job_type, status, started_at, completed_at, duration_ms, result, error_message, retry_count, created_at
		FROM job_executions
		WHERE status = 'running'
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get running executions: %w", err)
	}
	defer rows.Close()

	var executions []*job.JobExecution
	for rows.Next() {
		var e job.JobExecution
		var resultStr sql.NullString
		var startedAt, completedAt sql.NullTime
		var jobType, status string

		err := rows.Scan(
			&e.ID,
			&e.JobID,
			&e.UserID,
			&jobType,
			&status,
			&startedAt,
			&completedAt,
			&e.DurationMs,
			&resultStr,
			&e.ErrorMessage,
			&e.RetryCount,
			&e.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job execution: %w", err)
		}

		e.JobType = job.JobType(jobType)
		e.Status = job.ExecutionStatus(status)
		if resultStr.Valid {
			e.Result = json.RawMessage(resultStr.String)
		}
		if startedAt.Valid {
			e.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			e.CompletedAt = &completedAt.Time
		}

		executions = append(executions, &e)
	}

	return executions, nil
}

// CleanupOldExecutions removes executions older than the specified time
func (r *JobRepository) CleanupOldExecutions(ctx context.Context, olderThan interface{}) (int64, error) {
	query := `DELETE FROM job_executions WHERE created_at < ?`

	result, err := r.db.ExecContext(ctx, query, olderThan)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old executions: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected, nil
}
