package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/pratik-mahalle/infraudit/internal/api/handlers"
	"github.com/pratik-mahalle/infraudit/internal/api/middleware"
	"github.com/pratik-mahalle/infraudit/internal/config"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
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
}

func New(cfg *config.Config, log *logger.Logger, h *Handlers) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(log))
	r.Use(middleware.Recovery(log))
	r.Use(middleware.DefaultCORS(cfg.Server.FrontendURL))
	r.Use(middleware.RateLimit(100, 200)) // 100 req/sec, burst of 200

	// Public routes
	r.Group(func(r chi.Router) {
		// Swagger documentation
		r.Get("/swagger/*", httpSwagger.WrapHandler)

		// Health checks
		r.Get("/health", h.Health.Healthz)
		r.Get("/healthz", h.Health.Healthz)
		r.Get("/readyz", h.Health.Readyz)

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
	})

	// Protected routes (require authentication)
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg.Auth.JWTSecret))

		// Auth
		r.Get("/api/v1/auth/me", h.Auth.Me)
		r.Get("/api/auth/me", h.Auth.Me)

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
	})

	return r
}
