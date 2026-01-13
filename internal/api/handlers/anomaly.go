package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pratik-mahalle/infraudit/internal/api/dto"
	"github.com/pratik-mahalle/infraudit/internal/api/middleware"
	"github.com/pratik-mahalle/infraudit/internal/domain/anomaly"
	"github.com/pratik-mahalle/infraudit/internal/pkg/errors"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
	"github.com/pratik-mahalle/infraudit/internal/pkg/utils"
	"github.com/pratik-mahalle/infraudit/internal/pkg/validator"
)

type AnomalyHandler struct {
	service   anomaly.Service
	logger    *logger.Logger
	validator *validator.Validator
}

func NewAnomalyHandler(service anomaly.Service, log *logger.Logger, val *validator.Validator) *AnomalyHandler {
	return &AnomalyHandler{service: service, logger: log, validator: val}
}

// List returns all anomalies with pagination and filtering
// @Summary List anomalies
// @Description Get a paginated list of cost anomalies with optional filtering
// @Tags Anomalies
// @Produce json
// @Param resource_id query string false "Filter by resource ID"
// @Param type query string false "Filter by anomaly type"
// @Param severity query string false "Filter by severity"
// @Param status query string false "Filter by status"
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 20, max: 100)"
// @Success 200 {object} utils.PaginatedResponse{data=[]dto.AnomalyDTO} "List of anomalies"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /anomalies [get]
func (h *AnomalyHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filter := anomaly.Filter{
		ResourceID: r.URL.Query().Get("resource_id"),
		Type:       r.URL.Query().Get("type"),
		Severity:   r.URL.Query().Get("severity"),
		Status:     r.URL.Query().Get("status"),
	}

	offset := (page - 1) * pageSize
	anomalies, total, err := h.service.List(r.Context(), userID, filter, pageSize, offset)
	if err != nil {
		utils.WriteError(w, errors.Internal("Failed to list anomalies", err))
		return
	}

	dtos := make([]dto.AnomalyDTO, len(anomalies))
	for i, a := range anomalies {
		dtos[i] = dto.AnomalyDTO{
			ID: a.ID, ResourceID: a.ResourceID, AnomalyType: a.AnomalyType, Severity: a.Severity,
			Percentage: a.Percentage, PreviousCost: a.PreviousCost, CurrentCost: a.CurrentCost, DetectedAt: a.DetectedAt, Status: a.Status,
		}
	}

	utils.WriteSuccess(w, http.StatusOK, utils.NewPaginatedResponse(dtos, page, pageSize, total))
}

// Get returns a single anomaly by ID
// @Summary Get anomaly by ID
// @Description Get detailed information about a specific cost anomaly
// @Tags Anomalies
// @Produce json
// @Param id path int true "Anomaly ID"
// @Success 200 {object} dto.AnomalyDTO "Anomaly details"
// @Failure 404 {object} utils.ErrorResponse "Anomaly not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /anomalies/{id} [get]
func (h *AnomalyHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	a, err := h.service.GetByID(r.Context(), userID, id)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			utils.WriteError(w, appErr)
		} else {
			utils.WriteError(w, errors.Internal("Failed to get anomaly", err))
		}
		return
	}

	utils.WriteSuccess(w, http.StatusOK, dto.AnomalyDTO{
		ID: a.ID, ResourceID: a.ResourceID, AnomalyType: a.AnomalyType, Severity: a.Severity,
		Percentage: a.Percentage, PreviousCost: a.PreviousCost, CurrentCost: a.CurrentCost, DetectedAt: a.DetectedAt, Status: a.Status,
	})
}

// Create creates a new anomaly
// @Summary Create anomaly
// @Description Create a new cost anomaly record
// @Tags Anomalies
// @Accept json
// @Produce json
// @Param request body dto.CreateAnomalyRequest true "Anomaly details"
// @Success 201 {object} map[string]int64 "Anomaly created successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid request or validation error"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /anomalies [post]
func (h *AnomalyHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)

	var req dto.CreateAnomalyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, errors.BadRequest("Invalid request body"))
		return
	}

	if errs := h.validator.Validate(req); len(errs) > 0 {
		utils.WriteError(w, errors.ValidationError("Validation failed", errs))
		return
	}

	a := &anomaly.Anomaly{
		UserID: userID, ResourceID: req.ResourceID, AnomalyType: req.AnomalyType, Severity: req.Severity,
		Percentage: req.Percentage, PreviousCost: req.PreviousCost, CurrentCost: req.CurrentCost, Status: req.Status,
	}

	id, err := h.service.Create(r.Context(), a)
	if err != nil {
		utils.WriteError(w, errors.Internal("Failed to create anomaly", err))
		return
	}

	utils.WriteSuccess(w, http.StatusCreated, map[string]int64{"id": id})
}

// Update updates an existing anomaly
// @Summary Update anomaly
// @Description Update an existing cost anomaly
// @Tags Anomalies
// @Accept json
// @Produce json
// @Param id path int true "Anomaly ID"
// @Param request body dto.UpdateAnomalyRequest true "Anomaly update details"
// @Success 200 {object} utils.SuccessResponse "Anomaly updated successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid request"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /anomalies/{id} [put]
func (h *AnomalyHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	var req dto.UpdateAnomalyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, errors.BadRequest("Invalid request body"))
		return
	}

	updates := make(map[string]interface{})
	if req.ResourceID != nil {
		updates["resource_id"] = *req.ResourceID
	}
	if req.AnomalyType != nil {
		updates["anomaly_type"] = *req.AnomalyType
	}
	if req.Severity != nil {
		updates["severity"] = *req.Severity
	}
	if req.Percentage != nil {
		updates["percentage"] = *req.Percentage
	}
	if req.PreviousCost != nil {
		updates["previous_cost"] = *req.PreviousCost
	}
	if req.CurrentCost != nil {
		updates["current_cost"] = *req.CurrentCost
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if err := h.service.Update(r.Context(), userID, id, updates); err != nil {
		utils.WriteError(w, errors.Internal("Failed to update anomaly", err))
		return
	}

	utils.WriteSuccessWithMessage(w, http.StatusOK, "Anomaly updated successfully", nil)
}

// Delete deletes an anomaly
// @Summary Delete anomaly
// @Description Delete a cost anomaly by ID
// @Tags Anomalies
// @Produce json
// @Param id path int true "Anomaly ID"
// @Success 200 {object} utils.SuccessResponse "Anomaly deleted successfully"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /anomalies/{id} [delete]
func (h *AnomalyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if err := h.service.Delete(r.Context(), userID, id); err != nil {
		utils.WriteError(w, errors.Internal("Failed to delete anomaly", err))
		return
	}

	utils.WriteSuccessWithMessage(w, http.StatusOK, "Anomaly deleted successfully", nil)
}

// GetSummary returns anomaly summary statistics
// @Summary Get anomaly summary
// @Description Get summary statistics of cost anomalies
// @Tags Anomalies
// @Produce json
// @Success 200 {object} map[string]interface{} "Anomaly summary"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /anomalies/summary [get]
func (h *AnomalyHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)

	summary, err := h.service.GetSummary(r.Context(), userID)
	if err != nil {
		utils.WriteError(w, errors.Internal("Failed to get summary", err))
		return
	}

	utils.WriteSuccess(w, http.StatusOK, summary)
}
