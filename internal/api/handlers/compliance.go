package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pratik-mahalle/infraudit/internal/api/dto"
	"github.com/pratik-mahalle/infraudit/internal/domain/compliance"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
)

// ComplianceHandler handles compliance-related HTTP requests
type ComplianceHandler struct {
	complianceService compliance.Service
	logger            *logger.Logger
}

// NewComplianceHandler creates a new compliance handler
func NewComplianceHandler(complianceService compliance.Service, log *logger.Logger) *ComplianceHandler {
	return &ComplianceHandler{
		complianceService: complianceService,
		logger:            log,
	}
}

// ListFrameworks handles GET /api/v1/compliance/frameworks
func (h *ComplianceHandler) ListFrameworks(w http.ResponseWriter, r *http.Request) {
	frameworks, err := h.complianceService.ListFrameworks(r.Context())
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to list frameworks")
		respondError(w, http.StatusInternalServerError, "failed to list frameworks")
		return
	}

	response := dto.ListFrameworksResponse{
		Frameworks: make([]dto.ComplianceFrameworkResponse, 0, len(frameworks)),
	}

	for _, f := range frameworks {
		response.Frameworks = append(response.Frameworks, dto.ComplianceFrameworkResponse{
			ID:          f.ID,
			Name:        f.Name,
			Version:     f.Version,
			Description: f.Description,
			Provider:    f.Provider,
			IsEnabled:   f.IsEnabled,
		})
	}

	respondJSON(w, http.StatusOK, response)
}

// GetFramework handles GET /api/v1/compliance/frameworks/{id}
func (h *ComplianceHandler) GetFramework(w http.ResponseWriter, r *http.Request) {
	frameworkID := chi.URLParam(r, "id")
	if frameworkID == "" {
		respondError(w, http.StatusBadRequest, "framework id is required")
		return
	}

	framework, err := h.complianceService.GetFramework(r.Context(), frameworkID)
	if err != nil {
		respondError(w, http.StatusNotFound, "framework not found")
		return
	}

	respondJSON(w, http.StatusOK, dto.ComplianceFrameworkResponse{
		ID:          framework.ID,
		Name:        framework.Name,
		Version:     framework.Version,
		Description: framework.Description,
		Provider:    framework.Provider,
		IsEnabled:   framework.IsEnabled,
	})
}

// EnableFramework handles POST /api/v1/compliance/frameworks/{id}/enable
func (h *ComplianceHandler) EnableFramework(w http.ResponseWriter, r *http.Request) {
	frameworkID := chi.URLParam(r, "id")
	if frameworkID == "" {
		respondError(w, http.StatusBadRequest, "framework id is required")
		return
	}

	if err := h.complianceService.EnableFramework(r.Context(), frameworkID); err != nil {
		h.logger.ErrorWithErr(err, "Failed to enable framework")
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "framework enabled"})
}

// DisableFramework handles POST /api/v1/compliance/frameworks/{id}/disable
func (h *ComplianceHandler) DisableFramework(w http.ResponseWriter, r *http.Request) {
	frameworkID := chi.URLParam(r, "id")
	if frameworkID == "" {
		respondError(w, http.StatusBadRequest, "framework id is required")
		return
	}

	if err := h.complianceService.DisableFramework(r.Context(), frameworkID); err != nil {
		h.logger.ErrorWithErr(err, "Failed to disable framework")
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "framework disabled"})
}

// ListControls handles GET /api/v1/compliance/frameworks/{id}/controls
func (h *ComplianceHandler) ListControls(w http.ResponseWriter, r *http.Request) {
	frameworkID := chi.URLParam(r, "id")
	if frameworkID == "" {
		respondError(w, http.StatusBadRequest, "framework id is required")
		return
	}

	category := r.URL.Query().Get("category")

	controls, err := h.complianceService.GetControls(r.Context(), frameworkID, category)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to list controls")
		respondError(w, http.StatusInternalServerError, "failed to list controls")
		return
	}

	response := dto.ListControlsResponse{
		Controls: make([]dto.ComplianceControlResponse, 0, len(controls)),
		Total:    len(controls),
	}

	for _, c := range controls {
		response.Controls = append(response.Controls, dto.ComplianceControlResponse{
			ID:           c.ID,
			ControlID:    c.ControlID,
			Title:        c.Title,
			Description:  c.Description,
			Category:     c.Category,
			Severity:     c.Severity,
			Remediation:  c.Remediation,
			ReferenceURL: c.ReferenceURL,
		})
	}

	respondJSON(w, http.StatusOK, response)
}

