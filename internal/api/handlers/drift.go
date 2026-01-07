package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pratik-mahalle/infraudit/internal/api/dto"
	"github.com/pratik-mahalle/infraudit/internal/api/middleware"
	"github.com/pratik-mahalle/infraudit/internal/domain/drift"
	"github.com/pratik-mahalle/infraudit/internal/pkg/errors"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
	"github.com/pratik-mahalle/infraudit/internal/pkg/utils"
	"github.com/pratik-mahalle/infraudit/internal/pkg/validator"
)

type DriftHandler struct {
	service   drift.Service
	logger    *logger.Logger
	validator *validator.Validator
}

func NewDriftHandler(service drift.Service, log *logger.Logger, val *validator.Validator) *DriftHandler {
	return &DriftHandler{service: service, logger: log, validator: val}
}

// List returns all drifts with pagination and filtering
// @Summary List drifts
// @Description Get a paginated list of configuration drifts with optional filtering
// @Tags Drifts
// @Produce json
// @Param resource_id query string false "Filter by resource ID"
// @Param drift_type query string false "Filter by drift type"
// @Param severity query string false "Filter by severity"
// @Param status query string false "Filter by status"
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 20, max: 100)"
// @Success 200 {object} utils.PaginatedResponse{data=[]dto.DriftDTO} "List of drifts"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /drifts [get]
func (h *DriftHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filter := drift.Filter{
		ResourceID: r.URL.Query().Get("resource_id"),
		DriftType:  r.URL.Query().Get("drift_type"),
		Severity:   r.URL.Query().Get("severity"),
		Status:     r.URL.Query().Get("status"),
	}

	offset := (page - 1) * pageSize
	drifts, total, err := h.service.List(r.Context(), userID, filter, pageSize, offset)
	if err != nil {
		utils.WriteError(w, errors.Internal("Failed to list drifts", err))
		return
	}

	dtos := make([]dto.DriftDTO, len(drifts))
	for i, d := range drifts {
		dtos[i] = dto.DriftDTO{
			ID: d.ID, ResourceID: d.ResourceID, DriftType: d.DriftType, Severity: d.Severity,
			Details: d.Details, DetectedAt: d.DetectedAt, Status: d.Status,
		}
	}

	utils.WriteSuccess(w, http.StatusOK, utils.NewPaginatedResponse(dtos, page, pageSize, total))
}

// Get returns a single drift by ID
// @Summary Get drift by ID
// @Description Get detailed information about a specific configuration drift
// @Tags Drifts
// @Produce json
// @Param id path int true "Drift ID"
// @Success 200 {object} dto.DriftDTO "Drift details"
// @Failure 404 {object} utils.ErrorResponse "Drift not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /drifts/{id} [get]
func (h *DriftHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	d, err := h.service.GetByID(r.Context(), userID, id)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			utils.WriteError(w, appErr)
		} else {
			utils.WriteError(w, errors.Internal("Failed to get drift", err))
		}
		return
	}

	utils.WriteSuccess(w, http.StatusOK, dto.DriftDTO{
		ID: d.ID, ResourceID: d.ResourceID, DriftType: d.DriftType, Severity: d.Severity,
		Details: d.Details, DetectedAt: d.DetectedAt, Status: d.Status,
	})
}

// Create creates a new drift
// @Summary Create drift
// @Description Create a new configuration drift record
// @Tags Drifts
// @Accept json
// @Produce json
// @Param request body dto.CreateDriftRequest true "Drift details"
// @Success 201 {object} map[string]int64 "Drift created successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid request or validation error"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /drifts [post]
func (h *DriftHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)

	var req dto.CreateDriftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, errors.BadRequest("Invalid request body"))
		return
	}

	if errs := h.validator.Validate(req); len(errs) > 0 {
		utils.WriteError(w, errors.ValidationError("Validation failed", errs))
		return
	}

	d := &drift.Drift{
		UserID: userID, ResourceID: req.ResourceID, DriftType: req.DriftType,
		Severity: req.Severity, Details: req.Details, Status: req.Status,
	}

	id, err := h.service.Create(r.Context(), d)
	if err != nil {
		utils.WriteError(w, errors.Internal("Failed to create drift", err))
		return
	}

	utils.WriteSuccess(w, http.StatusCreated, map[string]int64{"id": id})
}

// Update updates an existing drift
// @Summary Update drift
// @Description Update an existing configuration drift
// @Tags Drifts
// @Accept json
// @Produce json
// @Param id path int true "Drift ID"
// @Param request body dto.UpdateDriftRequest true "Drift update details"
// @Success 200 {object} utils.SuccessResponse "Drift updated successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid request"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /drifts/{id} [put]
func (h *DriftHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	var req dto.UpdateDriftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, errors.BadRequest("Invalid request body"))
		return
	}

	updates := make(map[string]interface{})
	if req.ResourceID != nil {
		updates["resource_id"] = *req.ResourceID
	}
	if req.DriftType != nil {
		updates["drift_type"] = *req.DriftType
	}
	if req.Severity != nil {
		updates["severity"] = *req.Severity
	}
	if req.Details != nil {
		updates["details"] = *req.Details
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if err := h.service.Update(r.Context(), userID, id, updates); err != nil {
		utils.WriteError(w, errors.Internal("Failed to update drift", err))
		return
	}

	utils.WriteSuccessWithMessage(w, http.StatusOK, "Drift updated successfully", nil)
}

// Delete deletes a drift
// @Summary Delete drift
// @Description Delete a configuration drift by ID
// @Tags Drifts
// @Produce json
// @Param id path int true "Drift ID"
// @Success 200 {object} utils.SuccessResponse "Drift deleted successfully"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /drifts/{id} [delete]
func (h *DriftHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if err := h.service.Delete(r.Context(), userID, id); err != nil {
		utils.WriteError(w, errors.Internal("Failed to delete drift", err))
		return
	}

	utils.WriteSuccessWithMessage(w, http.StatusOK, "Drift deleted successfully", nil)
}

// GetSummary returns drift summary statistics
// @Summary Get drift summary
// @Description Get summary statistics of configuration drifts
// @Tags Drifts
// @Produce json
// @Success 200 {object} map[string]interface{} "Drift summary"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /drifts/summary [get]
func (h *DriftHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)

	summary, err := h.service.GetSummary(r.Context(), userID)
	if err != nil {
		utils.WriteError(w, errors.Internal("Failed to get summary", err))
		return
	}

	utils.WriteSuccess(w, http.StatusOK, summary)
}

// Detect triggers drift detection for all user resources
// @Summary Detect drifts
// @Description Trigger drift detection for all user resources
// @Tags Drifts
// @Produce json
// @Success 200 {object} map[string]string "Drift detection completed successfully"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /drifts/detect [post]
func (h *DriftHandler) Detect(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)

	h.logger.WithFields(map[string]interface{}{
		"user_id": userID,
	}).Info("Manual drift detection triggered")

	err := h.service.DetectDrifts(r.Context(), userID)
	if err != nil {
		utils.WriteError(w, errors.Internal("Failed to detect drifts", err))
		return
	}

	utils.WriteSuccess(w, http.StatusOK, map[string]string{
		"message": "Drift detection completed successfully",
	})
}
