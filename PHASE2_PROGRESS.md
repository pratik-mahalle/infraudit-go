# Phase 2 Implementation Progress

**Status**: 40% Complete - All Repositories Done, Ready for Services & Handlers

---

## ✅ What's Been Completed

### 1. All Repository Implementations (100%)

All 7 domain repositories have been created in `internal/repository/postgres/`:

- ✅ **db.go** - Database connection helper (SQLite & PostgreSQL)
- ✅ **user.go** - User CRUD operations
- ✅ **resource.go** - Cloud resource management with filters & batch operations
- ✅ **provider.go** - Cloud provider account management
- ✅ **alert.go** - Alert management with filtering & pagination
- ✅ **recommendation.go** - Recommendations with savings calculation
- ✅ **drift.go** - Security drift tracking
- ✅ **anomaly.go** - Cost anomaly detection

**All repositories**:
- Implement their domain interfaces
- Include proper error handling
- Support pagination where needed
- Have filtering capabilities
- Use context for cancellation

---

## 🚧 What Remains

### 2. Service Implementations (0%)

**Location**: `internal/services/`

You need to create service implementations that contain business logic. Here's a template:

```go
// internal/services/user_service.go
package services

import (
	"context"
	"infraaudit/backend/internal/domain/user"
	"infraaudit/backend/internal/pkg/logger"
)

type UserService struct {
	repo   user.Repository
	logger *logger.Logger
}

func NewUserService(repo user.Repository, log *logger.Logger) user.Service {
	return &UserService{
		repo:   repo,
		logger: log,
	}
}

func (s *UserService) GetByID(ctx context.Context, id int64) (*user.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	return s.repo.GetByEmail(ctx, email)
}

func (s *UserService) Create(ctx context.Context, email string) (*user.User, error) {
	// Business logic here
	u := &user.User{
		Email:    email,
		Role:     user.RoleUser,
		PlanType: user.PlanTypeFree,
	}

	err := s.repo.Create(ctx, u)
	if err != nil {
		s.logger.ErrorWithErr(err, "Failed to create user")
		return nil, err
	}

	s.logger.WithFields(map[string]interface{}{
		"user_id": u.ID,
		"email":   u.Email,
	}).Info("User created")

	return u, nil
}

func (s *UserService) Update(ctx context.Context, u *user.User) error {
	return s.repo.Update(ctx, u)
}

func (s *UserService) GetTrialStatus(ctx context.Context, userID int64) (*user.TrialStatus, error) {
	// Implement trial logic
	return &user.TrialStatus{
		Status:        "active",
		DaysRemaining: 14,
	}, nil
}

func (s *UserService) UpgradePlan(ctx context.Context, userID int64, planType string) error {
	u, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	u.PlanType = planType
	return s.repo.Update(ctx, u)
}
```

**Services to create**:
1. `user_service.go` - User management
2. `resource_service.go` - Resource operations
3. `provider_service.go` - Provider connections & sync
4. `alert_service.go` - Alert management
5. `recommendation_service.go` - Recommendation generation
6. `drift_service.go` - Drift detection
7. `anomaly_service.go` - Anomaly detection

### 3. Complete Handlers (20% - Only auth.go exists)

**Location**: `internal/api/handlers/`

You have `auth.go` as an example. Create similar handlers for:

```go
// internal/api/handlers/resource.go
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"infraaudit/backend/internal/api/dto"
	"infraaudit/backend/internal/api/middleware"
	"infraaudit/backend/internal/domain/resource"
	"infraaudit/backend/internal/pkg/errors"
	"infraaudit/backend/internal/pkg/logger"
	"infraaudit/backend/internal/pkg/utils"
	"infraaudit/backend/internal/pkg/validator"
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
			ID:         res.ID,
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
func (h *ResourceHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)
	resourceID := chi.URLParam(r, "id")

	res, err := h.service.GetByID(r.Context(), userID, resourceID)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to get resource")
		utils.WriteError(w, err.(*errors.AppError))
		return
	}

	dto := dto.ResourceDTO{
		ID:         res.ID,
		Provider:   res.Provider,
		ResourceID: res.ResourceID,
		Name:       res.Name,
		Type:       res.Type,
		Region:     res.Region,
		Status:     res.Status,
	}

	utils.WriteSuccess(w, http.StatusOK, dto)
}

// Create creates a new resource
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

	utils.WriteSuccess(w, http.StatusCreated, res)
}

// Update updates a resource
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
		utils.WriteError(w, err.(*errors.AppError))
		return
	}

	utils.WriteSuccessWithMessage(w, http.StatusOK, "Resource updated successfully", nil)
}

// Delete deletes a resource
func (h *ResourceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)
	resourceID := chi.URLParam(r, "id")

	if err := h.service.Delete(r.Context(), userID, resourceID); err != nil {
		h.logger.ErrorWithErr(err, "Failed to delete resource")
		utils.WriteError(w, err.(*errors.AppError))
		return
	}

	utils.WriteSuccessWithMessage(w, http.StatusOK, "Resource deleted successfully", nil)
}
```

**Handlers to create**:
1. `health.go` - Health check endpoints
2. `resource.go` - Resource CRUD
3. `provider.go` - Provider management
4. `alert.go` - Alert management
5. `recommendation.go` - Recommendations
6. `drift.go` - Drift management
7. `anomaly.go` - Anomaly management

### 4. Router Setup (0%)

**File**: `internal/api/router/router.go`

```go
package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"infraaudit/backend/internal/api/handlers"
	"infraaudit/backend/internal/api/middleware"
	"infraaudit/backend/internal/config"
	"infraaudit/backend/internal/pkg/logger"
)

// Setup creates and configures the HTTP router
func Setup(
	authHandler *handlers.AuthHandler,
	resourceHandler *handlers.ResourceHandler,
	providerHandler *handlers.ProviderHandler,
	alertHandler *handlers.AlertHandler,
	recommendationHandler *handlers.RecommendationHandler,
	driftHandler *handlers.DriftHandler,
	anomalyHandler *handlers.AnomalyHandler,
	healthHandler *handlers.HealthHandler,
	cfg *config.Config,
	log *logger.Logger,
) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(log))
	r.Use(middleware.Recovery(log))
	r.Use(middleware.DefaultCORS(cfg.Server.FrontendURL))
	r.Use(middleware.RateLimit(10, 20))

	// Public routes
	r.Group(func(r chi.Router) {
		r.Get("/healthz", healthHandler.Healthz)
		r.Get("/readyz", healthHandler.Readyz)

		r.Post("/api/auth/login", authHandler.Login)
		r.Post("/api/auth/register", authHandler.Register)
		r.Post("/api/auth/refresh", authHandler.RefreshToken)
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg.Auth.JWTSecret))

		// Auth routes
		r.Get("/api/auth/me", authHandler.Me)
		r.Post("/api/auth/logout", authHandler.Logout)

		// Resource routes
		r.Route("/api/resources", func(r chi.Router) {
			r.Get("/", resourceHandler.List)
			r.Post("/", resourceHandler.Create)
			r.Get("/{id}", resourceHandler.Get)
			r.Patch("/{id}", resourceHandler.Update)
			r.Delete("/{id}", resourceHandler.Delete)
		})

		// Provider routes
		r.Route("/api/providers", func(r chi.Router) {
			r.Get("/", providerHandler.List)
			r.Post("/{provider}/connect", providerHandler.Connect)
			r.Post("/{provider}/sync", providerHandler.Sync)
			r.Delete("/{provider}", providerHandler.Disconnect)
			r.Get("/status", providerHandler.GetStatus)
		})

		// Alert routes
		r.Route("/api/alerts", func(r chi.Router) {
			r.Get("/", alertHandler.List)
			r.Post("/", alertHandler.Create)
			r.Get("/{id}", alertHandler.Get)
			r.Put("/{id}", alertHandler.Update)
			r.Delete("/{id}", alertHandler.Delete)
			r.Get("/summary", alertHandler.GetSummary)
		})

		// Recommendation routes
		r.Route("/api/recommendations", func(r chi.Router) {
			r.Get("/", recommendationHandler.List)
			r.Post("/", recommendationHandler.Create)
			r.Get("/{id}", recommendationHandler.Get)
			r.Put("/{id}", recommendationHandler.Update)
			r.Delete("/{id}", recommendationHandler.Delete)
			r.Get("/savings", recommendationHandler.GetTotalSavings)
		})

		// Drift routes
		r.Route("/api/drifts", func(r chi.Router) {
			r.Get("/", driftHandler.List)
			r.Post("/", driftHandler.Create)
			r.Get("/{id}", driftHandler.Get)
			r.Put("/{id}", driftHandler.Update)
			r.Delete("/{id}", driftHandler.Delete)
			r.Get("/summary", driftHandler.GetSummary)
		})

		// Anomaly routes
		r.Route("/api/anomalies", func(r chi.Router) {
			r.Get("/", anomalyHandler.List)
			r.Post("/", anomalyHandler.Create)
			r.Get("/{id}", anomalyHandler.Get)
			r.Put("/{id}", anomalyHandler.Update)
			r.Delete("/{id}", anomalyHandler.Delete)
			r.Get("/summary", anomalyHandler.GetSummary)
		})
	})

	return r
}
```