// RunAssessment handles POST /api/v1/compliance/assess
func (h *ComplianceHandler) RunAssessment(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.RunAssessmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.FrameworkID == "" {
		respondError(w, http.StatusBadRequest, "framework_id is required")
		return
	}

	assessment, err := h.complianceService.RunAssessment(r.Context(), userID, req.FrameworkID)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to run assessment")
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusAccepted, dto.ComplianceAssessmentResponse{
		ID:                    assessment.ID,
		FrameworkID:           assessment.FrameworkID,
		FrameworkName:         assessment.FrameworkName,
		AssessmentDate:        assessment.AssessmentDate,
		TotalControls:         assessment.TotalControls,
		PassedControls:        assessment.PassedControls,
		FailedControls:        assessment.FailedControls,
		NotApplicableControls: assessment.NotApplicableControls,
		CompliancePercent:     assessment.CompliancePercent,
		Status:                assessment.Status,
	})
}

// ListAssessments handles GET /api/v1/compliance/assessments
func (h *ComplianceHandler) ListAssessments(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	frameworkID := r.URL.Query().Get("framework_id")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	assessments, total, err := h.complianceService.ListAssessments(r.Context(), userID, frameworkID, limit, offset)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to list assessments")
		respondError(w, http.StatusInternalServerError, "failed to list assessments")
		return
	}

	response := dto.ListAssessmentsResponse{
		Assessments: make([]dto.ComplianceAssessmentResponse, 0, len(assessments)),
		Total:       total,
	}

	for _, a := range assessments {
		response.Assessments = append(response.Assessments, dto.ComplianceAssessmentResponse{
			ID:                    a.ID,
			FrameworkID:           a.FrameworkID,
			FrameworkName:         a.FrameworkName,
			AssessmentDate:        a.AssessmentDate,
			TotalControls:         a.TotalControls,
			PassedControls:        a.PassedControls,
			FailedControls:        a.FailedControls,
			NotApplicableControls: a.NotApplicableControls,
			CompliancePercent:     a.CompliancePercent,
			Status:                a.Status,
		})
	}

	respondJSON(w, http.StatusOK, response)
}

// GetAssessment handles GET /api/v1/compliance/assessments/{id}
func (h *ComplianceHandler) GetAssessment(w http.ResponseWriter, r *http.Request) {
	assessmentID := chi.URLParam(r, "id")
	if assessmentID == "" {
		respondError(w, http.StatusBadRequest, "assessment id is required")
		return
	}

	assessment, err := h.complianceService.GetAssessment(r.Context(), assessmentID)
	if err != nil {
		respondError(w, http.StatusNotFound, "assessment not found")
		return
	}

	var findings []dto.AssessmentFindingDTO
	if len(assessment.Findings) > 0 {
		var rawFindings []compliance.AssessmentFinding
		json.Unmarshal(assessment.Findings, &rawFindings)
		for _, f := range rawFindings {
			findings = append(findings, dto.AssessmentFindingDTO{
				ControlID:         f.ControlID,
				ControlTitle:      f.ControlTitle,
				Category:          f.Category,
				Severity:          f.Severity,
				Status:            f.Status,
				AffectedCount:     f.AffectedCount,
				AffectedResources: f.AffectedResources,
				Remediation:       f.Remediation,
			})
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"assessment": dto.ComplianceAssessmentResponse{
			ID:                    assessment.ID,
			FrameworkID:           assessment.FrameworkID,
			FrameworkName:         assessment.FrameworkName,
			AssessmentDate:        assessment.AssessmentDate,
			TotalControls:         assessment.TotalControls,
			PassedControls:        assessment.PassedControls,
			FailedControls:        assessment.FailedControls,
			NotApplicableControls: assessment.NotApplicableControls,
			CompliancePercent:     assessment.CompliancePercent,
			Status:                assessment.Status,
		},
		"findings": findings,
	})
}

