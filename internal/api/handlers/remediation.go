package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pratik-mahalle/infraudit/internal/api/dto"
	"github.com/pratik-mahalle/infraudit/internal/domain/remediation"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
)

// RemediationHandler handles remediation-related HTTP requests
type RemediationHandler struct {
	remediationService remediation.Service
	logger             *logger.Logger
}

// NewRemediationHandler creates a new remediation handler
func NewRemediationHandler(remediationService remediation.Service, log *logger.Logger) *RemediationHandler {
	return &RemediationHandler{
		remediationService: remediationService,
		logger:             log,
	}
}

// SuggestForDrift handles POST /api/v1/remediation/suggest/drift/{id}
func (h *RemediationHandler) SuggestForDrift(w http.ResponseWriter, r *http.Request) {
	driftID := chi.URLParam(r, "id")
	if driftID == "" {
		respondError(w, http.StatusBadRequest, "drift id is required")
		return
	}

	suggestions, err := h.remediationService.SuggestForDrift(r.Context(), driftID)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to generate suggestions for drift")
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := make([]dto.RemediationSuggestionResponse, 0, len(suggestions))
	for _, s := range suggestions {
		response = append(response, mapSuggestionToResponse(s))
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"suggestions": response})
}

// SuggestForVulnerability handles POST /api/v1/remediation/suggest/vulnerability/{id}
func (h *RemediationHandler) SuggestForVulnerability(w http.ResponseWriter, r *http.Request) {
	vulnID := chi.URLParam(r, "id")
	if vulnID == "" {
		respondError(w, http.StatusBadRequest, "vulnerability id is required")
		return
	}

	suggestions, err := h.remediationService.SuggestForVulnerability(r.Context(), vulnID)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to generate suggestions for vulnerability")
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := make([]dto.RemediationSuggestionResponse, 0, len(suggestions))
	for _, s := range suggestions {
		response = append(response, mapSuggestionToResponse(s))
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"suggestions": response})
}

