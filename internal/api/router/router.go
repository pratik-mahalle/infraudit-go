package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/pratik-mahalle/infraudit/internal/api/handlers"
	"github.com/pratik-mahalle/infraudit/internal/api/middleware"
	"github.com/pratik-mahalle/infraudit/internal/config"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
	"github.com/pratik-mahalle/infraudit/internal/pkg/metrics"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Handlers struct {
	Health         *handlers.HealthHandler
	Auth           *handlers.AuthHandler
	Resource       *handlers.ResourceHandler
	Provider       *handlers.ProviderHandler
	Alert          *handlers.AlertHandler
	Recommendation *handlers.RecommendationHandler
	Drift          *handlers.DriftHandler
	Anomaly        *handlers.AnomalyHandler
	Baseline       *handlers.BaselineHandler
	Vulnerability  *handlers.VulnerabilityHandler
	IaC            *handlers.IaCHandler
	Kubernetes     *handlers.KubernetesHandler
	Billing        *handlers.BillingHandler
	WebSocket      *handlers.WebSocketHandler
	// Phase 3: Cloud Cost Analytics
	Cost *handlers.CostHandler
	// Phase 4: Compliance Framework
	Compliance *handlers.ComplianceHandler
	// Phase 5 & 6: Automation & Notifications
	Job          *handlers.JobHandler
	Remediation  *handlers.RemediationHandler
	Notification *handlers.NotificationHandler
}

