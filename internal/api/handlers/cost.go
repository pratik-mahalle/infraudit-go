package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pratik-mahalle/infraudit/internal/api/dto"
	"github.com/pratik-mahalle/infraudit/internal/domain/cost"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
)

// CostHandler handles cost-related HTTP requests
type CostHandler struct {
	costService cost.Service
	logger      *logger.Logger
}

// NewCostHandler creates a new cost handler
func NewCostHandler(costService cost.Service, log *logger.Logger) *CostHandler {
	return &CostHandler{
		costService: costService,
		logger:      log,
	}
}

// GetOverview handles GET /api/v1/costs
func (h *CostHandler) GetOverview(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	overview, err := h.costService.GetCostOverview(r.Context(), userID)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to get cost overview")
		respondError(w, http.StatusInternalServerError, "failed to get cost overview")
		return
	}

	response := dto.CostOverviewResponse{
		TotalCost:        overview.TotalCost,
		MonthlyCost:      overview.MonthlyCost,
		DailyCost:        overview.DailyCost,
		Currency:         overview.Currency,
		ByProvider:       overview.ByProvider,
		PotentialSavings: overview.PotentialSavings,
		AnomalyCount:     overview.AnomalyCount,
		TopServices:      make([]dto.ServiceCostDTO, 0, len(overview.TopServices)),
	}

	for _, svc := range overview.TopServices {
		response.TopServices = append(response.TopServices, dto.ServiceCostDTO{
			Provider:    svc.Provider,
			ServiceName: svc.ServiceName,
			Cost:        svc.Cost,
			Percentage:  svc.Percentage,
		})
	}

	if overview.Trend != nil {
		response.Trend = &dto.CostTrendDTO{
			Period:        overview.Trend.Period,
			CurrentCost:   overview.Trend.CurrentCost,
			PreviousCost:  overview.Trend.PreviousCost,
			ChangePercent: overview.Trend.ChangePercent,
			Trend:         overview.Trend.Trend,
		}
	}

	respondJSON(w, http.StatusOK, response)
}

// GetByProvider handles GET /api/v1/costs/{provider}
func (h *CostHandler) GetByProvider(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	provider := chi.URLParam(r, "provider")
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "monthly"
	}

	summary, err := h.costService.GetCostsByProvider(r.Context(), userID, provider, cost.Filter{}, period)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to get costs by provider")
		respondError(w, http.StatusInternalServerError, "failed to get costs")
		return
	}

	response := dto.CostSummaryResponse{
		Provider:  summary.Provider,
		TotalCost: summary.TotalCost,
		Currency:  summary.Currency,
		Period:    summary.Period,
		StartDate: summary.StartDate.Format("2006-01-02"),
		EndDate:   summary.EndDate.Format("2006-01-02"),
		ByService: summary.ByService,
		ByRegion:  summary.ByRegion,
	}

	respondJSON(w, http.StatusOK, response)
}

// GetTrends handles GET /api/v1/costs/trends
func (h *CostHandler) GetTrends(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	provider := r.URL.Query().Get("provider")
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "monthly"
	}

	trend, err := h.costService.GetCostTrends(r.Context(), userID, provider, period)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to get cost trends")
		respondError(w, http.StatusInternalServerError, "failed to get trends")
		return
	}

	response := dto.CostTrendDTO{
		Period:        trend.Period,
		CurrentCost:   trend.CurrentCost,
		PreviousCost:  trend.PreviousCost,
		ChangePercent: trend.ChangePercent,
		Trend:         trend.Trend,
		DataPoints:    make([]dto.CostDataPointDTO, 0, len(trend.DataPoints)),
	}

	for _, dp := range trend.DataPoints {
		response.DataPoints = append(response.DataPoints, dto.CostDataPointDTO{
			Date: dp.Date.Format("2006-01-02"),
			Cost: dp.Cost,
		})
	}

	respondJSON(w, http.StatusOK, response)
}