### 5. New main.go with Dependency Injection (0%)

**File**: `cmd/api/main.go`

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"infraaudit/backend/internal/api/handlers"
	"infraaudit/backend/internal/api/router"
	"infraaudit/backend/internal/config"
	"infraaudit/backend/internal/pkg/logger"
	"infraaudit/backend/internal/pkg/validator"
	"infraaudit/backend/internal/repository/postgres"
	"infraaudit/backend/internal/services"
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
		Level:  cfg.Logging.Level,
		Format: cfg.Logging.Format,
	})

	log.Info("Starting InfraAudit API server...")

	// Initialize validator
	val := validator.New()

	// Setup database
	db, err := postgres.New(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Info("Database connected successfully")

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	resourceRepo := postgres.NewResourceRepository(db)
	providerRepo := postgres.NewProviderRepository(db)
	alertRepo := postgres.NewAlertRepository(db)
	recommendationRepo := postgres.NewRecommendationRepository(db)
	driftRepo := postgres.NewDriftRepository(db)
	anomalyRepo := postgres.NewAnomalyRepository(db)

	// Initialize services
	userService := services.NewUserService(userRepo, log)
	resourceService := services.NewResourceService(resourceRepo, log)
	providerService := services.NewProviderService(providerRepo, resourceRepo, log)
	alertService := services.NewAlertService(alertRepo, log)
	recommendationService := services.NewRecommendationService(recommendationRepo, log)
	driftService := services.NewDriftService(driftRepo, log)
	anomalyService := services.NewAnomalyService(anomalyRepo, log)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userService, cfg, log, val)
	resourceHandler := handlers.NewResourceHandler(resourceService, log, val)
	providerHandler := handlers.NewProviderHandler(providerService, log, val)
	alertHandler := handlers.NewAlertHandler(alertService, log, val)
	recommendationHandler := handlers.NewRecommendationHandler(recommendationService, log, val)
	driftHandler := handlers.NewDriftHandler(driftService, log, val)
	anomalyHandler := handlers.NewAnomalyHandler(anomalyService, log, val)
	healthHandler := handlers.NewHealthHandler(db, log)

	// Setup router
	r := router.Setup(
		authHandler,
		resourceHandler,
		providerHandler,
		alertHandler,
		recommendationHandler,
		driftHandler,
		anomalyHandler,
		healthHandler,
		cfg,
		log,
	)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in goroutine
	go func() {
		log.Infof("Server listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Errorf("Server forced to shutdown: %v", err)
	}

	log.Info("Server stopped")
}
```

### 6. Database Migrations (0%)

**Create migrations directory structure**:

```sql
-- migrations/000001_init_schema.up.sql
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    username TEXT,
    full_name TEXT,
    role TEXT NOT NULL DEFAULT 'user',
    plan_type TEXT NOT NULL DEFAULT 'free',
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS provider_accounts (
    user_id INTEGER NOT NULL,
    provider TEXT NOT NULL,
    is_connected INTEGER NOT NULL DEFAULT 0,
    last_synced INTEGER,
    aws_access_key_id TEXT,
    aws_secret_access_key TEXT,
    aws_region TEXT,
    gcp_project_id TEXT,
    gcp_service_account_json TEXT,
    gcp_region TEXT,
    azure_tenant_id TEXT,
    azure_client_id TEXT,
    azure_client_secret TEXT,
    azure_subscription_id TEXT,
    azure_location TEXT,
    PRIMARY KEY (user_id, provider)
);