// CreateAction handles POST /api/v1/remediation/actions
func (h *RemediationHandler) CreateAction(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		IssueType       string                      `json:"issue_type"`
		IssueID         string                      `json:"issue_id"`
		RemediationType string                      `json:"remediation_type"`
		Strategy        *dto.RemediationStrategyDTO `json:"strategy,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Create suggestion from request
	suggestion := &remediation.Suggestion{
		IssueType:       req.IssueType,
		IssueID:         req.IssueID,
		RemediationType: remediation.RemediationType(req.RemediationType),
	}

	if req.Strategy != nil {
		suggestion.Strategy = mapDTOToStrategy(req.Strategy)
	}

	action, err := h.remediationService.Create(r.Context(), userID, suggestion)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to create remediation action")
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, mapActionToResponse(action))
}

// ListActions handles GET /api/v1/remediation/actions
func (h *RemediationHandler) ListActions(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	filter := remediation.Filter{
		UserID: userID,
	}

	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = remediation.ActionStatus(status)
	}
	if remType := r.URL.Query().Get("type"); remType != "" {
		filter.RemediationType = remediation.RemediationType(remType)
	}

	actions, total, err := h.remediationService.ListActions(r.Context(), filter, limit, offset)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to list remediation actions")
		respondError(w, http.StatusInternalServerError, "failed to list actions")
		return
	}

	response := dto.ListRemediationActionsResponse{
		Actions: make([]dto.RemediationActionResponse, 0, len(actions)),
		Total:   total,
	}

	for _, a := range actions {
		response.Actions = append(response.Actions, mapActionToResponse(a))
	}

	respondJSON(w, http.StatusOK, response)
}

// GetAction handles GET /api/v1/remediation/actions/{id}
func (h *RemediationHandler) GetAction(w http.ResponseWriter, r *http.Request) {
	actionID := chi.URLParam(r, "id")
	if actionID == "" {
		respondError(w, http.StatusBadRequest, "action id is required")
		return
	}

	action, err := h.remediationService.GetAction(r.Context(), actionID)
	if err != nil {
		respondError(w, http.StatusNotFound, "action not found")
		return
	}

	respondJSON(w, http.StatusOK, mapActionToResponse(action))
}

// ExecuteAction handles POST /api/v1/remediation/actions/{id}/execute
func (h *RemediationHandler) ExecuteAction(w http.ResponseWriter, r *http.Request) {
	actionID := chi.URLParam(r, "id")
	if actionID == "" {
		respondError(w, http.StatusBadRequest, "action id is required")
		return
	}

	if err := h.remediationService.Execute(r.Context(), actionID); err != nil {
		h.logger.ErrorWithErr(err, "Failed to execute remediation action")
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusAccepted, map[string]string{"message": "execution started"})
}

// ApproveAction handles POST /api/v1/remediation/actions/{id}/approve
func (h *RemediationHandler) ApproveAction(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	actionID := chi.URLParam(r, "id")
	if actionID == "" {
		respondError(w, http.StatusBadRequest, "action id is required")
		return
	}

	var req dto.ApproveRemediationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Approved {
		if err := h.remediationService.Approve(r.Context(), actionID, userID); err != nil {
			h.logger.ErrorWithErr(err, "Failed to approve remediation action")
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondJSON(w, http.StatusOK, map[string]string{"message": "action approved"})
	} else {
		if err := h.remediationService.Reject(r.Context(), actionID, req.Reason); err != nil {
			h.logger.ErrorWithErr(err, "Failed to reject remediation action")
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondJSON(w, http.StatusOK, map[string]string{"message": "action rejected"})
	}
}

// RollbackAction handles POST /api/v1/remediation/actions/{id}/rollback
func (h *RemediationHandler) RollbackAction(w http.ResponseWriter, r *http.Request) {
	actionID := chi.URLParam(r, "id")
	if actionID == "" {
		respondError(w, http.StatusBadRequest, "action id is required")
		return
	}

	if err := h.remediationService.Rollback(r.Context(), actionID); err != nil {
		h.logger.ErrorWithErr(err, "Failed to rollback remediation action")
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "action rolled back"})
}

// GetPendingApprovals handles GET /api/v1/remediation/pending
func (h *RemediationHandler) GetPendingApprovals(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	actions, err := h.remediationService.GetPendingApprovals(r.Context(), userID)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to get pending approvals")
		respondError(w, http.StatusInternalServerError, "failed to get pending approvals")
		return
	}

	response := make([]dto.RemediationActionResponse, 0, len(actions))
	for _, a := range actions {
		response = append(response, mapActionToResponse(a))
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"pending_approvals": response})
}

// GetSummary handles GET /api/v1/remediation/summary
func (h *RemediationHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	summary, err := h.remediationService.GetSummary(r.Context(), userID)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to get remediation summary")
		respondError(w, http.StatusInternalServerError, "failed to get summary")
		return
	}

	response := dto.RemediationSummaryResponse{
		Pending:    summary[remediation.ActionStatusPending],
		Approved:   summary[remediation.ActionStatusApproved],
		InProgress: summary[remediation.ActionStatusInProgress],
		Completed:  summary[remediation.ActionStatusCompleted],
		Failed:     summary[remediation.ActionStatusFailed],
		Rejected:   summary[remediation.ActionStatusRejected],
		RolledBack: summary[remediation.ActionStatusRolledBack],
	}

	respondJSON(w, http.StatusOK, response)
}

// Helper functions

func mapSuggestionToResponse(s *remediation.Suggestion) dto.RemediationSuggestionResponse {
	resp := dto.RemediationSuggestionResponse{
		ID:              s.ID,
		IssueType:       s.IssueType,
		IssueID:         s.IssueID,
		Title:           s.Title,
		Description:     s.Description,
		Severity:        s.Severity,
		RemediationType: string(s.RemediationType),
		Risk:            s.Risk,
		Impact:          s.Impact,
		EstimatedTime:   s.EstimatedTime,
	}

	if s.Strategy != nil {
		resp.Strategy = mapStrategyToDTO(s.Strategy)
	}

	return resp
}

func mapActionToResponse(a *remediation.Action) dto.RemediationActionResponse {
	resp := dto.RemediationActionResponse{
		ID:               a.ID,
		UserID:           a.UserID,
		DriftID:          a.DriftID,
		VulnerabilityID:  a.VulnerabilityID,
		RemediationType:  string(a.RemediationType),
		Status:           string(a.Status),
		ApprovalRequired: a.ApprovalRequired,
		ApprovedBy:       a.ApprovedBy,
		ApprovedAt:       a.ApprovedAt,
		StartedAt:        a.StartedAt,
		CompletedAt:      a.CompletedAt,
		Result:           a.Result,
		ErrorMessage:     a.ErrorMessage,
		CreatedAt:        a.CreatedAt,
		UpdatedAt:        a.UpdatedAt,
	}

	if a.Strategy != nil {
		resp.Strategy = mapStrategyToDTO(a.Strategy)
	}

	return resp
}

func mapStrategyToDTO(s *remediation.Strategy) *dto.RemediationStrategyDTO {
	d := &dto.RemediationStrategyDTO{
		Type:        string(s.Type),
		Description: s.Description,
		Parameters:  s.Parameters,
		Steps:       make([]dto.RemediationStepDTO, 0, len(s.Steps)),
	}

	for _, step := range s.Steps {
		d.Steps = append(d.Steps, dto.RemediationStepDTO{
			Order:       step.Order,
			Name:        step.Name,
			Description: step.Description,
			Status:      step.Status,
		})
	}

	return d
}

func mapDTOToStrategy(d *dto.RemediationStrategyDTO) *remediation.Strategy {
	s := &remediation.Strategy{
		Type:        remediation.RemediationType(d.Type),
		Description: d.Description,
		Parameters:  d.Parameters,
		Steps:       make([]remediation.RemediationStep, 0, len(d.Steps)),
	}

	for _, step := range d.Steps {
		s.Steps = append(s.Steps, remediation.RemediationStep{
			Order:       step.Order,
			Name:        step.Name,
			Description: step.Description,
			Status:      step.Status,
		})
	}

	return s
}
