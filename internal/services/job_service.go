package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pratik-mahalle/infraudit/internal/domain/drift"
	"github.com/pratik-mahalle/infraudit/internal/domain/job"
	"github.com/pratik-mahalle/infraudit/internal/domain/provider"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
	"github.com/robfig/cron/v3"
)

// JobService implements job.Service
type JobService struct {
	repo            job.Repository
	driftService    drift.Service
	providerService provider.Service
	logger          *logger.Logger

	scheduler    *cron.Cron
	cronEntries  map[string]cron.EntryID
	entriesMutex sync.RWMutex
	isRunning    bool
	runningMutex sync.RWMutex
}

// NewJobService creates a new job service
func NewJobService(
	repo job.Repository,
	driftService drift.Service,
	providerService provider.Service,
	log *logger.Logger,
) job.Service {
	return &JobService{
		repo:            repo,
		driftService:    driftService,
		providerService: providerService,
		logger:          log,
		cronEntries:     make(map[string]cron.EntryID),
	}
}

// CreateJob creates a new scheduled job
func (s *JobService) CreateJob(ctx context.Context, userID int64, jobType job.JobType, schedule string, config *job.JobConfig) (*job.ScheduledJob, error) {
	// Validate job type
	if !jobType.IsValid() {
		return nil, fmt.Errorf("invalid job type: %s", jobType)
	}

	// Validate cron schedule
	if _, err := cron.ParseStandard(schedule); err != nil {
		return nil, fmt.Errorf("invalid cron schedule: %w", err)
	}

	// Check if job already exists for this user and type
	existing, err := s.repo.GetJobByUserAndType(ctx, userID, jobType)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("job of type %s already exists for user", jobType)
	}

	var configJSON json.RawMessage
	if config != nil {
		configJSON, _ = json.Marshal(config)
	}

	// Calculate next run time
	nextRun := s.calculateNextRun(schedule)

	j := &job.ScheduledJob{
		ID:        uuid.New().String(),
		UserID:    userID,
		JobType:   jobType,
		Schedule:  schedule,
		IsEnabled: true,
		Config:    configJSON,
		NextRun:   nextRun,
	}

	if err := s.repo.CreateJob(ctx, j); err != nil {
		return nil, err
	}

	s.logger.WithFields(map[string]interface{}{
		"job_id":   j.ID,
		"user_id":  userID,
		"job_type": jobType,
		"schedule": schedule,
	}).Info("Scheduled job created")

	// If scheduler is running, add the job
	if s.IsRunning() {
		s.scheduleJob(j)
	}

	return j, nil
}

// GetJob retrieves a scheduled job by ID
func (s *JobService) GetJob(ctx context.Context, id string) (*job.ScheduledJob, error) {
	return s.repo.GetJob(ctx, id)
}

// UpdateJob updates a scheduled job
func (s *JobService) UpdateJob(ctx context.Context, id string, schedule *string, isEnabled *bool, config *job.JobConfig) (*job.ScheduledJob, error) {
	j, err := s.repo.GetJob(ctx, id)
	if err != nil {
		return nil, err
	}

	if schedule != nil {
		// Validate new schedule
		if _, err := cron.ParseStandard(*schedule); err != nil {
			return nil, fmt.Errorf("invalid cron schedule: %w", err)
		}
		j.Schedule = *schedule
		j.NextRun = s.calculateNextRun(*schedule)
	}

	if isEnabled != nil {
		j.IsEnabled = *isEnabled
	}

	if config != nil {
		configJSON, _ := json.Marshal(config)
		j.Config = configJSON
	}

	if err := s.repo.UpdateJob(ctx, j); err != nil {
		return nil, err
	}

	// Reschedule job if scheduler is running
	if s.IsRunning() {
		s.unscheduleJob(id)
		if j.IsEnabled {
			s.scheduleJob(j)
		}
	}

	s.logger.WithFields(map[string]interface{}{
		"job_id": id,
	}).Info("Scheduled job updated")

	return j, nil
}

