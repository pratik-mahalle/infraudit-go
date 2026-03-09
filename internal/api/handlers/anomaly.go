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
		Type:     r.URL.Query().Get("type"),
		Severity: r.URL.Query().Get("severity"),
		Status:   r.URL.Query().Get("status"),
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
			ID: a.ID, AnomalyType: a.AnomalyType, ServiceName: a.ServiceName, Region: a.Region,
			Severity: a.Severity, DeviationPercentage: a.DeviationPercentage,
			ExpectedCost: a.ExpectedCost, ActualCost: a.ActualCost,
			Description: a.Description, DetectedAt: a.DetectedAt, Status: a.Status,
		}
	}

	utils.WriteSuccess(w, http.StatusOK, utils.NewPaginatedResponse(dtos, page, pageSize, total))
}

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
		ID: a.ID, AnomalyType: a.AnomalyType, ServiceName: a.ServiceName, Region: a.Region,
		Severity: a.Severity, DeviationPercentage: a.DeviationPercentage,
		ExpectedCost: a.ExpectedCost, ActualCost: a.ActualCost,
		Description: a.Description, DetectedAt: a.DetectedAt, Status: a.Status,
	})
}

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
		UserID: userID, AnomalyType: req.AnomalyType, ServiceName: req.ServiceName, Region: req.Region,
		Severity: req.Severity, DeviationPercentage: req.DeviationPercentage,
		ExpectedCost: req.ExpectedCost, ActualCost: req.ActualCost,
		Description: req.Description, Status: req.Status,
	}

	id, err := h.service.Create(r.Context(), a)
	if err != nil {
		utils.WriteError(w, errors.Internal("Failed to create anomaly", err))
		return
	}

	utils.WriteSuccess(w, http.StatusCreated, map[string]int64{"id": id})
}

func (h *AnomalyHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	var req dto.UpdateAnomalyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, errors.BadRequest("Invalid request body"))
		return
	}

	updates := make(map[string]interface{})
	if req.AnomalyType != nil {
		updates["anomaly_type"] = *req.AnomalyType
	}
	if req.Severity != nil {
		updates["severity"] = *req.Severity
	}
	if req.DeviationPercentage != nil {
		updates["deviation_percentage"] = *req.DeviationPercentage
	}
	if req.ExpectedCost != nil {
		updates["expected_cost"] = *req.ExpectedCost
	}
	if req.ActualCost != nil {
		updates["actual_cost"] = *req.ActualCost
	}
	if req.Description != nil {
		updates["description"] = *req.Description
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

func (h *AnomalyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if err := h.service.Delete(r.Context(), userID, id); err != nil {
		utils.WriteError(w, errors.Internal("Failed to delete anomaly", err))
		return
	}

	utils.WriteSuccessWithMessage(w, http.StatusOK, "Anomaly deleted successfully", nil)
}

func (h *AnomalyHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)

	summary, err := h.service.GetSummary(r.Context(), userID)
	if err != nil {
		utils.WriteError(w, errors.Internal("Failed to get summary", err))
		return
	}

	utils.WriteSuccess(w, http.StatusOK, summary)
}
