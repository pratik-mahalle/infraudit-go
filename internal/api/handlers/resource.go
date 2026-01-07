package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pratik-mahalle/infraudit/internal/api/dto"
	"github.com/pratik-mahalle/infraudit/internal/api/middleware"
	"github.com/pratik-mahalle/infraudit/internal/domain/resource"
	"github.com/pratik-mahalle/infraudit/internal/pkg/errors"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
	"github.com/pratik-mahalle/infraudit/internal/pkg/utils"
	"github.com/pratik-mahalle/infraudit/internal/pkg/validator"
)

type ResourceHandler struct {
	service   resource.Service
	logger    *logger.Logger
	validator *validator.Validator
}

func NewResourceHandler(service resource.Service, log *logger.Logger, val *validator.Validator) *ResourceHandler {
	return &ResourceHandler{
		service:   service,
		logger:    log,
		validator: val,
	}
}

// List returns all resources with pagination
// @Summary List resources
// @Description Get a paginated list of cloud resources
// @Tags Resources
// @Produce json
// @Param provider query string false "Filter by provider (aws, azure, gcp)"
// @Param type query string false "Filter by resource type"
// @Param region query string false "Filter by region"
// @Param status query string false "Filter by status"
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 20, max: 100)"
// @Success 200 {object} utils.PaginatedResponse{data=[]dto.ResourceDTO} "List of resources"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /resources [get]
func (h *ResourceHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)

	// Parse query parameters
	provider := r.URL.Query().Get("provider")
	resourceType := r.URL.Query().Get("type")
	region := r.URL.Query().Get("region")
	status := r.URL.Query().Get("status")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filter := resource.Filter{
		Provider: provider,
		Type:     resourceType,
		Region:   region,
		Status:   status,
	}

	offset := (page - 1) * pageSize
	resources, total, err := h.service.List(r.Context(), userID, filter, pageSize, offset)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to list resources")
		utils.WriteError(w, errors.Internal("Failed to list resources", err))
		return
	}

	// Convert to DTOs
	dtos := make([]dto.ResourceDTO, len(resources))
	for i, res := range resources {
		dtos[i] = dto.ResourceDTO{
			Provider:   res.Provider,
			ResourceID: res.ResourceID,
			Name:       res.Name,
			Type:       res.Type,
			Region:     res.Region,
			Status:     res.Status,
		}
	}

	response := utils.NewPaginatedResponse(dtos, page, pageSize, total)
	utils.WriteSuccess(w, http.StatusOK, response)
}

// Get returns a single resource
// @Summary Get resource by ID
// @Description Get detailed information about a specific cloud resource
// @Tags Resources
// @Produce json
// @Param id path string true "Resource ID"
// @Success 200 {object} dto.ResourceDTO "Resource details"
// @Failure 404 {object} utils.ErrorResponse "Resource not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /resources/{id} [get]
func (h *ResourceHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)
	resourceID := chi.URLParam(r, "id")

	res, err := h.service.GetByID(r.Context(), userID, resourceID)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to get resource")
		if appErr, ok := err.(*errors.AppError); ok {
			utils.WriteError(w, appErr)
		} else {
			utils.WriteError(w, errors.Internal("Failed to get resource", err))
		}
		return
	}

	resourceDTO := dto.ResourceDTO{
		Provider:   res.Provider,
		ResourceID: res.ResourceID,
		Name:       res.Name,
		Type:       res.Type,
		Region:     res.Region,
		Status:     res.Status,
	}

	utils.WriteSuccess(w, http.StatusOK, resourceDTO)
}

// Create creates a new resource
// @Summary Create resource
// @Description Create a new cloud resource
// @Tags Resources
// @Accept json
// @Produce json
// @Param request body dto.CreateResourceRequest true "Resource details"
// @Success 201 {object} map[string]string "Resource created successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid request or validation error"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /resources [post]
func (h *ResourceHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)

	var req dto.CreateResourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, errors.BadRequest("Invalid request body"))
		return
	}

	if errs := h.validator.Validate(req); len(errs) > 0 {
		utils.WriteError(w, errors.ValidationError("Validation failed", errs))
		return
	}

	res := &resource.Resource{
		UserID:     userID,
		Provider:   req.Provider,
		ResourceID: req.ResourceID,
		Name:       req.Name,
		Type:       req.Type,
		Region:     req.Region,
		Status:     req.Status,
	}

	if err := h.service.Create(r.Context(), res); err != nil {
		h.logger.ErrorWithErr(err, "Failed to create resource")
		utils.WriteError(w, errors.Internal("Failed to create resource", err))
		return
	}

	utils.WriteSuccess(w, http.StatusCreated, map[string]string{
		"message":     "Resource created successfully",
		"resource_id": res.ResourceID,
	})
}

// Update updates a resource
// @Summary Update resource
// @Description Update an existing cloud resource
// @Tags Resources
// @Accept json
// @Produce json
// @Param id path string true "Resource ID"
// @Param request body dto.UpdateResourceRequest true "Resource update details"
// @Success 200 {object} utils.SuccessResponse "Resource updated successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid request"
// @Failure 404 {object} utils.ErrorResponse "Resource not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /resources/{id} [put]
func (h *ResourceHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)
	resourceID := chi.URLParam(r, "id")

	var req dto.UpdateResourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, errors.BadRequest("Invalid request body"))
		return
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.Region != nil {
		updates["region"] = *req.Region
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if err := h.service.Update(r.Context(), userID, resourceID, updates); err != nil {
		h.logger.ErrorWithErr(err, "Failed to update resource")
		if appErr, ok := err.(*errors.AppError); ok {
			utils.WriteError(w, appErr)
		} else {
			utils.WriteError(w, errors.Internal("Failed to update resource", err))
		}
		return
	}

	utils.WriteSuccessWithMessage(w, http.StatusOK, "Resource updated successfully", nil)
}

// Delete deletes a resource
// @Summary Delete resource
// @Description Delete a cloud resource by ID
// @Tags Resources
// @Produce json
// @Param id path string true "Resource ID"
// @Success 200 {object} utils.SuccessResponse "Resource deleted successfully"
// @Failure 404 {object} utils.ErrorResponse "Resource not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /resources/{id} [delete]
func (h *ResourceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)
	resourceID := chi.URLParam(r, "id")

	if err := h.service.Delete(r.Context(), userID, resourceID); err != nil {
		h.logger.ErrorWithErr(err, "Failed to delete resource")
		if appErr, ok := err.(*errors.AppError); ok {
			utils.WriteError(w, appErr)
		} else {
			utils.WriteError(w, errors.Internal("Failed to delete resource", err))
		}
		return
	}

	utils.WriteSuccessWithMessage(w, http.StatusOK, "Resource deleted successfully", nil)
}
