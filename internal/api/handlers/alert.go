package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pratik-mahalle/infraudit/internal/api/dto"
	"github.com/pratik-mahalle/infraudit/internal/api/middleware"
	"github.com/pratik-mahalle/infraudit/internal/domain/alert"
	"github.com/pratik-mahalle/infraudit/internal/pkg/errors"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
	"github.com/pratik-mahalle/infraudit/internal/pkg/utils"
	"github.com/pratik-mahalle/infraudit/internal/pkg/validator"
)

type AlertHandler struct {
	service   alert.Service
	logger    *logger.Logger
	validator *validator.Validator
}

func NewAlertHandler(service alert.Service, log *logger.Logger, val *validator.Validator) *AlertHandler {
	return &AlertHandler{service: service, logger: log, validator: val}
}

// List returns all alerts with pagination and filtering
// @Summary List alerts
// @Description Get a paginated list of alerts with optional filtering
// @Tags Alerts
// @Produce json
// @Param type query string false "Filter by alert type"
// @Param severity query string false "Filter by severity"
// @Param status query string false "Filter by status"
// @Param resource query string false "Filter by resource"
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 20, max: 100)"
// @Success 200 {object} utils.PaginatedResponse{data=[]dto.AlertDTO} "List of alerts"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /alerts [get]
func (h *AlertHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filter := alert.Filter{
		Type:     r.URL.Query().Get("type"),
		Severity: r.URL.Query().Get("severity"),
		Status:   r.URL.Query().Get("status"),
		Resource: r.URL.Query().Get("resource"),
	}

	offset := (page - 1) * pageSize
	alerts, total, err := h.service.List(r.Context(), userID, filter, pageSize, offset)
	if err != nil {
		utils.WriteError(w, errors.Internal("Failed to list alerts", err))
		return
	}

	dtos := make([]dto.AlertDTO, len(alerts))
	for i, a := range alerts {
		dtos[i] = dto.AlertDTO{
			ID: a.ID, Type: a.Type, Severity: a.Severity, Title: a.Title,
			Description: a.Description, Resource: a.Resource, Status: a.Status, CreatedAt: a.CreatedAt,
		}
	}

	utils.WriteSuccess(w, http.StatusOK, utils.NewPaginatedResponse(dtos, page, pageSize, total))
}

// Get returns a single alert by ID
// @Summary Get alert by ID
// @Description Get detailed information about a specific alert
// @Tags Alerts
// @Produce json
// @Param id path int true "Alert ID"
// @Success 200 {object} dto.AlertDTO "Alert details"
// @Failure 404 {object} utils.ErrorResponse "Alert not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /alerts/{id} [get]
func (h *AlertHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	a, err := h.service.GetByID(r.Context(), userID, id)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			utils.WriteError(w, appErr)
		} else {
			utils.WriteError(w, errors.Internal("Failed to get alert", err))
		}
		return
	}

	utils.WriteSuccess(w, http.StatusOK, dto.AlertDTO{
		ID: a.ID, Type: a.Type, Severity: a.Severity, Title: a.Title,
		Description: a.Description, Resource: a.Resource, Status: a.Status, CreatedAt: a.CreatedAt,
	})
}

// Create creates a new alert
// @Summary Create alert
// @Description Create a new alert
// @Tags Alerts
// @Accept json
// @Produce json
// @Param request body dto.CreateAlertRequest true "Alert details"
// @Success 201 {object} map[string]int64 "Alert created successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid request or validation error"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /alerts [post]
func (h *AlertHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)

	var req dto.CreateAlertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, errors.BadRequest("Invalid request body"))
		return
	}

	if errs := h.validator.Validate(req); len(errs) > 0 {
		utils.WriteError(w, errors.ValidationError("Validation failed", errs))
		return
	}

	a := &alert.Alert{
		UserID: userID, Type: req.Type, Severity: req.Severity, Title: req.Title,
		Description: req.Description, Resource: req.Resource, Status: req.Status,
	}

	id, err := h.service.Create(r.Context(), a)
	if err != nil {
		utils.WriteError(w, errors.Internal("Failed to create alert", err))
		return
	}

	utils.WriteSuccess(w, http.StatusCreated, map[string]int64{"id": id})
}

// Update updates an existing alert
// @Summary Update alert
// @Description Update an existing alert
// @Tags Alerts
// @Accept json
// @Produce json
// @Param id path int true "Alert ID"
// @Param request body dto.UpdateAlertRequest true "Alert update details"
// @Success 200 {object} utils.SuccessResponse "Alert updated successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid request"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /alerts/{id} [put]
func (h *AlertHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	var req dto.UpdateAlertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, errors.BadRequest("Invalid request body"))
		return
	}

	updates := make(map[string]interface{})
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.Severity != nil {
		updates["severity"] = *req.Severity
	}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Resource != nil {
		updates["resource"] = *req.Resource
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if err := h.service.Update(r.Context(), userID, id, updates); err != nil {
		utils.WriteError(w, errors.Internal("Failed to update alert", err))
		return
	}

	utils.WriteSuccessWithMessage(w, http.StatusOK, "Alert updated successfully", nil)
}

// Delete deletes an alert
// @Summary Delete alert
// @Description Delete an alert by ID
// @Tags Alerts
// @Produce json
// @Param id path int true "Alert ID"
// @Success 200 {object} utils.SuccessResponse "Alert deleted successfully"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /alerts/{id} [delete]
func (h *AlertHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if err := h.service.Delete(r.Context(), userID, id); err != nil {
		utils.WriteError(w, errors.Internal("Failed to delete alert", err))
		return
	}

	utils.WriteSuccessWithMessage(w, http.StatusOK, "Alert deleted successfully", nil)
}

// GetSummary returns alert summary statistics
// @Summary Get alert summary
// @Description Get summary statistics of alerts
// @Tags Alerts
// @Produce json
// @Success 200 {object} map[string]interface{} "Alert summary"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /alerts/summary [get]
func (h *AlertHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)

	summary, err := h.service.GetSummary(r.Context(), userID)
	if err != nil {
		utils.WriteError(w, errors.Internal("Failed to get summary", err))
		return
	}

	utils.WriteSuccess(w, http.StatusOK, summary)
}
