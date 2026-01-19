package job

import "context"

// Repository defines the job repository interface
type Repository interface {
	// Scheduled Jobs
	CreateJob(ctx context.Context, job *ScheduledJob) error
	GetJob(ctx context.Context, id string) (*ScheduledJob, error)
	GetJobByUserAndType(ctx context.Context, userID int64, jobType JobType) (*ScheduledJob, error)
	UpdateJob(ctx context.Context, job *ScheduledJob) error
	DeleteJob(ctx context.Context, id string) error
	ListJobs(ctx context.Context, userID int64, filter Filter, limit, offset int) ([]*ScheduledJob, int64, error)
	GetEnabledJobs(ctx context.Context) ([]*ScheduledJob, error)
	UpdateLastRun(ctx context.Context, id string, lastRun, nextRun interface{}) error

	// Job Executions
	CreateExecution(ctx context.Context, execution *JobExecution) error
	GetExecution(ctx context.Context, id string) (*JobExecution, error)
	UpdateExecution(ctx context.Context, execution *JobExecution) error
	ListExecutions(ctx context.Context, filter ExecutionFilter, limit, offset int) ([]*JobExecution, int64, error)
	GetLatestExecution(ctx context.Context, jobID string) (*JobExecution, error)
	GetRunningExecutions(ctx context.Context) ([]*JobExecution, error)
	CleanupOldExecutions(ctx context.Context, olderThan interface{}) (int64, error)
}
