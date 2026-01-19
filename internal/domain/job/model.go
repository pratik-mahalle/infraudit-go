package job

import (
	"encoding/json"
	"time"
)

// ScheduledJob represents a scheduled background job
type ScheduledJob struct {
	ID        string          `json:"id"`
	UserID    int64           `json:"user_id"`
	JobType   JobType         `json:"job_type"`
	Schedule  string          `json:"schedule"` // Cron expression
	IsEnabled bool            `json:"is_enabled"`
	Config    json.RawMessage `json:"config,omitempty"`
	LastRun   *time.Time      `json:"last_run,omitempty"`
	NextRun   *time.Time      `json:"next_run,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// JobExecution represents a single execution of a scheduled job
type JobExecution struct {
	ID           string          `json:"id"`
	JobID        string          `json:"job_id"`
	UserID       int64           `json:"user_id"`
	JobType      JobType         `json:"job_type"`
	Status       ExecutionStatus `json:"status"`
	StartedAt    *time.Time      `json:"started_at,omitempty"`
	CompletedAt  *time.Time      `json:"completed_at,omitempty"`
	DurationMs   int             `json:"duration_ms,omitempty"`
	Result       json.RawMessage `json:"result,omitempty"`
	ErrorMessage string          `json:"error_message,omitempty"`
	RetryCount   int             `json:"retry_count"`
	CreatedAt    time.Time       `json:"created_at"`
}

// JobType represents different types of scheduled jobs
type JobType string

const (
	JobTypeResourceSync         JobType = "resource_sync"
	JobTypeDriftDetection       JobType = "drift_detection"
	JobTypeVulnerabilityScan    JobType = "vulnerability_scan"
	JobTypeCostSync             JobType = "cost_sync"
	JobTypeIaCScan              JobType = "iac_scan"
	JobTypeComplianceAssessment JobType = "compliance_assessment"
	JobTypeRecommendation       JobType = "recommendation"
	JobTypeAnomalyDetection     JobType = "anomaly_detection"
)

// ExecutionStatus represents the status of a job execution
type ExecutionStatus string

const (
	ExecutionStatusPending   ExecutionStatus = "pending"
	ExecutionStatusRunning   ExecutionStatus = "running"
	ExecutionStatusCompleted ExecutionStatus = "completed"
	ExecutionStatusFailed    ExecutionStatus = "failed"
	ExecutionStatusCancelled ExecutionStatus = "cancelled"
)

// DefaultSchedules provides default cron schedules for each job type
var DefaultSchedules = map[JobType]string{
	JobTypeResourceSync:         "0 */6 * * *",  // Every 6 hours
	JobTypeDriftDetection:       "0 */4 * * *",  // Every 4 hours
	JobTypeVulnerabilityScan:    "0 2 * * *",    // Daily at 2 AM
	JobTypeCostSync:             "0 1 * * *",    // Daily at 1 AM
	JobTypeIaCScan:              "0 3 * * *",    // Daily at 3 AM
	JobTypeComplianceAssessment: "0 3 * * 0",    // Weekly on Sunday at 3 AM
	JobTypeRecommendation:       "0 4 * * *",    // Daily at 4 AM
	JobTypeAnomalyDetection:     "0 */12 * * *", // Every 12 hours
}

// JobConfig represents job-specific configuration
type JobConfig struct {
	// Common config
	Timeout    int  `json:"timeout_seconds,omitempty"`
	RetryCount int  `json:"retry_count,omitempty"`
	DryRun     bool `json:"dry_run,omitempty"`

	// Provider-specific config
	Providers []string `json:"providers,omitempty"`

	// IaC scan config
	IaCDefinitionID string `json:"iac_definition_id,omitempty"`

	// Additional options
	Options map[string]interface{} `json:"options,omitempty"`
}

// Filter contains job filtering options
type Filter struct {
	JobType   JobType
	IsEnabled *bool
	UserID    int64
}

// ExecutionFilter contains job execution filtering options
type ExecutionFilter struct {
	JobID   string
	UserID  int64
	Status  ExecutionStatus
	JobType JobType
	From    *time.Time
	To      *time.Time
}

// JobResult represents the result of a job execution
type JobResult struct {
	Success       bool                   `json:"success"`
	ItemsScanned  int                    `json:"items_scanned,omitempty"`
	IssuesFound   int                    `json:"issues_found,omitempty"`
	DriftsFound   int                    `json:"drifts_found,omitempty"`
	ErrorCount    int                    `json:"error_count,omitempty"`
	Details       map[string]interface{} `json:"details,omitempty"`
	Notifications []string               `json:"notifications_sent,omitempty"`
}

// IsValid checks if the job type is valid
func (jt JobType) IsValid() bool {
	switch jt {
	case JobTypeResourceSync, JobTypeDriftDetection, JobTypeVulnerabilityScan,
		JobTypeCostSync, JobTypeIaCScan, JobTypeComplianceAssessment,
		JobTypeRecommendation, JobTypeAnomalyDetection:
		return true
	default:
		return false
	}
}

// String returns the string representation of the job type
func (jt JobType) String() string {
	return string(jt)
}

// IsTerminal checks if the execution status is terminal (completed, failed, or cancelled)
func (es ExecutionStatus) IsTerminal() bool {
	return es == ExecutionStatusCompleted || es == ExecutionStatusFailed || es == ExecutionStatusCancelled
}
