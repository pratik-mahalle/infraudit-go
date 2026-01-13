package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pratik-mahalle/infraudit/internal/api/dto"
	"github.com/pratik-mahalle/infraudit/internal/api/middleware"
	"github.com/pratik-mahalle/infraudit/internal/domain/recommendation"
	"github.com/pratik-mahalle/infraudit/internal/pkg/errors"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
	"github.com/pratik-mahalle/infraudit/internal/pkg/utils"
	"github.com/pratik-mahalle/infraudit/internal/pkg/validator"
)

type RecommendationHandler struct {
	service   recommendation.Service
	logger    *logger.Logger
	validator *validator.Validator
}

func NewRecommendationHandler(service recommendation.Service, log *logger.Logger, val *validator.Validator) *RecommendationHandler {
	return &RecommendationHandler{service: service, logger: log, validator: val}
}

// List returns all recommendations with pagination and filtering
// @Summary List recommendations
// @Description Get a paginated list of recommendations with optional filtering
// @Tags Recommendations
// @Produce json
// @Param type query string false "Filter by recommendation type"
// @Param priority query string false "Filter by priority"
// @Param category query string false "Filter by category"
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 20, max: 100)"
// @Success 200 {object} utils.PaginatedResponse{data=[]dto.RecommendationDTO} "List of recommendations"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /recommendations [get]
func (h *RecommendationHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filter := recommendation.Filter{
		Type:     r.URL.Query().Get("type"),
		Priority: r.URL.Query().Get("priority"),
		Category: r.URL.Query().Get("category"),
	}

	offset := (page - 1) * pageSize
	recs, total, err := h.service.List(r.Context(), userID, filter, pageSize, offset)
	if err != nil {
		utils.WriteError(w, errors.Internal("Failed to list recommendations", err))
		return
	}

	dtos := make([]dto.RecommendationDTO, len(recs))
	for i, rec := range recs {
		dtos[i] = dto.RecommendationDTO{
			ID: rec.ID, Type: rec.Type, Priority: rec.Priority, Title: rec.Title, Description: rec.Description,
			Savings: rec.Savings, Effort: rec.Effort, Impact: rec.Impact, Category: rec.Category, Resources: rec.Resources, CreatedAt: rec.CreatedAt,
		}
	}

	utils.WriteSuccess(w, http.StatusOK, utils.NewPaginatedResponse(dtos, page, pageSize, total))
}

// Get returns a single recommendation by ID
// @Summary Get recommendation by ID
// @Description Get detailed information about a specific recommendation
// @Tags Recommendations
// @Produce json
// @Param id path int true "Recommendation ID"
// @Success 200 {object} dto.RecommendationDTO "Recommendation details"
// @Failure 404 {object} utils.ErrorResponse "Recommendation not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /recommendations/{id} [get]
func (h *RecommendationHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	rec, err := h.service.GetByID(r.Context(), userID, id)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			utils.WriteError(w, appErr)
		} else {
			utils.WriteError(w, errors.Internal("Failed to get recommendation", err))
		}
		return
	}

	utils.WriteSuccess(w, http.StatusOK, dto.RecommendationDTO{
		ID: rec.ID, Type: rec.Type, Priority: rec.Priority, Title: rec.Title, Description: rec.Description,
		Savings: rec.Savings, Effort: rec.Effort, Impact: rec.Impact, Category: rec.Category, Resources: rec.Resources, CreatedAt: rec.CreatedAt,
	})
}