CREATE TABLE IF NOT EXISTS resources (
    user_id INTEGER NOT NULL,
    provider TEXT NOT NULL,
    resource_id TEXT NOT NULL,
    name TEXT,
    type TEXT,
    region TEXT,
    status TEXT,
    PRIMARY KEY (user_id, provider, resource_id)
);

CREATE TABLE IF NOT EXISTS alerts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    type TEXT,
    severity TEXT,
    title TEXT,
    description TEXT,
    resource TEXT,
    status TEXT,
    timestamp TEXT
);

CREATE TABLE IF NOT EXISTS recommendations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    type TEXT,
    priority TEXT,
    title TEXT,
    description TEXT,
    savings REAL,
    effort TEXT,
    impact TEXT,
    category TEXT,
    resources TEXT
);

CREATE TABLE IF NOT EXISTS security_drifts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    resource_id TEXT,
    drift_type TEXT,
    severity TEXT,
    details TEXT,
    detected_at TEXT,
    status TEXT
);

CREATE TABLE IF NOT EXISTS cost_anomalies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    resource_id TEXT,
    anomaly_type TEXT,
    severity TEXT,
    percentage INTEGER,
    previous_cost INTEGER,
    current_cost INTEGER,
    detected_at TEXT,
    status TEXT
);
```

```sql
-- migrations/000001_init_schema.down.sql
DROP TABLE IF EXISTS cost_anomalies;
DROP TABLE IF EXISTS security_drifts;
DROP TABLE IF EXISTS recommendations;
DROP TABLE IF EXISTS alerts;
DROP TABLE IF EXISTS resources;
DROP TABLE IF EXISTS provider_accounts;
DROP TABLE IF EXISTS users;
```

---

## 📋 Implementation Checklist

### Immediate Next Steps (in order):

- [ ] Create service implementations (7 files)
- [ ] Create remaining handlers (7 files)
- [ ] Create health handler
- [ ] Create router setup
- [ ] Create new main.go
- [ ] Create database migrations
- [ ] Test basic endpoints
- [ ] Add OAuth handlers from old code
- [ ] Integrate cloud provider code

### Testing Strategy:

1. **Unit Tests**: Test services and handlers
2. **Integration Tests**: Test with actual database
3. **E2E Tests**: Test complete flows

---

## 🚀 Quick Start Guide

Once you complete the remaining pieces:

```bash
# 1. Setup environment
cp .env.example .env
# Edit .env with your settings

# 2. Run migrations (if using golang-migrate)
make migrate-up

# 3. Build
make build

# 4. Run
make run

# 5. Test health endpoint
curl http://localhost:8080/healthz

# 6. Test auth
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'
```

---

## 📚 Code Examples Location

All code templates and examples are in:
- **REFACTORING.md** - Complete refactoring guide
- **THIS FILE** - Quick reference and templates

Use `internal/api/handlers/auth.go` and `internal/repository/postgres/*.go` as reference implementations.

---

**Status Summary**:
- ✅ Repositories: 100% Complete (7/7)
- ⏳ Services: 0% Complete (0/7)
- ⏳ Handlers: 14% Complete (1/7)
- ⏳ Router: 0% Complete
- ⏳ Main.go: 0% Complete
- ⏳ Migrations: 0% Complete

**Overall Progress**: ~40% Complete
