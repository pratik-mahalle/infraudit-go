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
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
	"github.com/pratik-mahalle/infraudit/internal/pkg/validator"
	"github.com/pratik-mahalle/infraudit/internal/repository/postgres"
	"github.com/pratik-mahalle/infraudit/internal/scanners"
	"github.com/pratik-mahalle/infraudit/internal/services"
)

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

	// Initialize scanners
	trivyScanner := scanners.NewTrivyScanner(log, cfg.Scanner.TrivyPath, cfg.Scanner.TrivyCacheDir)
	nvdScanner := scanners.NewNVDScanner(log, cfg.Scanner.NVDAPIKey)

	// Initialize services
	userService := services.NewUserService(userRepo, log)
	resourceService := services.NewResourceService(resourceRepo, log)
	providerService := services.NewProviderService(providerRepo, resourceRepo, log)
	alertService := services.NewAlertService(alertRepo, log)
	recommendationService := services.NewRecommendationService(recommendationRepo, log)
	baselineService := services.NewBaselineService(baselineRepo, log)
	driftService := services.NewDriftService(driftRepo, baselineRepo, resourceRepo, log)
	anomalyService := services.NewAnomalyService(anomalyRepo, log)
	vulnerabilityService := services.NewVulnerabilityService(vulnerabilityRepo, log, trivyScanner, nvdScanner)

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

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.WithError(err).Error("Server forced to shutdown")
	}

	log.Info("Server exited successfully")
}