// Create creates a new recommendation
// @Summary Create recommendation
// @Description Create a new recommendation
// @Tags Recommendations
// @Accept json
// @Produce json
// @Param request body dto.CreateRecommendationRequest true "Recommendation details"
// @Success 201 {object} map[string]int64 "Recommendation created successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid request or validation error"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /recommendations [post]
func (h *RecommendationHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)

	var req dto.CreateRecommendationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, errors.BadRequest("Invalid request body"))
		return
	}

	if errs := h.validator.Validate(req); len(errs) > 0 {
		utils.WriteError(w, errors.ValidationError("Validation failed", errs))
		return
	}

	rec := &recommendation.Recommendation{
		UserID: userID, Type: req.Type, Priority: req.Priority, Title: req.Title, Description: req.Description,
		Savings: req.Savings, Effort: req.Effort, Impact: req.Impact, Category: req.Category, Resources: req.Resources,
	}

	id, err := h.service.Create(r.Context(), rec)
	if err != nil {
		utils.WriteError(w, errors.Internal("Failed to create recommendation", err))
		return
	}

	utils.WriteSuccess(w, http.StatusCreated, map[string]int64{"id": id})
}

// Update updates an existing recommendation
// @Summary Update recommendation
// @Description Update an existing recommendation
// @Tags Recommendations
// @Accept json
// @Produce json
// @Param id path int true "Recommendation ID"
// @Param request body dto.UpdateRecommendationRequest true "Recommendation update details"
// @Success 200 {object} utils.SuccessResponse "Recommendation updated successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid request"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /recommendations/{id} [put]
func (h *RecommendationHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	var req dto.UpdateRecommendationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, errors.BadRequest("Invalid request body"))
		return
	}

	updates := make(map[string]interface{})
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Savings != nil {
		updates["savings"] = *req.Savings
	}
	if req.Effort != nil {
		updates["effort"] = *req.Effort
	}
	if req.Impact != nil {
		updates["impact"] = *req.Impact
	}
	if req.Category != nil {
		updates["category"] = *req.Category
	}
	if req.Resources != nil {
		updates["resources"] = *req.Resources
	}

	if err := h.service.Update(r.Context(), userID, id, updates); err != nil {
		utils.WriteError(w, errors.Internal("Failed to update recommendation", err))
		return
	}

	utils.WriteSuccessWithMessage(w, http.StatusOK, "Recommendation updated successfully", nil)
}

// Delete deletes a recommendation
// @Summary Delete recommendation
// @Description Delete a recommendation by ID
// @Tags Recommendations
// @Produce json
// @Param id path int true "Recommendation ID"
// @Success 200 {object} utils.SuccessResponse "Recommendation deleted successfully"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /recommendations/{id} [delete]
func (h *RecommendationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if err := h.service.Delete(r.Context(), userID, id); err != nil {
		utils.WriteError(w, errors.Internal("Failed to delete recommendation", err))
		return
	}

	utils.WriteSuccessWithMessage(w, http.StatusOK, "Recommendation deleted successfully", nil)
}

// GetTotalSavings returns total potential savings
// @Summary Get total savings
// @Description Get total potential cost savings from all recommendations
// @Tags Recommendations
// @Produce json
// @Success 200 {object} map[string]float64 "Total savings amount"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /recommendations/savings [get]
func (h *RecommendationHandler) GetTotalSavings(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)

	savings, err := h.service.GetTotalSavings(r.Context(), userID)
	if err != nil {
		utils.WriteError(w, errors.Internal("Failed to get total savings", err))
		return
	}

	utils.WriteSuccess(w, http.StatusOK, map[string]float64{"total_savings": savings})
}

// Generate triggers recommendation generation
// @Summary Generate recommendations
// @Description Trigger the generation of new recommendations based on current resources
// @Tags Recommendations
// @Produce json
// @Success 200 {object} utils.SuccessResponse "Recommendations generated successfully"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /recommendations/generate [post]
func (h *RecommendationHandler) Generate(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)

	h.logger.WithFields(map[string]interface{}{
		"user_id": userID,
	}).Info("Triggering recommendation generation")

	// Generate recommendations asynchronously would be better for production
	// but for now, we'll do it synchronously
	if err := h.service.GenerateRecommendations(r.Context(), userID); err != nil {
		utils.WriteError(w, errors.Internal("Failed to generate recommendations", err))
		return
	}

	utils.WriteSuccessWithMessage(w, http.StatusOK, "Recommendations generated successfully", nil)
}