// GetOverview handles GET /api/v1/compliance/overview
func (h *ComplianceHandler) GetOverview(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	overview, err := h.complianceService.GetComplianceOverview(r.Context(), userID)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to get compliance overview")
		respondError(w, http.StatusInternalServerError, "failed to get overview")
		return
	}

	response := dto.ComplianceOverviewResponse{
		TotalControls:     overview.TotalControls,
		PassedControls:    overview.PassedControls,
		FailedControls:    overview.FailedControls,
		CompliancePercent: overview.CompliancePercent,
		ByFramework:       make([]dto.FrameworkComplianceDTO, 0, len(overview.ByFramework)),
		BySeverity:        overview.BySeverity,
	}

	for _, fc := range overview.ByFramework {
		response.ByFramework = append(response.ByFramework, dto.FrameworkComplianceDTO{
			FrameworkID:       fc.FrameworkID,
			FrameworkName:     fc.FrameworkName,
			TotalControls:     fc.TotalControls,
			PassedControls:    fc.PassedControls,
			FailedControls:    fc.FailedControls,
			CompliancePercent: fc.CompliancePercent,
			LastAssessment:    fc.LastAssessment,
		})
	}

	respondJSON(w, http.StatusOK, response)
}

// GetTrend handles GET /api/v1/compliance/trend
func (h *ComplianceHandler) GetTrend(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	frameworkID := r.URL.Query().Get("framework_id")
	if frameworkID == "" {
		respondError(w, http.StatusBadRequest, "framework_id is required")
		return
	}

	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days <= 0 {
		days = 30
	}

	trend, err := h.complianceService.GetComplianceTrend(r.Context(), userID, frameworkID, days)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to get compliance trend")
		respondError(w, http.StatusInternalServerError, "failed to get trend")
		return
	}

	response := dto.ComplianceTrendResponse{
		FrameworkID:   trend.FrameworkID,
		FrameworkName: trend.FrameworkName,
		CurrentScore:  trend.CurrentScore,
		PreviousScore: trend.PreviousScore,
		ChangePercent: trend.ChangePercent,
		Trend:         trend.Trend,
		DataPoints:    make([]dto.TrendDataPointDTO, 0, len(trend.DataPoints)),
	}

	for _, dp := range trend.DataPoints {
		response.DataPoints = append(response.DataPoints, dto.TrendDataPointDTO{
			Date:            dp.Date,
			ComplianceScore: dp.ComplianceScore,
			PassedControls:  dp.PassedControls,
			TotalControls:   dp.TotalControls,
		})
	}

	respondJSON(w, http.StatusOK, response)
}

// GetFailingControls handles GET /api/v1/compliance/controls/failing
func (h *ComplianceHandler) GetFailingControls(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	frameworkID := r.URL.Query().Get("framework_id")
	if frameworkID == "" {
		respondError(w, http.StatusBadRequest, "framework_id is required")
		return
	}

	findings, err := h.complianceService.GetFailingControls(r.Context(), userID, frameworkID)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to get failing controls")
		respondError(w, http.StatusInternalServerError, "failed to get failing controls")
		return
	}

	response := make([]dto.AssessmentFindingDTO, 0, len(findings))
	for _, f := range findings {
		response = append(response, dto.AssessmentFindingDTO{
			ControlID:         f.ControlID,
			ControlTitle:      f.ControlTitle,
			Category:          f.Category,
			Severity:          f.Severity,
			Status:            f.Status,
			AffectedCount:     f.AffectedCount,
			AffectedResources: f.AffectedResources,
			Remediation:       f.Remediation,
		})
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"failing_controls": response})
}

// ExportAssessment handles GET /api/v1/compliance/assessments/{id}/export
func (h *ComplianceHandler) ExportAssessment(w http.ResponseWriter, r *http.Request) {
	assessmentID := chi.URLParam(r, "id")
	if assessmentID == "" {
		respondError(w, http.StatusBadRequest, "assessment id is required")
		return
	}

	export, err := h.complianceService.ExportAssessment(r.Context(), assessmentID)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to export assessment")
		respondError(w, http.StatusNotFound, "assessment not found")
		return
	}

	respondJSON(w, http.StatusOK, export)
}
