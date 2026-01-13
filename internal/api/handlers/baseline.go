package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/pratik-mahalle/infraudit/internal/api/middleware"
	"github.com/pratik-mahalle/infraudit/internal/domain/baseline"
	"github.com/pratik-mahalle/infraudit/internal/pkg/errors"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
	"github.com/pratik-mahalle/infraudit/internal/pkg/utils"
)

type BaselineHandler struct {
	service baseline.Service
	logger  *logger.Logger
}

func NewBaselineHandler(service baseline.Service, log *logger.Logger) *BaselineHandler {
	return &BaselineHandler{
		service: service,
		logger:  log,
	}
}

// CreateBaseline creates a new baseline
// @Summary Create baseline
// @Description Create a new resource configuration baseline
// @Tags Baselines
// @Accept json
// @Produce json
// @Param request body object{resource_id=string,provider=string,resource_type=string,configuration=string,baseline_type=string,description=string} true "Baseline details"
// @Success 201 {object} map[string]interface{} "Baseline created successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid request"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /baselines [post]
func (h *BaselineHandler) CreateBaseline(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)

	var req struct {
		ResourceID    string `json:"resource_id"`
		Provider      string `json:"provider"`
		ResourceType  string `json:"resource_type"`
		Configuration string `json:"configuration"`
		BaselineType  string `json:"baseline_type"`
		Description   string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, errors.BadRequest("Invalid request payload"))
		return
	}

	b := &baseline.Baseline{
		UserID:        userID,
		ResourceID:    req.ResourceID,
		Provider:      req.Provider,
		ResourceType:  req.ResourceType,
		Configuration: req.Configuration,
		BaselineType:  req.BaselineType,
		Description:   req.Description,
	}

	id, err := h.service.CreateBaseline(r.Context(), b)
	if err != nil {
		utils.WriteError(w, errors.Internal("Failed to create baseline", err))
		return
	}

	utils.WriteSuccess(w, http.StatusCreated, map[string]interface{}{
		"id":      id,
		"message": "Baseline created successfully",
	})
}

// GetBaseline retrieves a baseline for a resource
// @Summary Get baseline by resource ID
// @Description Get the baseline configuration for a specific resource
// @Tags Baselines
// @Produce json
// @Param resourceId path string true "Resource ID"
// @Param type query string false "Baseline type (default: approved)"
// @Success 200 {object} baseline.Baseline "Baseline details"
// @Failure 404 {object} utils.ErrorResponse "Baseline not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /baselines/resource/{resourceId} [get]
func (h *BaselineHandler) GetBaseline(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)
	resourceID := chi.URLParam(r, "resourceId")
	baselineType := r.URL.Query().Get("type")

	if baselineType == "" {
		baselineType = baseline.TypeApproved
	}

	b, err := h.service.GetBaseline(r.Context(), userID, resourceID, baselineType)
	if err != nil {
		utils.WriteError(w, errors.NotFound("Baseline"))
		return
	}

	utils.WriteSuccess(w, http.StatusOK, b)
}

// ListBaselines lists all baselines for a user
// @Summary List baselines
// @Description Get a list of all baselines for the authenticated user
// @Tags Baselines
// @Produce json
// @Success 200 {object} map[string]interface{} "List of baselines"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /baselines [get]
func (h *BaselineHandler) ListBaselines(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)

	baselines, err := h.service.ListBaselines(r.Context(), userID)
	if err != nil {
		utils.WriteError(w, errors.Internal("Failed to list baselines", err))
		return
	}

	utils.WriteSuccess(w, http.StatusOK, map[string]interface{}{
		"baselines": baselines,
		"count":     len(baselines),
	})
}

// DeleteBaseline deletes a baseline
// @Summary Delete baseline
// @Description Delete a baseline by ID
// @Tags Baselines
// @Produce json
// @Param id path int true "Baseline ID"
// @Success 200 {object} map[string]string "Baseline deleted successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid baseline ID"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /baselines/{id} [delete]
func (h *BaselineHandler) DeleteBaseline(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.WriteError(w, errors.BadRequest("Invalid baseline ID"))
		return
	}

	if err := h.service.DeleteBaseline(r.Context(), userID, id); err != nil {
		utils.WriteError(w, errors.Internal("Failed to delete baseline", err))
		return
	}

	utils.WriteSuccess(w, http.StatusOK, map[string]string{
		"message": "Baseline deleted successfully",
	})
}
