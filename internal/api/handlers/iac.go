package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pratik-mahalle/infraudit/internal/api/dto"
	"github.com/pratik-mahalle/infraudit/internal/api/middleware"
	"github.com/pratik-mahalle/infraudit/internal/domain/iac"
	"github.com/pratik-mahalle/infraudit/internal/pkg/errors"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
	"github.com/pratik-mahalle/infraudit/internal/pkg/utils"
	"github.com/pratik-mahalle/infraudit/internal/pkg/validator"
	"github.com/pratik-mahalle/infraudit/internal/services"
)

// IaCHandler handles IaC-related HTTP requests
type IaCHandler struct {
	service   *services.IaCService
	logger    *logger.Logger
	validator *validator.Validator
}

// NewIaCHandler creates a new IaC handler
func NewIaCHandler(service *services.IaCService, log *logger.Logger, val *validator.Validator) *IaCHandler {
	return &IaCHandler{
		service:   service,
		logger:    log,
		validator: val,
	}
}

// Upload uploads and parses an IaC file
// @Summary Upload IaC file
// @Description Upload and parse an Infrastructure as Code file (Terraform, CloudFormation, Kubernetes)
// @Tags IaC
// @Accept json
// @Produce json
// @Param request body dto.IaCUploadRequest true "IaC upload request"
// @Success 201 {object} utils.Response{data=dto.IaCDefinitionDTO} "IaC definition created"
// @Failure 400 {object} utils.ErrorResponse "Bad request"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /iac/upload [post]
func (h *IaCHandler) Upload(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)

	var req dto.IaCUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, errors.BadRequest("Invalid request body"))
		return
	}

	if validationErrs := h.validator.Validate(req); len(validationErrs) > 0 {
		utils.WriteError(w, errors.ValidationError("Invalid input", validationErrs))
		return
	}

	// Convert DTO to domain model
	definition, err := h.service.UploadAndParse(
		r.Context(),
		strconv.FormatInt(userID, 10),
		req.Name,
		iac.IaCType(req.IaCType),
		req.Content,
	)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to upload and parse IaC")
		utils.WriteError(w, errors.Internal("Failed to upload IaC", err))
		return
	}

	// Convert to DTO
	response := h.toDefinitionDTO(definition)
	utils.WriteSuccess(w, http.StatusCreated, response)
}

// ListDefinitions lists all IaC definitions
// @Summary List IaC definitions
// @Description Get a list of all Infrastructure as Code definitions
// @Tags IaC
// @Produce json
// @Param iac_type query string false "Filter by IaC type (terraform, cloudformation, kubernetes, helm)"
// @Success 200 {object} utils.Response{data=[]dto.IaCDefinitionDTO} "List of IaC definitions"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /iac/definitions [get]
func (h *IaCHandler) ListDefinitions(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)

	var iacType *iac.IaCType
	if typeStr := r.URL.Query().Get("iac_type"); typeStr != "" {
		t := iac.IaCType(typeStr)
		iacType = &t
	}

	definitions, err := h.service.ListDefinitions(r.Context(), strconv.FormatInt(userID, 10), iacType)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to list IaC definitions")
		utils.WriteError(w, errors.Internal("Failed to list definitions", err))
		return
	}

	// Convert to DTOs
	dtos := make([]dto.IaCDefinitionDTO, len(definitions))
	for i, def := range definitions {
		dtos[i] = h.toDefinitionDTO(def)
	}

	utils.WriteSuccess(w, http.StatusOK, dtos)
}

// GetDefinition retrieves a single IaC definition
// @Summary Get IaC definition
// @Description Get a specific Infrastructure as Code definition by ID
// @Tags IaC
// @Produce json
// @Param id path string true "Definition ID"
// @Success 200 {object} utils.Response{data=dto.IaCDefinitionDTO} "IaC definition"
// @Failure 404 {object} utils.ErrorResponse "Definition not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /iac/definitions/{id} [get]
func (h *IaCHandler) GetDefinition(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)
	definitionID := chi.URLParam(r, "id")

	definition, err := h.service.GetDefinition(r.Context(), strconv.FormatInt(userID, 10), definitionID)
	if err != nil {
		if err == iac.ErrDefinitionNotFound {
			utils.WriteError(w, errors.NotFound("IaC definition"))
			return
		}
		h.logger.ErrorWithErr(err, "Failed to get IaC definition")
		utils.WriteError(w, errors.Internal("Failed to get definition", err))
		return
	}

	response := h.toDefinitionDTO(definition)
	utils.WriteSuccess(w, http.StatusOK, response)
}

