package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pratik-mahalle/infraudit/internal/api/handlers"
	"github.com/pratik-mahalle/infraudit/internal/api/router"
	"github.com/pratik-mahalle/infraudit/internal/config"
	"github.com/pratik-mahalle/infraudit/internal/integrations"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
	"github.com/pratik-mahalle/infraudit/internal/pkg/validator"
	"github.com/pratik-mahalle/infraudit/internal/repository/postgres"
	"github.com/pratik-mahalle/infraudit/internal/scanners"
	"github.com/pratik-mahalle/infraudit/internal/services"
	"github.com/pratik-mahalle/infraudit/internal/worker"
	"github.com/pratik-mahalle/infraudit/migrations"

	_ "github.com/pratik-mahalle/infraudit/docs" // Swagger docs
)

// @title InfraAudit API
// @version 1.0
// @description Cloud Infrastructure Auditing and Security Platform API
// @description
// @description This API provides endpoints for managing cloud resources, security vulnerabilities,
// @description drift detection, anomaly detection, and AI-powered recommendations across AWS, Azure, and GCP.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@infraaudit.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @tag.name Auth
// @tag.description Authentication and authorization endpoints

// @tag.name Resources
// @tag.description Cloud resource management

// @tag.name Providers
// @tag.description Cloud provider connections (AWS, Azure, GCP)

// @tag.name Alerts
// @tag.description Alert management

// @tag.name Recommendations
// @tag.description AI-powered recommendations for cost and security optimization

// @tag.name Drifts
// @tag.description Infrastructure drift detection

// @tag.name Anomalies
// @tag.description Anomaly detection and monitoring

// @tag.name Baselines
// @tag.description Resource baseline management

// @tag.name Vulnerabilities
// @tag.description Security vulnerability scanning and management

// @tag.name IaC
// @tag.description Infrastructure as Code parsing and drift detection

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.New(logger.Config{
		Level:      cfg.Logging.Level,
		Format:     cfg.Logging.Format,
		OutputPath: cfg.Logging.OutputPath,
	})

	log.WithFields(map[string]interface{}{
		"environment": cfg.Server.Environment,
		"port":        cfg.Server.Port,
	}).Info("Starting infraaudit API server")

	// Connect to database
	db, err := postgres.New(cfg.Database)
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	// Run database migrations
	if err := postgres.RunMigrations(db, migrations.GetFS()); err != nil {
		log.WithError(err).Fatal("Failed to run database migrations")
	}
	log.Info("Database migrations completed")

	log.Info("Successfully connected to database")

	// Initialize validator
	val := validator.New()

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	resourceRepo := postgres.NewResourceRepository(db)
	providerRepo := postgres.NewProviderRepository(db)
	alertRepo := postgres.NewAlertRepository(db)
	recommendationRepo := postgres.NewRecommendationRepository(db)
	driftRepo := postgres.NewDriftRepository(db)
	anomalyRepo := postgres.NewAnomalyRepository(db)
	baselineRepo := postgres.NewBaselineRepository(db)
	vulnerabilityRepo := postgres.NewVulnerabilityRepository(db)
	iacRepo := postgres.NewIaCRepository(db)

	// Initialize scanners
	trivyScanner := scanners.NewTrivyScanner(log, cfg.Scanner.TrivyPath, cfg.Scanner.TrivyCacheDir)
	nvdScanner := scanners.NewNVDScanner(log, cfg.Scanner.NVDAPIKey)

	// Initialize AI integrations
	var geminiClient *integrations.GeminiClient
	var recommendationEngine *services.RecommendationEngine

	if cfg.Provider.GeminiAPIKey != "" {
		geminiClient = integrations.NewGeminiClient(cfg.Provider.GeminiAPIKey)
		log.Info("Gemini AI client initialized")
	} else {
		log.Warn("Gemini API key not configured - recommendation generation will be disabled")
	}

	// Initialize services
	userService := services.NewUserService(userRepo, log)
	resourceService := services.NewResourceService(resourceRepo, log)
	providerService := services.NewProviderService(providerRepo, resourceRepo, log)
	alertService := services.NewAlertService(alertRepo, log)
	baselineService := services.NewBaselineService(baselineRepo, log)
	driftService := services.NewDriftService(driftRepo, baselineRepo, resourceRepo, log)
	anomalyService := services.NewAnomalyService(anomalyRepo, log)
	vulnerabilityService := services.NewVulnerabilityService(vulnerabilityRepo, log, trivyScanner, nvdScanner)
	iacService := services.NewIaCService(iacRepo, resourceService.(*services.ResourceService), driftService.(*services.DriftService))

	// Initialize recommendation engine (if Gemini is available)
	if geminiClient != nil {
		recommendationEngine = services.NewRecommendationEngine(
			geminiClient,
			resourceRepo,
			vulnerabilityRepo,
			driftRepo,
			recommendationRepo,
			log,
		)
		log.Info("Recommendation engine initialized")
	}

	// Initialize recommendation service
	recommendationService := services.NewRecommendationService(recommendationRepo, recommendationEngine, log)

	// Initialize background drift scanner worker
	driftScanInterval := 30 * time.Minute // Default: scan every 30 minutes
	driftScanner := worker.NewDriftScanner(
		driftService,
		providerService,
		userService,
		userRepo,
		driftScanInterval,
		log,
	)
	log.WithFields(map[string]interface{}{
		"interval": driftScanInterval.String(),
	}).Info("Drift scanner worker initialized")

	// Initialize handlers
	handlers := &router.Handlers{
		Health:         handlers.NewHealthHandler(db, log),
		Auth:           handlers.NewAuthHandler(userService, cfg, log, val),
		Resource:       handlers.NewResourceHandler(resourceService, log, val),
		Provider:       handlers.NewProviderHandler(providerService, log, val),
		Alert:          handlers.NewAlertHandler(alertService, log, val),
		Recommendation: handlers.NewRecommendationHandler(recommendationService, log, val),
		Drift:          handlers.NewDriftHandler(driftService, log, val),
		Anomaly:        handlers.NewAnomalyHandler(anomalyService, log, val),
		Baseline:       handlers.NewBaselineHandler(baselineService, log),
		Vulnerability:  handlers.NewVulnerabilityHandler(vulnerabilityService, log, val),
		IaC:            handlers.NewIaCHandler(iacService, log, val),
		Kubernetes:     handlers.NewKubernetesHandler(log, val),
		Billing:        handlers.NewBillingHandler(log, val),
	}

	// Setup router
	r := router.New(cfg, log, handlers)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Create context for background workers
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()

	// Start background drift scanner
	go driftScanner.Start(workerCtx)
	log.Info("Background drift scanner started")

	// Start server in goroutine
	go func() {
		log.With("address", srv.Addr).Info("Server starting")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("Server failed to start")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Server shutting down...")

	// Stop background workers
	workerCancel()
	log.Info("Background workers stopped")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.WithError(err).Error("Server forced to shutdown")
	}

	log.Info("Server exited successfully")
}