// GetForecast handles GET /api/v1/costs/forecast
func (h *CostHandler) GetForecast(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	provider := r.URL.Query().Get("provider")
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days <= 0 {
		days = 30
	}

	forecast, err := h.costService.GetCostForecast(r.Context(), userID, provider, days)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to get cost forecast")
		respondError(w, http.StatusInternalServerError, "failed to get forecast")
		return
	}

	response := dto.CostForecastResponse{
		Provider:        forecast.Provider,
		Period:          forecast.Period,
		ForecastedCost:  forecast.ForecastedCost,
		ConfidenceLevel: forecast.ConfidenceLevel,
		LowerBound:      forecast.LowerBound,
		UpperBound:      forecast.UpperBound,
		Currency:        forecast.Currency,
		EndDate:         forecast.EndDate.Format("2006-01-02"),
	}

	respondJSON(w, http.StatusOK, response)
}

// SyncCosts handles POST /api/v1/costs/sync
func (h *CostHandler) SyncCosts(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	provider := r.URL.Query().Get("provider")

	var err error
	if provider != "" {
		err = h.costService.SyncCosts(r.Context(), userID, provider)
	} else {
		err = h.costService.SyncAllProviders(r.Context(), userID)
	}

	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to sync costs")
		respondError(w, http.StatusInternalServerError, "failed to sync costs")
		return
	}

	respondJSON(w, http.StatusAccepted, map[string]string{"message": "cost sync initiated"})
}

// ListAnomalies handles GET /api/v1/costs/anomalies
func (h *CostHandler) ListAnomalies(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	status := r.URL.Query().Get("status")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	anomalies, total, err := h.costService.GetAnomalies(r.Context(), userID, status, limit, offset)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to list anomalies")
		respondError(w, http.StatusInternalServerError, "failed to list anomalies")
		return
	}

	response := dto.ListAnomaliesResponse{
		Anomalies: make([]dto.CostAnomalyResponse, 0, len(anomalies)),
		Total:     total,
	}

	for _, a := range anomalies {
		response.Anomalies = append(response.Anomalies, dto.CostAnomalyResponse{
			ID:           a.ID,
			Provider:     a.Provider,
			ServiceName:  a.ServiceName,
			ResourceID:   a.ResourceID,
			AnomalyType:  a.AnomalyType,
			ExpectedCost: a.ExpectedCost,
			ActualCost:   a.ActualCost,
			Deviation:    a.Deviation,
			Severity:     a.Severity,
			Status:       a.Status,
			Notes:        a.Notes,
			DetectedAt:   a.DetectedAt,
		})
	}

	respondJSON(w, http.StatusOK, response)
}

// DetectAnomalies handles POST /api/v1/costs/anomalies/detect
func (h *CostHandler) DetectAnomalies(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	provider := r.URL.Query().Get("provider")

	anomalies, err := h.costService.DetectAnomalies(r.Context(), userID, provider)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to detect anomalies")
		respondError(w, http.StatusInternalServerError, "failed to detect anomalies")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":  "anomaly detection completed",
		"detected": len(anomalies),
	})
}

// ListOptimizations handles GET /api/v1/costs/optimizations
func (h *CostHandler) ListOptimizations(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	status := r.URL.Query().Get("status")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	optimizations, total, err := h.costService.GetOptimizations(r.Context(), userID, status, limit, offset)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to list optimizations")
		respondError(w, http.StatusInternalServerError, "failed to list optimizations")
		return
	}

	response := dto.ListOptimizationsResponse{
		Optimizations: make([]dto.CostOptimizationResponse, 0, len(optimizations)),
		Total:         total,
	}

	for _, o := range optimizations {
		response.Optimizations = append(response.Optimizations, dto.CostOptimizationResponse{
			ID:               o.ID,
			Provider:         o.Provider,
			ResourceID:       o.ResourceID,
			ResourceType:     o.ResourceType,
			OptimizationType: o.OptimizationType,
			Title:            o.Title,
			Description:      o.Description,
			CurrentCost:      o.CurrentCost,
			EstimatedSavings: o.EstimatedSavings,
			SavingsPercent:   o.SavingsPercent,
			Implementation:   o.Implementation,
			Status:           o.Status,
			Details:          o.Details,
		})
	}

	respondJSON(w, http.StatusOK, response)
}

// GetSavings handles GET /api/v1/costs/savings
func (h *CostHandler) GetSavings(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	savings, err := h.costService.GetPotentialSavings(r.Context(), userID)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to get potential savings")
		respondError(w, http.StatusInternalServerError, "failed to get savings")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"potential_savings": savings,
		"currency":          "USD",
	})
}
