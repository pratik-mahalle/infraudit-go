package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTP metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "infraudit",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "infraudit",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path", "status"},
	)

	httpRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "infraudit",
			Subsystem: "http",
			Name:      "requests_in_flight",
			Help:      "Number of HTTP requests currently being served",
		},
	)

	// Drift detection metrics
	driftDetectionTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "infraudit",
			Subsystem: "drift",
			Name:      "detections_total",
			Help:      "Total number of drift detections",
		},
		[]string{"severity", "provider"},
	)

	driftDetectionDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "infraudit",
			Subsystem: "drift",
			Name:      "detection_duration_seconds",
			Help:      "Duration of drift detection in seconds",
			Buckets:   []float64{1, 5, 10, 30, 60, 120, 300, 600},
		},
	)

	activeDrifts = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "infraudit",
			Subsystem: "drift",
			Name:      "active_count",
			Help:      "Number of active (unresolved) drifts",
		},
		[]string{"severity"},
	)

	// Provider sync metrics
	providerSyncTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "infraudit",
			Subsystem: "provider",
			Name:      "sync_total",
			Help:      "Total number of provider syncs",
		},
		[]string{"provider", "status"},
	)

	providerSyncDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "infraudit",
			Subsystem: "provider",
			Name:      "sync_duration_seconds",
			Help:      "Duration of provider sync in seconds",
			Buckets:   []float64{1, 5, 10, 30, 60, 120, 300},
		},
		[]string{"provider"},
	)

	connectedProviders = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "infraudit",
			Subsystem: "provider",
			Name:      "connected_count",
			Help:      "Number of connected providers",
		},
		[]string{"provider"},
	)

	// Vulnerability metrics
	vulnerabilitiesTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "infraudit",
			Subsystem: "vulnerability",
			Name:      "total_count",
			Help:      "Total number of vulnerabilities",
		},
		[]string{"severity"},
	)

	// Resource metrics
	resourcesTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "infraudit",
			Subsystem: "resource",
			Name:      "total_count",
			Help:      "Total number of managed resources",
		},
		[]string{"provider", "type"},
	)

	// User metrics
	activeUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "infraudit",
			Subsystem: "user",
			Name:      "active_count",
			Help:      "Number of active users",
		},
	)

	// Database metrics
	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "infraudit",
			Subsystem: "db",
			Name:      "query_duration_seconds",
			Help:      "Database query duration in seconds",
			Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"operation", "table"},
	)
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Middleware returns a middleware that records Prometheus metrics
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		httpRequestsInFlight.Inc()
		defer httpRequestsInFlight.Dec()

		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start).Seconds()

		// Get route pattern from chi
		routePattern := chi.RouteContext(r.Context()).RoutePattern()
		if routePattern == "" {
			routePattern = "unknown"
		}

		status := strconv.Itoa(wrapped.statusCode)

		httpRequestsTotal.WithLabelValues(r.Method, routePattern, status).Inc()
		httpRequestDuration.WithLabelValues(r.Method, routePattern, status).Observe(duration)
	})
}

// Handler returns the Prometheus metrics HTTP handler
func Handler() http.Handler {
	return promhttp.Handler()
}

// RecordDriftDetection records a drift detection event
func RecordDriftDetection(severity, provider string) {
	driftDetectionTotal.WithLabelValues(severity, provider).Inc()
}

// RecordDriftDetectionDuration records the duration of a drift detection scan
func RecordDriftDetectionDuration(duration time.Duration) {
	driftDetectionDuration.Observe(duration.Seconds())
}

// SetActiveDrifts sets the gauge for active drifts by severity
func SetActiveDrifts(severity string, count float64) {
	activeDrifts.WithLabelValues(severity).Set(count)
}

// RecordProviderSync records a provider sync event
func RecordProviderSync(provider, status string, duration time.Duration) {
	providerSyncTotal.WithLabelValues(provider, status).Inc()
	providerSyncDuration.WithLabelValues(provider).Observe(duration.Seconds())
}

// SetConnectedProviders sets the gauge for connected providers
func SetConnectedProviders(provider string, count float64) {
	connectedProviders.WithLabelValues(provider).Set(count)
}

// SetVulnerabilitiesCount sets the gauge for vulnerabilities by severity
func SetVulnerabilitiesCount(severity string, count float64) {
	vulnerabilitiesTotal.WithLabelValues(severity).Set(count)
}

// SetResourcesCount sets the gauge for resources by provider and type
func SetResourcesCount(provider, resourceType string, count float64) {
	resourcesTotal.WithLabelValues(provider, resourceType).Set(count)
}

// SetActiveUsers sets the gauge for active users
func SetActiveUsers(count float64) {
	activeUsers.Set(count)
}

// RecordDBQuery records a database query duration
func RecordDBQuery(operation, table string, duration time.Duration) {
	dbQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}