func New(cfg *config.Config, log *logger.Logger, h *Handlers) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(log))
	r.Use(middleware.Recovery(log))
	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.DefaultCORS(cfg.Server.FrontendURL))
	r.Use(middleware.RateLimit(100, 200)) // 100 req/sec, burst of 200
	r.Use(metrics.Middleware)             // Prometheus metrics

	// Public routes
	r.Group(func(r chi.Router) {
		// Swagger documentation
		r.Get("/swagger/*", httpSwagger.WrapHandler)

		// Health checks
		r.Get("/health", h.Health.Healthz)
		r.Get("/healthz", h.Health.Healthz)
		r.Get("/readyz", h.Health.Readyz)

		// Prometheus metrics endpoint
		r.Handle("/metrics", metrics.Handler())

		// WebSocket endpoint
		r.Get("/ws/drifts", h.WebSocket.HandleConnection)

		// Auth endpoints (v1)
		r.Post("/api/v1/auth/register", h.Auth.Register)
		r.Post("/api/v1/auth/login", h.Auth.Login)
		r.Post("/api/v1/auth/refresh", h.Auth.RefreshToken)

		// Auth endpoints (aliases for frontend compatibility)
		r.Post("/api/auth/register", h.Auth.Register)
		r.Post("/api/auth/login", h.Auth.Login)
		r.Post("/api/auth/refresh", h.Auth.RefreshToken)
		r.Post("/api/login", h.Auth.Login)
		r.Post("/api/register", h.Auth.Register)
		r.Post("/api/logout", h.Auth.Logout)
	})

	// Protected routes (require authentication)
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg.Auth.JWTSecret))

		// Auth
		r.Get("/api/v1/auth/me", h.Auth.Me)
		r.Get("/api/auth/me", h.Auth.Me)
		r.Get("/api/user", h.Auth.Me)
		r.Post("/api/v1/auth/logout", h.Auth.Logout)

		// Resources
		r.Route("/api/v1/resources", func(r chi.Router) {
			r.Get("/", h.Resource.List)
			r.Post("/", h.Resource.Create)
			r.Get("/{id}", h.Resource.Get)
			r.Put("/{id}", h.Resource.Update)
			r.Delete("/{id}", h.Resource.Delete)
		})

		// Providers
		r.Route("/api/v1/providers", func(r chi.Router) {
			r.Get("/", h.Provider.List)
			r.Get("/status", h.Provider.GetStatus)
			r.Post("/{provider}/connect", h.Provider.Connect)
			r.Post("/{provider}/sync", h.Provider.Sync)
			r.Delete("/{provider}", h.Provider.Disconnect)
		})

		// Alerts
		r.Route("/api/v1/alerts", func(r chi.Router) {
			r.Get("/", h.Alert.List)
			r.Post("/", h.Alert.Create)
			r.Get("/summary", h.Alert.GetSummary)
			r.Get("/{id}", h.Alert.Get)
			r.Put("/{id}", h.Alert.Update)
			r.Delete("/{id}", h.Alert.Delete)
		})

		// Recommendations
		r.Route("/api/v1/recommendations", func(r chi.Router) {
			r.Get("/", h.Recommendation.List)
			r.Post("/", h.Recommendation.Create)
			r.Post("/generate", h.Recommendation.Generate)
			r.Get("/savings", h.Recommendation.GetTotalSavings)
			r.Get("/{id}", h.Recommendation.Get)
			r.Put("/{id}", h.Recommendation.Update)
			r.Delete("/{id}", h.Recommendation.Delete)
		})

		// Drifts
		r.Route("/api/v1/drifts", func(r chi.Router) {
			r.Get("/", h.Drift.List)
			r.Post("/", h.Drift.Create)
			r.Post("/detect", h.Drift.Detect)
			r.Get("/summary", h.Drift.GetSummary)
			r.Get("/{id}", h.Drift.Get)
			r.Put("/{id}", h.Drift.Update)
			r.Delete("/{id}", h.Drift.Delete)
		})

		// Anomalies
		r.Route("/api/v1/anomalies", func(r chi.Router) {
			r.Get("/", h.Anomaly.List)
			r.Post("/", h.Anomaly.Create)
			r.Get("/summary", h.Anomaly.GetSummary)
			r.Get("/{id}", h.Anomaly.Get)
			r.Put("/{id}", h.Anomaly.Update)
			r.Delete("/{id}", h.Anomaly.Delete)
		})

		// Baselines
		r.Route("/api/v1/baselines", func(r chi.Router) {
			r.Get("/", h.Baseline.ListBaselines)
			r.Post("/", h.Baseline.CreateBaseline)
			r.Get("/resource/{resourceId}", h.Baseline.GetBaseline)
			r.Delete("/{id}", h.Baseline.DeleteBaseline)
		})

		// Vulnerabilities
		r.Route("/api/v1/vulnerabilities", func(r chi.Router) {
			r.Get("/", h.Vulnerability.List)
			r.Get("/summary", h.Vulnerability.GetSummary)
			r.Get("/top", h.Vulnerability.GetTopVulnerabilities)
			r.Post("/scan", h.Vulnerability.TriggerScan)
			r.Get("/{id}", h.Vulnerability.Get)
			r.Put("/{id}/status", h.Vulnerability.UpdateStatus)
			r.Delete("/{id}", h.Vulnerability.Delete)
			r.Get("/resource/{resourceId}", h.Vulnerability.GetByResource)
			r.Get("/scans", h.Vulnerability.ListScans)
			r.Get("/scans/{id}", h.Vulnerability.GetScan)
		})

		// Infrastructure as Code
		r.Route("/api/v1/iac", func(r chi.Router) {
			r.Post("/upload", h.IaC.Upload)
			r.Get("/definitions", h.IaC.ListDefinitions)
			r.Get("/definitions/{id}", h.IaC.GetDefinition)
			r.Delete("/definitions/{id}", h.IaC.DeleteDefinition)
			r.Post("/drifts/detect", h.IaC.DetectDrift)
			r.Get("/drifts", h.IaC.ListDrifts)
			r.Get("/drifts/summary", h.IaC.GetDriftSummary)
			r.Put("/drifts/{id}/status", h.IaC.UpdateDriftStatus)
		})

		// Kubernetes
		r.Route("/api/v1/kubernetes", func(r chi.Router) {
			r.Get("/clusters", h.Kubernetes.ListClusters)
			r.Post("/clusters", h.Kubernetes.RegisterCluster)
			r.Get("/clusters/{id}", h.Kubernetes.GetCluster)
			r.Delete("/clusters/{id}", h.Kubernetes.DeleteCluster)
			r.Post("/clusters/{id}/sync", h.Kubernetes.SyncCluster)
			r.Get("/clusters/{clusterId}/namespaces", h.Kubernetes.ListNamespaces)
			r.Get("/clusters/{clusterId}/deployments", h.Kubernetes.ListDeployments)
			r.Get("/clusters/{clusterId}/pods", h.Kubernetes.ListPods)
			r.Get("/clusters/{clusterId}/services", h.Kubernetes.ListServices)
			r.Get("/stats", h.Kubernetes.GetClusterStats)
		})

		// Kubernetes (alias for frontend compatibility)
		r.Route("/api/kubernetes", func(r chi.Router) {
			r.Get("/clusters", h.Kubernetes.ListClusters)
			r.Post("/clusters", h.Kubernetes.RegisterCluster)
			r.Get("/clusters/{id}", h.Kubernetes.GetCluster)
			r.Delete("/clusters/{id}", h.Kubernetes.DeleteCluster)
			r.Post("/clusters/{id}/sync", h.Kubernetes.SyncCluster)
			r.Get("/clusters/{clusterId}/namespaces", h.Kubernetes.ListNamespaces)
			r.Get("/clusters/{clusterId}/deployments", h.Kubernetes.ListDeployments)
			r.Get("/clusters/{clusterId}/pods", h.Kubernetes.ListPods)
			r.Get("/clusters/{clusterId}/services", h.Kubernetes.ListServices)
			r.Get("/stats", h.Kubernetes.GetClusterStats)
		})

		// Billing & Subscription
		r.Route("/api/v1/billing", func(r chi.Router) {
			r.Get("/plans", h.Billing.ListPlans)
			r.Get("/info", h.Billing.GetBillingInfo)
			r.Post("/subscription", h.Billing.UpdatePlan)
			r.Post("/checkout", h.Billing.CreateCheckoutSession)
		})

		// Billing (alias for frontend compatibility)
		r.Route("/api/billing", func(r chi.Router) {
			r.Get("/plans", h.Billing.ListPlans)
			r.Get("/info", h.Billing.GetBillingInfo)
			r.Post("/subscription", h.Billing.UpdatePlan)
			r.Post("/checkout", h.Billing.CreateCheckoutSession)
		})

		// ============================================
		// Phase 3: Cloud Cost Analytics
		// ============================================

		// Cost Analytics
		r.Route("/api/v1/costs", func(r chi.Router) {
			r.Get("/", h.Cost.GetOverview)
			r.Get("/trends", h.Cost.GetTrends)
			r.Get("/forecast", h.Cost.GetForecast)
			r.Post("/sync", h.Cost.SyncCosts)
			r.Get("/savings", h.Cost.GetSavings)
			r.Get("/{provider}", h.Cost.GetByProvider)
			r.Route("/anomalies", func(r chi.Router) {
				r.Get("/", h.Cost.ListAnomalies)
				r.Post("/detect", h.Cost.DetectAnomalies)
			})
			r.Route("/optimizations", func(r chi.Router) {
				r.Get("/", h.Cost.ListOptimizations)
			})
		})

		// ============================================
		// Phase 4: Compliance Framework
		// ============================================

		// Compliance
		r.Route("/api/v1/compliance", func(r chi.Router) {
			r.Get("/overview", h.Compliance.GetOverview)
			r.Get("/trend", h.Compliance.GetTrend)
			r.Post("/assess", h.Compliance.RunAssessment)
			r.Get("/controls/failing", h.Compliance.GetFailingControls)
			r.Route("/frameworks", func(r chi.Router) {
				r.Get("/", h.Compliance.ListFrameworks)
				r.Get("/{id}", h.Compliance.GetFramework)
				r.Post("/{id}/enable", h.Compliance.EnableFramework)
				r.Post("/{id}/disable", h.Compliance.DisableFramework)
				r.Get("/{id}/controls", h.Compliance.ListControls)
			})
			r.Route("/assessments", func(r chi.Router) {
				r.Get("/", h.Compliance.ListAssessments)
				r.Get("/{id}", h.Compliance.GetAssessment)
				r.Get("/{id}/export", h.Compliance.ExportAssessment)
			})
		})

		// ============================================
		// Phase 5: Automation & Orchestration
		// ============================================

		// Scheduled Jobs
		r.Route("/api/v1/jobs", func(r chi.Router) {
			r.Get("/", h.Job.ListJobs)
			r.Post("/", h.Job.CreateJob)
			r.Get("/types", h.Job.GetJobTypes)
			r.Get("/{id}", h.Job.GetJob)
			r.Put("/{id}", h.Job.UpdateJob)
			r.Delete("/{id}", h.Job.DeleteJob)
			r.Post("/{id}/run", h.Job.TriggerJob)
			r.Get("/{id}/executions", h.Job.ListJobExecutions)
		})

		// Job Executions
		r.Route("/api/v1/executions", func(r chi.Router) {
			r.Get("/{id}", h.Job.GetJobExecution)
			r.Post("/{id}/cancel", h.Job.CancelJobExecution)
		})

		// Remediation
		r.Route("/api/v1/remediation", func(r chi.Router) {
			r.Get("/summary", h.Remediation.GetSummary)
			r.Get("/pending", h.Remediation.GetPendingApprovals)
			r.Post("/suggest/drift/{id}", h.Remediation.SuggestForDrift)
			r.Post("/suggest/vulnerability/{id}", h.Remediation.SuggestForVulnerability)
			r.Route("/actions", func(r chi.Router) {
				r.Get("/", h.Remediation.ListActions)
				r.Post("/", h.Remediation.CreateAction)
				r.Get("/{id}", h.Remediation.GetAction)
				r.Post("/{id}/execute", h.Remediation.ExecuteAction)
				r.Post("/{id}/approve", h.Remediation.ApproveAction)
				r.Post("/{id}/rollback", h.Remediation.RollbackAction)
			})
		})

		// ============================================
		// Phase 6: Notifications & Integrations
		// ============================================

		// Notifications
		r.Route("/api/v1/notifications", func(r chi.Router) {
			r.Get("/preferences", h.Notification.GetPreferences)
			r.Put("/preferences/{channel}", h.Notification.UpdatePreference)
			r.Get("/history", h.Notification.GetHistory)
			r.Post("/send", h.Notification.SendNotification)
		})

		// Webhooks
		r.Route("/api/v1/webhooks", func(r chi.Router) {
			r.Get("/", h.Notification.ListWebhooks)
			r.Post("/", h.Notification.CreateWebhook)
			r.Get("/events", h.Notification.GetAvailableEvents)
			r.Get("/{id}", h.Notification.GetWebhook)
			r.Put("/{id}", h.Notification.UpdateWebhook)
			r.Delete("/{id}", h.Notification.DeleteWebhook)
			r.Post("/{id}/test", h.Notification.TestWebhook)
		})

		// ============================================
		// Frontend Compatibility Aliases (no /v1/)
		// ============================================

		// Resources aliases
		r.Get("/api/resources", h.Resource.List)
		r.Post("/api/resources", h.Resource.Create)
		r.Get("/api/resources/{id}", h.Resource.Get)
		r.Put("/api/resources/{id}", h.Resource.Update)
		r.Delete("/api/resources/{id}", h.Resource.Delete)

		// Security drifts aliases (frontend uses /api/security-drifts)
		r.Get("/api/security-drifts", h.Drift.List)
		r.Get("/api/drifts", h.Drift.List)
		r.Post("/api/drifts", h.Drift.Create)
		r.Post("/api/drifts/detect", h.Drift.Detect)
		r.Get("/api/drifts/summary", h.Drift.GetSummary)
		r.Get("/api/drifts/{id}", h.Drift.Get)
		r.Put("/api/drifts/{id}", h.Drift.Update)
		r.Delete("/api/drifts/{id}", h.Drift.Delete)

		// Alerts aliases
		r.Get("/api/alerts", h.Alert.List)
		r.Post("/api/alerts", h.Alert.Create)
		r.Get("/api/alerts/summary", h.Alert.GetSummary)
		r.Get("/api/alerts/{id}", h.Alert.Get)
		r.Put("/api/alerts/{id}", h.Alert.Update)
		r.Delete("/api/alerts/{id}", h.Alert.Delete)

		// Recommendations aliases
		r.Get("/api/recommendations", h.Recommendation.List)
		r.Post("/api/recommendations", h.Recommendation.Create)
		r.Post("/api/recommendations/generate", h.Recommendation.Generate)
		r.Get("/api/recommendations/savings", h.Recommendation.GetTotalSavings)
		r.Get("/api/recommendations/{id}", h.Recommendation.Get)

		// Anomalies aliases
		r.Get("/api/anomalies", h.Anomaly.List)
		r.Post("/api/anomalies", h.Anomaly.Create)
		r.Get("/api/anomalies/summary", h.Anomaly.GetSummary)
		r.Get("/api/anomalies/{id}", h.Anomaly.Get)

		// Providers aliases
		r.Get("/api/providers", h.Provider.List)
		r.Get("/api/providers/status", h.Provider.GetStatus)
		r.Post("/api/providers/{provider}/connect", h.Provider.Connect)
		r.Post("/api/providers/{provider}/sync", h.Provider.Sync)
		r.Delete("/api/providers/{provider}", h.Provider.Disconnect)

		// Baselines aliases
		r.Get("/api/baselines", h.Baseline.ListBaselines)
		r.Post("/api/baselines", h.Baseline.CreateBaseline)
		r.Get("/api/baselines/resource/{resourceId}", h.Baseline.GetBaseline)
		r.Delete("/api/baselines/{id}", h.Baseline.DeleteBaseline)

		// Vulnerabilities aliases
		r.Get("/api/vulnerabilities", h.Vulnerability.List)
		r.Get("/api/vulnerabilities/summary", h.Vulnerability.GetSummary)
		r.Get("/api/vulnerabilities/top", h.Vulnerability.GetTopVulnerabilities)
		r.Post("/api/vulnerabilities/scan", h.Vulnerability.TriggerScan)
		r.Get("/api/vulnerabilities/{id}", h.Vulnerability.Get)

		// IaC aliases
		r.Post("/api/iac/upload", h.IaC.Upload)
		r.Get("/api/iac/definitions", h.IaC.ListDefinitions)
		r.Post("/api/iac/drifts/detect", h.IaC.DetectDrift)
		r.Get("/api/iac/drifts", h.IaC.ListDrifts)
		r.Get("/api/iac/drifts/summary", h.IaC.GetDriftSummary)
	})

	return r
}