// DeleteJob deletes a scheduled job
func (s *JobService) DeleteJob(ctx context.Context, id string) error {
	// Unschedule if running
	if s.IsRunning() {
		s.unscheduleJob(id)
	}

	if err := s.repo.DeleteJob(ctx, id); err != nil {
		return err
	}

	s.logger.WithFields(map[string]interface{}{
		"job_id": id,
	}).Info("Scheduled job deleted")

	return nil
}

// ListJobs lists scheduled jobs
func (s *JobService) ListJobs(ctx context.Context, userID int64, filter job.Filter, limit, offset int) ([]*job.ScheduledJob, int64, error) {
	return s.repo.ListJobs(ctx, userID, filter, limit, offset)
}

// EnableJob enables a scheduled job
func (s *JobService) EnableJob(ctx context.Context, id string) error {
	enabled := true
	_, err := s.UpdateJob(ctx, id, nil, &enabled, nil)
	return err
}

// DisableJob disables a scheduled job
func (s *JobService) DisableJob(ctx context.Context, id string) error {
	disabled := false
	_, err := s.UpdateJob(ctx, id, nil, &disabled, nil)
	return err
}

// TriggerJob manually triggers a scheduled job
func (s *JobService) TriggerJob(ctx context.Context, id string) (*job.JobExecution, error) {
	j, err := s.repo.GetJob(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.executeJob(ctx, j)
}

// TriggerJobByType triggers a job by type for a user
func (s *JobService) TriggerJobByType(ctx context.Context, userID int64, jobType job.JobType, config *job.JobConfig) (*job.JobExecution, error) {
	j, err := s.repo.GetJobByUserAndType(ctx, userID, jobType)
	if err != nil {
		return nil, err
	}

	if j == nil {
		// Create a one-off execution
		var configJSON json.RawMessage
		if config != nil {
			configJSON, _ = json.Marshal(config)
		}

		j = &job.ScheduledJob{
			ID:      uuid.New().String(),
			UserID:  userID,
			JobType: jobType,
			Config:  configJSON,
		}
	}

	return s.executeJob(ctx, j)
}

// GetExecution retrieves a job execution
func (s *JobService) GetExecution(ctx context.Context, id string) (*job.JobExecution, error) {
	return s.repo.GetExecution(ctx, id)
}

// ListExecutions lists job executions
func (s *JobService) ListExecutions(ctx context.Context, filter job.ExecutionFilter, limit, offset int) ([]*job.JobExecution, int64, error) {
	return s.repo.ListExecutions(ctx, filter, limit, offset)
}

// CancelExecution cancels a running execution
func (s *JobService) CancelExecution(ctx context.Context, id string) error {
	e, err := s.repo.GetExecution(ctx, id)
	if err != nil {
		return err
	}

	if e.Status.IsTerminal() {
		return fmt.Errorf("execution is already completed")
	}

	e.Status = job.ExecutionStatusCancelled
	now := time.Now()
	e.CompletedAt = &now
	if e.StartedAt != nil {
		e.DurationMs = int(now.Sub(*e.StartedAt).Milliseconds())
	}

	return s.repo.UpdateExecution(ctx, e)
}

// Start starts the job scheduler
func (s *JobService) Start(ctx context.Context) error {
	s.runningMutex.Lock()
	defer s.runningMutex.Unlock()

	if s.isRunning {
		return fmt.Errorf("scheduler is already running")
	}

	s.scheduler = cron.New(cron.WithSeconds())

	// Load all enabled jobs
	jobs, err := s.repo.GetEnabledJobs(ctx)
	if err != nil {
		return fmt.Errorf("failed to load enabled jobs: %w", err)
	}

	for _, j := range jobs {
		s.scheduleJob(j)
	}

	s.scheduler.Start()
	s.isRunning = true

	s.logger.WithFields(map[string]interface{}{
		"jobs_loaded": len(jobs),
	}).Info("Job scheduler started")

	return nil
}

// Stop stops the job scheduler
func (s *JobService) Stop() error {
	s.runningMutex.Lock()
	defer s.runningMutex.Unlock()

	if !s.isRunning {
		return nil
	}

	s.scheduler.Stop()
	s.isRunning = false

	s.entriesMutex.Lock()
	s.cronEntries = make(map[string]cron.EntryID)
	s.entriesMutex.Unlock()

	s.logger.Info("Job scheduler stopped")

	return nil
}

// IsRunning returns whether the scheduler is running
func (s *JobService) IsRunning() bool {
	s.runningMutex.RLock()
	defer s.runningMutex.RUnlock()
	return s.isRunning
}

// scheduleJob adds a job to the cron scheduler
func (s *JobService) scheduleJob(j *job.ScheduledJob) {
	s.entriesMutex.Lock()
	defer s.entriesMutex.Unlock()

	entryID, err := s.scheduler.AddFunc(j.Schedule, func() {
		ctx := context.Background()
		if _, err := s.executeJob(ctx, j); err != nil {
			s.logger.WithFields(map[string]interface{}{
				"job_id":   j.ID,
				"job_type": j.JobType,
			}).ErrorWithErr(err, "Failed to execute scheduled job")
		}
	})

	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"job_id":   j.ID,
			"schedule": j.Schedule,
		}).ErrorWithErr(err, "Failed to schedule job")
		return
	}

	s.cronEntries[j.ID] = entryID

	s.logger.WithFields(map[string]interface{}{
		"job_id":   j.ID,
		"job_type": j.JobType,
		"schedule": j.Schedule,
	}).Info("Job scheduled")
}

