package job

import "context"

// Service defines the job service interface
type Service interface {
	// Scheduled Jobs Management
	CreateJob(ctx context.Context, userID int64, jobType JobType, schedule string, config *JobConfig) (*ScheduledJob, error)
	GetJob(ctx context.Context, id string) (*ScheduledJob, error)
	UpdateJob(ctx context.Context, id string, schedule *string, isEnabled *bool, config *JobConfig) (*ScheduledJob, error)
	DeleteJob(ctx context.Context, id string) error
	ListJobs(ctx context.Context, userID int64, filter Filter, limit, offset int) ([]*ScheduledJob, int64, error)
	EnableJob(ctx context.Context, id string) error
	DisableJob(ctx context.Context, id string) error

	// Manual Job Execution
	TriggerJob(ctx context.Context, id string) (*JobExecution, error)
	TriggerJobByType(ctx context.Context, userID int64, jobType JobType, config *JobConfig) (*JobExecution, error)

	// Job Executions
	GetExecution(ctx context.Context, id string) (*JobExecution, error)
	ListExecutions(ctx context.Context, filter ExecutionFilter, limit, offset int) ([]*JobExecution, int64, error)
	CancelExecution(ctx context.Context, id string) error

	// Scheduler Management
	Start(ctx context.Context) error
	Stop() error
	IsRunning() bool
}
