package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pratik-mahalle/infraudit/internal/api/dto"
	"github.com/pratik-mahalle/infraudit/internal/domain/job"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
)

// JobHandler handles job-related HTTP requests
type JobHandler struct {
	jobService job.Service
	logger     *logger.Logger
}

// NewJobHandler creates a new job handler
func NewJobHandler(jobService job.Service, log *logger.Logger) *JobHandler {
	return &JobHandler{
		jobService: jobService,
		logger:     log,
	}
}

// ListJobs handles GET /api/v1/jobs
func (h *JobHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Parse query parameters
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	filter := job.Filter{
		UserID: userID,
	}

	if jobType := r.URL.Query().Get("job_type"); jobType != "" {
		filter.JobType = job.JobType(jobType)
	}
	if enabled := r.URL.Query().Get("is_enabled"); enabled != "" {
		isEnabled := enabled == "true"
		filter.IsEnabled = &isEnabled
	}

	jobs, total, err := h.jobService.ListJobs(r.Context(), userID, filter, limit, offset)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to list jobs")
		respondError(w, http.StatusInternalServerError, "failed to list jobs")
		return
	}

	response := dto.ListJobsResponse{
		Jobs:  make([]dto.JobResponse, 0, len(jobs)),
		Total: total,
	}

	for _, j := range jobs {
		response.Jobs = append(response.Jobs, mapJobToResponse(j))
	}

	respondJSON(w, http.StatusOK, response)
}

// CreateJob handles POST /api/v1/jobs
func (h *JobHandler) CreateJob(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.CreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.JobType == "" || req.Schedule == "" {
		respondError(w, http.StatusBadRequest, "job_type and schedule are required")
		return
	}

	jobType := job.JobType(req.JobType)
	if !jobType.IsValid() {
		respondError(w, http.StatusBadRequest, "invalid job type")
		return
	}

	var config *job.JobConfig
	if req.Config != nil {
		configJSON, _ := json.Marshal(req.Config)
		config = &job.JobConfig{}
		json.Unmarshal(configJSON, config)
	}

	j, err := h.jobService.CreateJob(r.Context(), userID, jobType, req.Schedule, config)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to create job")
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, mapJobToResponse(j))
}

// GetJob handles GET /api/v1/jobs/{id}
func (h *JobHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "id")
	if jobID == "" {
		respondError(w, http.StatusBadRequest, "job id is required")
		return
	}

	j, err := h.jobService.GetJob(r.Context(), jobID)
	if err != nil {
		respondError(w, http.StatusNotFound, "job not found")
		return
	}

	respondJSON(w, http.StatusOK, mapJobToResponse(j))
}

// UpdateJob handles PUT /api/v1/jobs/{id}
func (h *JobHandler) UpdateJob(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "id")
	if jobID == "" {
		respondError(w, http.StatusBadRequest, "job id is required")
		return
	}

	var req dto.UpdateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var config *job.JobConfig
	if req.Config != nil {
		configJSON, _ := json.Marshal(req.Config)
		config = &job.JobConfig{}
		json.Unmarshal(configJSON, config)
	}

	j, err := h.jobService.UpdateJob(r.Context(), jobID, req.Schedule, req.IsEnabled, config)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to update job")
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, mapJobToResponse(j))
}

// DeleteJob handles DELETE /api/v1/jobs/{id}
func (h *JobHandler) DeleteJob(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "id")
	if jobID == "" {
		respondError(w, http.StatusBadRequest, "job id is required")
		return
	}

	if err := h.jobService.DeleteJob(r.Context(), jobID); err != nil {
		h.logger.ErrorWithErr(err, "Failed to delete job")
		respondError(w, http.StatusNotFound, "job not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "job deleted"})
}

// TriggerJob handles POST /api/v1/jobs/{id}/run
func (h *JobHandler) TriggerJob(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "id")
	if jobID == "" {
		respondError(w, http.StatusBadRequest, "job id is required")
		return
	}

	execution, err := h.jobService.TriggerJob(r.Context(), jobID)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to trigger job")
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusAccepted, mapJobExecutionToResponse(execution))
}

