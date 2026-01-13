package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pratik-mahalle/infraudit/internal/api/dto"
	"github.com/pratik-mahalle/infraudit/internal/api/middleware"
	"github.com/pratik-mahalle/infraudit/internal/domain/provider"
	"github.com/pratik-mahalle/infraudit/internal/pkg/errors"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
	"github.com/pratik-mahalle/infraudit/internal/pkg/utils"
	"github.com/pratik-mahalle/infraudit/internal/pkg/validator"
)

type ProviderHandler struct {
	service   provider.Service
	logger    *logger.Logger
	validator *validator.Validator
}

func NewProviderHandler(service provider.Service, log *logger.Logger, val *validator.Validator) *ProviderHandler {
	return &ProviderHandler{
		service:   service,
		logger:    log,
		validator: val,
	}
}

// List returns all connected providers
// @Summary List connected providers
// @Description Get a list of all connected cloud providers
// @Tags Providers
// @Produce json
// @Success 200 {array} dto.ProviderDTO "List of connected providers"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /providers [get]
func (h *ProviderHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)

	providers, err := h.service.List(r.Context(), userID)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to list providers")
		utils.WriteError(w, errors.Internal("Failed to list providers", err))
		return
	}

	// Convert to DTOs (hide credentials)
	dtos := make([]dto.ProviderDTO, len(providers))
	for i, p := range providers {
		dtos[i] = dto.ProviderDTO{
			Provider:    p.Provider,
			IsConnected: p.IsConnected,
			LastSynced:  p.LastSynced,
		}
	}

	utils.WriteSuccess(w, http.StatusOK, dtos)
}

// Connect connects a cloud provider
// @Summary Connect cloud provider
// @Description Connect a cloud provider (AWS, Azure, or GCP) with credentials
// @Tags Providers
// @Accept json
// @Produce json
// @Param provider path string true "Provider type (aws, azure, gcp)"
// @Param request body dto.ConnectProviderRequest true "Provider credentials"
// @Success 200 {object} utils.SuccessResponse "Provider connected successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid request or validation error"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /providers/{provider}/connect [post]
func (h *ProviderHandler) Connect(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)
	providerType := chi.URLParam(r, "provider")

	var req dto.ConnectProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, errors.BadRequest("Invalid request body"))
		return
	}

	req.Provider = providerType // Override with URL param

	if errs := h.validator.Validate(req); len(errs) > 0 {
		utils.WriteError(w, errors.ValidationError("Validation failed", errs))
		return
	}

	// Build credentials based on provider type
	creds := provider.Credentials{}
	switch providerType {
	case "aws":
		if req.AWSAccessKeyID == nil || req.AWSSecretAccessKey == nil {
			utils.WriteError(w, errors.BadRequest("AWS credentials required"))
			return
		}
		creds.AWSAccessKeyID = *req.AWSAccessKeyID
		creds.AWSSecretAccessKey = *req.AWSSecretAccessKey
		if req.AWSRegion != nil {
			creds.AWSRegion = *req.AWSRegion
		}

	case "gcp":
		if req.GCPProjectID == nil || req.GCPServiceAccountJSON == nil {
			utils.WriteError(w, errors.BadRequest("GCP credentials required"))
			return
		}
		creds.GCPProjectID = *req.GCPProjectID
		creds.GCPServiceAccountJSON = *req.GCPServiceAccountJSON
		if req.GCPRegion != nil {
			creds.GCPRegion = *req.GCPRegion
		}

	case "azure":
		if req.AzureTenantID == nil || req.AzureClientID == nil || req.AzureClientSecret == nil || req.AzureSubscriptionID == nil {
			utils.WriteError(w, errors.BadRequest("Azure credentials required"))
			return
		}
		creds.AzureTenantID = *req.AzureTenantID
		creds.AzureClientID = *req.AzureClientID
		creds.AzureClientSecret = *req.AzureClientSecret
		creds.AzureSubscriptionID = *req.AzureSubscriptionID
		if req.AzureLocation != nil {
			creds.AzureLocation = *req.AzureLocation
		}

	default:
		utils.WriteError(w, errors.BadRequest("Unsupported provider type"))
		return
	}

	if err := h.service.Connect(r.Context(), userID, providerType, creds); err != nil {
		h.logger.ErrorWithErr(err, "Failed to connect provider")
		if appErr, ok := err.(*errors.AppError); ok {
			utils.WriteError(w, appErr)
		} else {
			utils.WriteError(w, errors.Internal("Failed to connect provider", err))
		}
		return
	}

	utils.WriteSuccessWithMessage(w, http.StatusOK, "Provider connected successfully", nil)
}

// Sync syncs resources from a provider
// @Summary Sync provider resources
// @Description Sync resources from a connected cloud provider
// @Tags Providers
// @Produce json
// @Param provider path string true "Provider type (aws, azure, gcp)"
// @Success 200 {object} utils.SuccessResponse "Provider sync initiated"
// @Failure 400 {object} utils.ErrorResponse "Invalid provider or not connected"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /providers/{provider}/sync [post]
func (h *ProviderHandler) Sync(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)
	providerType := chi.URLParam(r, "provider")

	if err := h.service.Sync(r.Context(), userID, providerType); err != nil {
		h.logger.ErrorWithErr(err, "Failed to sync provider")
		if appErr, ok := err.(*errors.AppError); ok {
			utils.WriteError(w, appErr)
		} else {
			utils.WriteError(w, errors.Internal("Failed to sync provider", err))
		}
		return
	}

	utils.WriteSuccessWithMessage(w, http.StatusOK, "Provider sync initiated", nil)
}

// Disconnect disconnects a provider
// @Summary Disconnect provider
// @Description Disconnect a cloud provider and remove stored credentials
// @Tags Providers
// @Produce json
// @Param provider path string true "Provider type (aws, azure, gcp)"
// @Success 200 {object} utils.SuccessResponse "Provider disconnected successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid provider"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /providers/{provider} [delete]
func (h *ProviderHandler) Disconnect(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)
	providerType := chi.URLParam(r, "provider")

	if err := h.service.Disconnect(r.Context(), userID, providerType); err != nil {
		h.logger.ErrorWithErr(err, "Failed to disconnect provider")
		if appErr, ok := err.(*errors.AppError); ok {
			utils.WriteError(w, appErr)
		} else {
			utils.WriteError(w, errors.Internal("Failed to disconnect provider", err))
		}
		return
	}

	utils.WriteSuccessWithMessage(w, http.StatusOK, "Provider disconnected successfully", nil)
}

// GetStatus gets the sync status for all providers
// @Summary Get provider sync status
// @Description Get the sync status and resource counts for all providers
// @Tags Providers
// @Produce json
// @Success 200 {array} dto.ProviderStatusResponse "Provider sync statuses"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /providers/status [get]
func (h *ProviderHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)

	statuses, err := h.service.GetSyncStatus(r.Context(), userID)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to get provider status")
		utils.WriteError(w, errors.Internal("Failed to get provider status", err))
		return
	}

	// Convert to DTOs
	dtos := make([]dto.ProviderStatusResponse, len(statuses))
	for i, s := range statuses {
		dtos[i] = dto.ProviderStatusResponse{
			Provider:      s.Provider,
			IsConnected:   s.IsConnected,
			LastSynced:    s.LastSynced,
			ResourceCount: s.ResourceCount,
			Status:        s.Status,
			Message:       s.Message,
		}
	}

	utils.WriteSuccess(w, http.StatusOK, dtos)
}