// DeleteDefinition deletes an IaC definition
// @Summary Delete IaC definition
// @Description Delete an Infrastructure as Code definition
// @Tags IaC
// @Param id path string true "Definition ID"
// @Success 204 "No content"
// @Failure 404 {object} utils.ErrorResponse "Definition not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /iac/definitions/{id} [delete]
func (h *IaCHandler) DeleteDefinition(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)
	definitionID := chi.URLParam(r, "id")

	if err := h.service.DeleteDefinition(r.Context(), strconv.FormatInt(userID, 10), definitionID); err != nil {
		if err == iac.ErrDefinitionNotFound {
			utils.WriteError(w, errors.NotFound("IaC definition"))
			return
		}
		h.logger.ErrorWithErr(err, "Failed to delete IaC definition")
		utils.WriteError(w, errors.Internal("Failed to delete definition", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DetectDrift detects configuration drift between IaC and actual resources
// @Summary Detect IaC drift
// @Description Compare IaC definition with deployed resources to detect configuration drift
// @Tags IaC
// @Accept json
// @Produce json
// @Param request body dto.IaCDriftDetectRequest true "Drift detection request"
// @Success 200 {object} utils.Response{data=[]dto.IaCDriftResultDTO} "Drift results"
// @Failure 400 {object} utils.ErrorResponse "Bad request"
// @Failure 404 {object} utils.ErrorResponse "Definition not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /iac/drifts/detect [post]
func (h *IaCHandler) DetectDrift(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)

	var req dto.IaCDriftDetectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, errors.BadRequest("Invalid request body"))
		return
	}

	if validationErrs := h.validator.Validate(req); len(validationErrs) > 0 {
		utils.WriteError(w, errors.ValidationError("Invalid input", validationErrs))
		return
	}

	drifts, err := h.service.DetectDrift(r.Context(), strconv.FormatInt(userID, 10), req.DefinitionID)
	if err != nil {
		if err == iac.ErrDefinitionNotFound {
			utils.WriteError(w, errors.NotFound("IaC definition"))
			return
		}
		h.logger.ErrorWithErr(err, "Failed to detect drift")
		utils.WriteError(w, errors.Internal("Failed to detect drift", err))
		return
	}

	// Convert to DTOs
	dtos := make([]dto.IaCDriftResultDTO, len(drifts))
	for i, drift := range drifts {
		dtos[i] = h.toDriftResultDTO(drift)
	}

	utils.WriteSuccess(w, http.StatusOK, dtos)
}

// ListDrifts lists IaC drift results
// @Summary List IaC drifts
// @Description Get a list of detected configuration drifts
// @Tags IaC
// @Produce json
// @Param definition_id query string false "Filter by definition ID"
// @Param category query string false "Filter by drift category (missing, shadow, modified, compliant)"
// @Param status query string false "Filter by status (detected, acknowledged, resolved, ignored)"
// @Success 200 {object} utils.Response{data=[]dto.IaCDriftResultDTO} "List of drift results"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /iac/drifts [get]
func (h *IaCHandler) ListDrifts(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)

	var definitionID *string
	if defID := r.URL.Query().Get("definition_id"); defID != "" {
		definitionID = &defID
	}

	var category *iac.DriftCategory
	if cat := r.URL.Query().Get("category"); cat != "" {
		c := iac.DriftCategory(cat)
		category = &c
	}

	var status *iac.DriftStatus
	if st := r.URL.Query().Get("status"); st != "" {
		s := iac.DriftStatus(st)
		status = &s
	}

	drifts, err := h.service.GetDriftResults(r.Context(), strconv.FormatInt(userID, 10), definitionID, category, status)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to list drifts")
		utils.WriteError(w, errors.Internal("Failed to list drifts", err))
		return
	}

	// Convert to DTOs
	dtos := make([]dto.IaCDriftResultDTO, len(drifts))
	for i, drift := range drifts {
		dtos[i] = h.toDriftResultDTO(drift)
	}

	utils.WriteSuccess(w, http.StatusOK, dtos)
}