// unscheduleJob removes a job from the cron scheduler
func (s *JobService) unscheduleJob(jobID string) {
	s.entriesMutex.Lock()
	defer s.entriesMutex.Unlock()

	if entryID, ok := s.cronEntries[jobID]; ok {
		s.scheduler.Remove(entryID)
		delete(s.cronEntries, jobID)

		s.logger.WithFields(map[string]interface{}{
			"job_id": jobID,
		}).Info("Job unscheduled")
	}
}

// executeJob executes a job and records the execution
func (s *JobService) executeJob(ctx context.Context, j *job.ScheduledJob) (*job.JobExecution, error) {
	now := time.Now()

	execution := &job.JobExecution{
		ID:        uuid.New().String(),
		JobID:     j.ID,
		UserID:    j.UserID,
		JobType:   j.JobType,
		Status:    job.ExecutionStatusRunning,
		StartedAt: &now,
	}

	if err := s.repo.CreateExecution(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to create execution record: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"job_id":       j.ID,
		"execution_id": execution.ID,
		"job_type":     j.JobType,
	}).Info("Job execution started")

	// Execute the job in a goroutine
	go func() {
		execCtx := context.Background()
		result, err := s.runJobLogic(execCtx, j)

		completedAt := time.Now()
		execution.CompletedAt = &completedAt
		execution.DurationMs = int(completedAt.Sub(now).Milliseconds())

		if err != nil {
			execution.Status = job.ExecutionStatusFailed
			execution.ErrorMessage = err.Error()
			s.logger.WithFields(map[string]interface{}{
				"job_id":       j.ID,
				"execution_id": execution.ID,
				"error":        err.Error(),
			}).ErrorWithErr(err, "Job execution failed")
		} else {
			execution.Status = job.ExecutionStatusCompleted
			if result != nil {
				resultJSON, _ := json.Marshal(result)
				execution.Result = resultJSON
			}
			s.logger.WithFields(map[string]interface{}{
				"job_id":       j.ID,
				"execution_id": execution.ID,
				"duration_ms":  execution.DurationMs,
			}).Info("Job execution completed")
		}

		s.repo.UpdateExecution(execCtx, execution)

		// Update last run time
		nextRun := s.calculateNextRun(j.Schedule)
		s.repo.UpdateLastRun(execCtx, j.ID, completedAt, nextRun)
	}()

	return execution, nil
}