// ListJobExecutions handles GET /api/v1/jobs/{id}/executions
func (h *JobHandler) ListJobExecutions(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	jobID := chi.URLParam(r, "id")

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	filter := job.ExecutionFilter{
		UserID: userID,
	}

	if jobID != "" {
		filter.JobID = jobID
	}
	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = job.ExecutionStatus(status)
	}

	executions, total, err := h.jobService.ListExecutions(r.Context(), filter, limit, offset)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to list job executions")
		respondError(w, http.StatusInternalServerError, "failed to list executions")
		return
	}

	response := dto.ListJobExecutionsResponse{
		Executions: make([]dto.JobExecutionResponse, 0, len(executions)),
		Total:      total,
	}

	for _, e := range executions {
		response.Executions = append(response.Executions, mapJobExecutionToResponse(e))
	}

	respondJSON(w, http.StatusOK, response)
}

// GetJobExecution handles GET /api/v1/executions/{id}
func (h *JobHandler) GetJobExecution(w http.ResponseWriter, r *http.Request) {
	executionID := chi.URLParam(r, "id")
	if executionID == "" {
		respondError(w, http.StatusBadRequest, "execution id is required")
		return
	}

	execution, err := h.jobService.GetExecution(r.Context(), executionID)
	if err != nil {
		respondError(w, http.StatusNotFound, "execution not found")
		return
	}

	respondJSON(w, http.StatusOK, mapJobExecutionToResponse(execution))
}

// CancelJobExecution handles POST /api/v1/executions/{id}/cancel
func (h *JobHandler) CancelJobExecution(w http.ResponseWriter, r *http.Request) {
	executionID := chi.URLParam(r, "id")
	if executionID == "" {
		respondError(w, http.StatusBadRequest, "execution id is required")
		return
	}

	if err := h.jobService.CancelExecution(r.Context(), executionID); err != nil {
		h.logger.ErrorWithErr(err, "Failed to cancel job execution")
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "execution cancelled"})
}

// GetJobTypes handles GET /api/v1/jobs/types
func (h *JobHandler) GetJobTypes(w http.ResponseWriter, _ *http.Request) {
	types := []map[string]interface{}{
		{"type": "resource_sync", "description": "Sync resources from cloud providers", "default_schedule": job.DefaultSchedules[job.JobTypeResourceSync]},
		{"type": "drift_detection", "description": "Detect configuration drifts", "default_schedule": job.DefaultSchedules[job.JobTypeDriftDetection]},
		{"type": "vulnerability_scan", "description": "Scan for vulnerabilities", "default_schedule": job.DefaultSchedules[job.JobTypeVulnerabilityScan]},
		{"type": "cost_sync", "description": "Sync cost data from providers", "default_schedule": job.DefaultSchedules[job.JobTypeCostSync]},
		{"type": "iac_scan", "description": "Scan Infrastructure as Code", "default_schedule": job.DefaultSchedules[job.JobTypeIaCScan]},
		{"type": "compliance_assessment", "description": "Run compliance assessments", "default_schedule": job.DefaultSchedules[job.JobTypeComplianceAssessment]},
		{"type": "recommendation", "description": "Generate AI recommendations", "default_schedule": job.DefaultSchedules[job.JobTypeRecommendation]},
		{"type": "anomaly_detection", "description": "Detect anomalies", "default_schedule": job.DefaultSchedules[job.JobTypeAnomalyDetection]},
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"job_types": types})
}

// Helper functions

func mapJobToResponse(j *job.ScheduledJob) dto.JobResponse {
	return dto.JobResponse{
		ID:        j.ID,
		UserID:    j.UserID,
		JobType:   string(j.JobType),
		Schedule:  j.Schedule,
		IsEnabled: j.IsEnabled,
		Config:    j.Config,
		LastRun:   j.LastRun,
		NextRun:   j.NextRun,
		CreatedAt: j.CreatedAt,
		UpdatedAt: j.UpdatedAt,
	}
}

func mapJobExecutionToResponse(e *job.JobExecution) dto.JobExecutionResponse {
	return dto.JobExecutionResponse{
		ID:           e.ID,
		JobID:        e.JobID,
		UserID:       e.UserID,
		JobType:      string(e.JobType),
		Status:       string(e.Status),
		StartedAt:    e.StartedAt,
		CompletedAt:  e.CompletedAt,
		DurationMs:   e.DurationMs,
		Result:       e.Result,
		ErrorMessage: e.ErrorMessage,
		CreatedAt:    e.CreatedAt,
	}
}