// GetDriftSummary returns a summary of drift results
// @Summary Get drift summary
// @Description Get a summary of detected drifts by category and severity
// @Tags IaC
// @Produce json
// @Param definition_id query string false "Filter by definition ID"
// @Success 200 {object} utils.Response{data=dto.IaCDriftSummaryDTO} "Drift summary"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /iac/drifts/summary [get]
func (h *IaCHandler) GetDriftSummary(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)

	var definitionID *string
	if defID := r.URL.Query().Get("definition_id"); defID != "" {
		definitionID = &defID
	}

	summary, err := h.service.GetDriftSummary(r.Context(), strconv.FormatInt(userID, 10), definitionID)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to get drift summary")
		utils.WriteError(w, errors.Internal("Failed to get summary", err))
		return
	}

	// Convert to DTO
	summaryDTO := dto.IaCDriftSummaryDTO{
		Total:      summary["total"].(int),
		ByCategory: summary["by_category"].(map[string]int),
		BySeverity: summary["by_severity"].(map[string]int),
	}

	utils.WriteSuccess(w, http.StatusOK, summaryDTO)
}

// UpdateDriftStatus updates the status of a drift
// @Summary Update drift status
// @Description Update the status of a detected drift (acknowledged, resolved, ignored)
// @Tags IaC
// @Accept json
// @Produce json
// @Param id path string true "Drift ID"
// @Param request body dto.IaCDriftStatusUpdate true "Status update request"
// @Success 200 {object} utils.Response "Success"
// @Failure 400 {object} utils.ErrorResponse "Bad request"
// @Failure 404 {object} utils.ErrorResponse "Drift not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /iac/drifts/{id}/status [put]
func (h *IaCHandler) UpdateDriftStatus(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)
	driftID := chi.URLParam(r, "id")

	var req dto.IaCDriftStatusUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, errors.BadRequest("Invalid request body"))
		return
	}

	if validationErrs := h.validator.Validate(req); len(validationErrs) > 0 {
		utils.WriteError(w, errors.ValidationError("Invalid input", validationErrs))
		return
	}

	if err := h.service.UpdateDriftStatus(r.Context(), strconv.FormatInt(userID, 10), driftID, iac.DriftStatus(req.Status)); err != nil {
		if err == iac.ErrDriftNotFound {
			utils.WriteError(w, errors.NotFound("Drift result"))
			return
		}
		h.logger.ErrorWithErr(err, "Failed to update drift status")
		utils.WriteError(w, errors.Internal("Failed to update status", err))
		return
	}

	utils.WriteSuccess(w, http.StatusOK, map[string]string{"message": "Drift status updated successfully"})
}

// Helper methods to convert domain models to DTOs

func (h *IaCHandler) toDefinitionDTO(def *iac.IaCDefinition) dto.IaCDefinitionDTO {
	return dto.IaCDefinitionDTO{
		ID:              def.ID,
		UserID:          def.UserID,
		Name:            def.Name,
		IaCType:         string(def.IaCType),
		FilePath:        def.FilePath,
		ParsedResources: def.ParsedResources,
		LastParsed:      def.LastParsed,
		CreatedAt:       def.CreatedAt,
		UpdatedAt:       def.UpdatedAt,
	}
}

func (h *IaCHandler) toDriftResultDTO(drift *iac.IaCDriftResult) dto.IaCDriftResultDTO {
	var severity *string
	if drift.Severity != nil {
		s := string(*drift.Severity)
		severity = &s
	}

	return dto.IaCDriftResultDTO{
		ID:               drift.ID,
		IaCDefinitionID:  drift.IaCDefinitionID,
		IaCResourceID:    drift.IaCResourceID,
		ActualResourceID: drift.ActualResourceID,
		DriftCategory:    string(drift.DriftCategory),
		Severity:         severity,
		Details:          drift.Details,
		DetectedAt:       drift.DetectedAt,
		Status:           string(drift.Status),
		ResolvedAt:       drift.ResolvedAt,
	}
}