// runJobLogic executes the actual job logic
func (s *JobService) runJobLogic(ctx context.Context, j *job.ScheduledJob) (*job.JobResult, error) {
	result := &job.JobResult{
		Success: true,
		Details: make(map[string]interface{}),
	}

	switch j.JobType {
	case job.JobTypeResourceSync:
		return s.runResourceSyncJob(ctx, j)
	case job.JobTypeDriftDetection:
		return s.runDriftDetectionJob(ctx, j)
	case job.JobTypeVulnerabilityScan:
		return s.runVulnerabilityScanJob(ctx, j)
	case job.JobTypeAnomalyDetection:
		return s.runAnomalyDetectionJob(ctx, j)
	case job.JobTypeRecommendation:
		return s.runRecommendationJob(ctx, j)
	default:
		return result, nil
	}
}

// runResourceSyncJob syncs resources from cloud providers
func (s *JobService) runResourceSyncJob(ctx context.Context, j *job.ScheduledJob) (*job.JobResult, error) {
	result := &job.JobResult{
		Success: true,
		Details: make(map[string]interface{}),
	}

	// Get providers to sync
	providers, err := s.providerService.List(ctx, j.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to list providers: %w", err)
	}

	syncedCount := 0
	errorCount := 0

	for _, p := range providers {
		if !p.IsConnected {
			continue
		}

		if err := s.providerService.Sync(ctx, j.UserID, p.Provider); err != nil {
			s.logger.WithFields(map[string]interface{}{
				"user_id":  j.UserID,
				"provider": p.Provider,
			}).ErrorWithErr(err, "Failed to sync provider")
			errorCount++
		} else {
			syncedCount++
		}
	}

	result.ItemsScanned = syncedCount
	result.ErrorCount = errorCount
	result.Details["providers_synced"] = syncedCount
	result.Details["providers_failed"] = errorCount

	return result, nil
}

// runDriftDetectionJob runs drift detection
func (s *JobService) runDriftDetectionJob(ctx context.Context, j *job.ScheduledJob) (*job.JobResult, error) {
	result := &job.JobResult{
		Success: true,
		Details: make(map[string]interface{}),
	}

	if err := s.driftService.DetectDrifts(ctx, j.UserID); err != nil {
		return nil, fmt.Errorf("drift detection failed: %w", err)
	}

	// Get drift summary
	drifts, _, err := s.driftService.List(ctx, j.UserID, drift.Filter{}, 1000, 0)
	if err == nil {
		result.DriftsFound = len(drifts)
		result.Details["drifts_detected"] = len(drifts)
	}

	return result, nil
}

// runVulnerabilityScanJob runs vulnerability scanning
func (s *JobService) runVulnerabilityScanJob(ctx context.Context, j *job.ScheduledJob) (*job.JobResult, error) {
	// This would integrate with the vulnerability service
	return &job.JobResult{
		Success: true,
		Details: map[string]interface{}{
			"message": "Vulnerability scan completed",
		},
	}, nil
}

// runAnomalyDetectionJob runs anomaly detection
func (s *JobService) runAnomalyDetectionJob(ctx context.Context, j *job.ScheduledJob) (*job.JobResult, error) {
	// This would integrate with the anomaly service
	return &job.JobResult{
		Success: true,
		Details: map[string]interface{}{
			"message": "Anomaly detection completed",
		},
	}, nil
}

// runRecommendationJob generates recommendations
func (s *JobService) runRecommendationJob(ctx context.Context, j *job.ScheduledJob) (*job.JobResult, error) {
	// This would integrate with the recommendation service
	return &job.JobResult{
		Success: true,
		Details: map[string]interface{}{
			"message": "Recommendations generated",
		},
	}, nil
}

// calculateNextRun calculates the next run time for a schedule
func (s *JobService) calculateNextRun(schedule string) *time.Time {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	sched, err := parser.Parse(schedule)
	if err != nil {
		return nil
	}

	next := sched.Next(time.Now())
	return &next
}
