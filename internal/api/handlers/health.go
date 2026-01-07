package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
	"github.com/pratik-mahalle/infraudit/internal/pkg/utils"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	db     *sql.DB
	logger *logger.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *sql.DB, log *logger.Logger) *HealthHandler {
	return &HealthHandler{
		db:     db,
		logger: log,
	}
}

// Healthz handles liveness probe
// @Summary Liveness probe
// @Description Check if the application is alive
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]string "Application is alive"
// @Router /health [get]
func (h *HealthHandler) Healthz(w http.ResponseWriter, r *http.Request) {
	utils.WriteSuccess(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

// Readyz handles readiness probe
// @Summary Readiness probe
// @Description Check if the application is ready to serve requests
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]string "Application is ready"
// @Failure 503 {object} utils.ErrorResponse "Service unavailable"
// @Router /readyz [get]
func (h *HealthHandler) Readyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	// Check database connection
	if err := h.db.PingContext(ctx); err != nil {
		h.logger.ErrorWithErr(err, "Database ping failed")
		utils.WriteErrorMessage(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Database connection failed")
		return
	}

	utils.WriteSuccess(w, http.StatusOK, map[string]string{
		"status":   "ready",
		"database": "connected",
	})
}
