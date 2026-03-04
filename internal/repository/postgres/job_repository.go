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

// JobRepository implements job.Repository for PostgreSQL
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
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
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
		WHERE id = $1
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
		WHERE user_id = $1 AND job_type = $2
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
		SET schedule = $1, is_enabled = $2, config = $3, last_run = $4, next_run = $5, updated_at = $6
		WHERE id = $7
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
	query := `DELETE FROM scheduled_jobs WHERE id = $1`

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
	paramN := 1
	query := fmt.Sprintf(`
		SELECT id, user_id, job_type, schedule, is_enabled, config, last_run, next_run, created_at, updated_at
		FROM scheduled_jobs
		WHERE user_id = $%d
	`, paramN)
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM scheduled_jobs WHERE user_id = $%d`, paramN)
	args := []interface{}{userID}
	paramN++

	if filter.JobType != "" {
		query += fmt.Sprintf(" AND job_type = $%d", paramN)
		countQuery += fmt.Sprintf(" AND job_type = $%d", paramN)
		args = append(args, string(filter.JobType))
		paramN++
	}
	if filter.IsEnabled != nil {
		query += fmt.Sprintf(" AND is_enabled = $%d", paramN)
		countQuery += fmt.Sprintf(" AND is_enabled = $%d", paramN)
		args = append(args, *filter.IsEnabled)
		paramN++
	}

	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count scheduled jobs: %w", err)
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", paramN, paramN+1)
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
		SET last_run = $1, next_run = $2, updated_at = $3
		WHERE id = $4
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
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
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
		WHERE id = $1
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
		SET status = $1, started_at = $2, completed_at = $3, duration_ms = $4, result = $5, error_message = $6, retry_count = $7
		WHERE id = $8
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
	paramN := 1
	query := `
		SELECT id, job_id, user_id, job_type, status, started_at, completed_at, duration_ms, result, error_message, retry_count, created_at
		FROM job_executions
		WHERE 1=1
	`
	countQuery := `SELECT COUNT(*) FROM job_executions WHERE 1=1`
	var args []interface{}

	if filter.JobID != "" {
		query += fmt.Sprintf(" AND job_id = $%d", paramN)
		countQuery += fmt.Sprintf(" AND job_id = $%d", paramN)
		args = append(args, filter.JobID)
		paramN++
	}
	if filter.UserID > 0 {
		query += fmt.Sprintf(" AND user_id = $%d", paramN)
		countQuery += fmt.Sprintf(" AND user_id = $%d", paramN)
		args = append(args, filter.UserID)
		paramN++
	}
	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", paramN)
		countQuery += fmt.Sprintf(" AND status = $%d", paramN)
		args = append(args, string(filter.Status))
		paramN++
	}
	if filter.JobType != "" {
		query += fmt.Sprintf(" AND job_type = $%d", paramN)
		countQuery += fmt.Sprintf(" AND job_type = $%d", paramN)
		args = append(args, string(filter.JobType))
		paramN++
	}
	if filter.From != nil {
		query += fmt.Sprintf(" AND started_at >= $%d", paramN)
		countQuery += fmt.Sprintf(" AND started_at >= $%d", paramN)
		args = append(args, *filter.From)
		paramN++
	}
	if filter.To != nil {
		query += fmt.Sprintf(" AND started_at <= $%d", paramN)
		countQuery += fmt.Sprintf(" AND started_at <= $%d", paramN)
		args = append(args, *filter.To)
		paramN++
	}

	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count job executions: %w", err)
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", paramN, paramN+1)
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
		WHERE job_id = $1
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
	query := `DELETE FROM job_executions WHERE created_at < $1`

	result, err := r.db.ExecContext(ctx, query, olderThan)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old executions: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected, nil
}
