# InfraAudit Go Backend Refactoring Guide

This document outlines the refactoring from a monolithic structure to a production-ready, clean architecture structure.

## 📋 Table of Contents

1. [Overview](#overview)
2. [What Has Been Completed](#what-has-been-completed)
3. [New Directory Structure](#new-directory-structure)
4. [Architecture Patterns](#architecture-patterns)
5. [What Remains To Be Done](#what-remains-to-be-done)
6. [Migration Steps](#migration-steps)
7. [How to Use the New Structure](#how-to-use-the-new-structure)

---

## 🎯 Overview

The refactoring transforms the codebase from:
- **Before**: Monolithic 2,317-line `main.go` with all logic mixed together
- **After**: Clean architecture with proper separation of concerns, testability, and scalability

### Key Improvements

- ✅ **Proper layering**: Domain → Repository → Service → Handler
- ✅ **Dependency injection**: Services receive dependencies via constructors
- ✅ **Middleware**: Authentication, logging, rate limiting, recovery, CORS
- ✅ **Configuration management**: Environment-based, type-safe config
- ✅ **Error handling**: Structured errors with proper HTTP codes
- ✅ **Validation**: Request validation using go-playground/validator
- ✅ **Logging**: Structured logging with zerolog
- ✅ **DTOs**: Clear API contracts with request/response models
- ✅ **Docker support**: Production-ready docker-compose setup

---

## ✅ What Has Been Completed

### 1. Foundation Packages

#### Configuration Management (`internal/config/`)
- Type-safe configuration loading from environment variables
- Validation at startup
- Support for multiple database drivers (SQLite, PostgreSQL)
- OAuth, Redis, logging configuration

```go
// Usage
cfg, err := config.Load()
if err != nil {
    log.Fatal(err)
}
```

#### Logging (`internal/pkg/logger/`)
- Structured logging with zerolog
- Multiple log levels (debug, info, warn, error, fatal)
- JSON and console output formats
- Request ID tracking

```go
// Usage
log := logger.New(logger.Config{
    Level:  "info",
    Format: "json",
})
log.Info("Server started")
log.WithFields(map[string]interface{}{
    "user_id": 123,
    "action": "login",
}).Info("User logged in")
```

#### Error Handling (`internal/pkg/errors/`)
- Custom error types with HTTP status codes
- Error codes for client handling
- Structured error responses

```go
// Usage
return errors.NotFound("User")
return errors.ValidationError("Invalid input", validationErrors)
return errors.Internal("Database error", err)
```

#### Validation (`internal/pkg/validator/`)
- Request validation using go-playground/validator
- Custom validation messages
- Automatic JSON tag name extraction

```go
// Usage
validator := validator.New()
errs := validator.Validate(request)
if len(errs) > 0 {
    return errors.ValidationError("Validation failed", errs)
}
```

### 2. Domain Layer

Created domain models with clear separation:

#### Domain Structure
```
internal/domain/
├── user/           # User domain
│   ├── model.go       # User entity
│   ├── repository.go  # Repository interface
│   └── service.go     # Service interface
├── resource/       # Cloud resource domain
├── provider/       # Cloud provider accounts
├── alert/          # Security & operational alerts
├── recommendation/ # Cost & performance recommendations
├── drift/          # Security configuration drifts
└── anomaly/        # Cost anomalies
```

#### Domain Models Include:
- **User**: Authentication, profile, plan management
- **Resource**: Cloud resources (EC2, S3, GCE, etc.)
- **Provider**: Cloud provider accounts (AWS, GCP, Azure)
- **Alert**: Security and operational alerts
- **Recommendation**: Cost optimization suggestions
- **Drift**: Security configuration drifts
- **Anomaly**: Cost anomalies and spikes

Each domain has:
- **Model**: Data structure
- **Repository Interface**: Data access methods
- **Service Interface**: Business logic methods

### 3. Middleware Layer

Complete middleware stack in `internal/api/middleware/`:

#### Available Middleware

**Authentication (`auth.go`)**
```go
// Require authentication
router.Use(middleware.AuthMiddleware(jwtSecret))

// Optional authentication
router.Use(middleware.OptionalAuthMiddleware(jwtSecret))

// Get user from context
userID, ok := middleware.GetUserID(r)
```

**Logging (`logger.go`)**
```go
router.Use(middleware.Logger(log))
// Logs: method, path, status, duration, bytes, IP, user agent
```

**Rate Limiting (`rate_limit.go`)**
```go
// IP-based rate limiting
router.Use(middleware.RateLimit(10, 20)) // 10 req/sec, burst 20

// User-based rate limiting
router.Use(middleware.UserRateLimit(100, 200))
```

**Request ID (`request_id.go`)**
```go
router.Use(middleware.RequestID())
requestID := middleware.GetRequestID(r)
```

**Recovery (`recovery.go`)**
```go
router.Use(middleware.Recovery(log))
// Recovers from panics and returns 500 errors
```

**CORS (`cors.go`)**
```go
router.Use(middleware.DefaultCORS(frontendURL))
```

### 4. DTOs (Data Transfer Objects)

API request/response models in `internal/api/dto/`:

```go
// Example: Alert creation
type CreateAlertRequest struct {
    Type        string `json:"type" validate:"required,oneof=security compliance performance"`
    Severity    string `json:"severity" validate:"required,oneof=critical high medium low"`
    Title       string `json:"title" validate:"required"`
    Description string `json:"description" validate:"required"`
}
```

DTOs for all domains:
- `auth.go`: Login, register, token refresh
- `user.go`: User profile, updates
- `resource.go`: Resource CRUD
- `provider.go`: Provider connection, sync
- `alert.go`: Alert management
- `recommendation.go`: Recommendations
- `drift.go`: Security drifts
- `anomaly.go`: Cost anomalies

### 5. Example Handler

Created `internal/api/handlers/auth.go` demonstrating:
- Dependency injection
- Request validation
- Error handling
- Response formatting
- Swagger documentation comments

```go
// Handler structure
type AuthHandler struct {
    userService user.Service
    config      *config.Config
    logger      *logger.Logger
    validator   *validator.Validator
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    // 1. Parse request
    // 2. Validate
    // 3. Call service
    // 4. Return response
}
```

### 6. Development Environment

#### `.env.example`
Complete environment variable template with:
- Server configuration
- Database settings (SQLite & PostgreSQL)
- JWT & session secrets
- OAuth credentials
- Redis configuration
- Logging settings
- Cloud provider API keys

#### `docker-compose.yml`
Full development stack:
- API service
- PostgreSQL database
- Redis cache
- Prometheus (optional)
- Grafana (optional)

```bash
# Start stack
docker-compose up -d

# View logs
docker-compose logs -f api

# Stop stack
docker-compose down
```

#### `Makefile.new`
Comprehensive build automation:
- `make build`: Build binaries
- `make run`: Run locally
- `make test`: Run tests
- `make lint`: Run linters
- `make docker-build`: Build Docker image
- `make docker-compose-up`: Start stack
- And many more...

---

## 📁 New Directory Structure

```
infraudit-go/
├── cmd/                                # Entry points
│   ├── api/                           # API server
│   │   └── main.go                    # NEW: Clean main.go with DI
│   ├── worker/                        # Background worker
│   └── migrate/                       # Migration tool
│
├── internal/                          # Private code
│   ├── api/                          # API layer
│   │   ├── handlers/                 # ✅ HTTP handlers (auth.go created)
│   │   ├── middleware/               # ✅ ALL middleware complete
│   │   ├── router/                   # TODO: Router setup
│   │   └── dto/                      # ✅ ALL DTOs complete
│   │
│   ├── domain/                       # ✅ ALL domains complete
│   │   ├── user/
│   │   ├── resource/
│   │   ├── provider/
│   │   ├── alert/
│   │   ├── recommendation/
│   │   ├── drift/
│   │   └── anomaly/
│   │
│   ├── repository/                   # TODO: Repository implementations
│   │   └── postgres/                 # TODO: PostgreSQL repos
│   │
│   ├── providers/                    # TODO: Move & refactor cloud SDKs
│   │   ├── aws/
│   │   ├── gcp/
│   │   └── azure/
│   │
│   ├── auth/                         # ✅ JWT auth (existing code)
│   ├── integrations/                 # TODO: Move integrations
│   ├── config/                       # ✅ Complete
│   └── pkg/                          # ✅ All utilities complete
│       ├── logger/
│       ├── errors/
│       ├── validator/
│       └── utils/
│
├── migrations/                       # TODO: Create SQL migrations
├── tests/                           # TODO: Add tests
├── api/                             # TODO: Swagger docs
├── deployments/                     # ✅ Docker files
│   ├── docker/
│   │   └── Dockerfile              # ✅ Complete
│   └── kubernetes/                  # TODO: K8s manifests
│
├── .env.example                     # ✅ Complete
├── docker-compose.yml               # ✅ Complete
├── Makefile.new                     # ✅ Complete
└── REFACTORING.md                   # ✅ This file
```

---

## 🏗️ Architecture Patterns

### Clean Architecture Layers

```
┌─────────────────────────────────────┐
│     HTTP Handlers (Presentation)    │
│  Parse requests, validate, respond  │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│      Services (Business Logic)      │
│   Orchestrate operations, policies  │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│    Repositories (Data Access)       │
│     Database queries, caching       │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│         Domain Models               │
│    Core entities and interfaces     │
└─────────────────────────────────────┘
```

### Request Flow

```
HTTP Request
    ↓
Middleware (Request ID, CORS, etc.)
    ↓
Authentication Middleware
    ↓
Rate Limiting
    ↓
Logging Middleware
    ↓
Handler
    ↓
Request Validation (DTO)
    ↓
Service Layer
    ↓
Repository Layer
    ↓
Database
    ↓
Response (with proper error handling)
```

### Dependency Injection Pattern

```go
// main.go
func main() {
    // Load config
    cfg := config.Load()

    // Initialize infrastructure
    log := logger.New(cfg.Logging)
    db := setupDatabase(cfg.Database)
    validator := validator.New()

    // Initialize repositories
    userRepo := postgres.NewUserRepository(db)

    // Initialize services
    userService := services.NewUserService(userRepo, log)

    // Initialize handlers
    authHandler := handlers.NewAuthHandler(userService, cfg, log, validator)

    // Setup router
    r := router.Setup(authHandler, cfg, log)

    // Start server
    server := &http.Server{
        Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
        Handler: r,
    }
    server.ListenAndServe()
}
```

---

## 🚧 What Remains To Be Done

### Priority 1: Core Functionality

#### 1. Repository Implementations
Create PostgreSQL implementations in `internal/repository/postgres/`:

**Files needed:**
- `user.go`: Implement `user.Repository` interface
- `resource.go`: Implement `resource.Repository` interface
- `provider.go`: Implement `provider.Repository` interface
- `alert.go`: Implement `alert.Repository` interface
- `recommendation.go`: Implement `recommendation.Repository` interface
- `drift.go`: Implement `drift.Repository` interface
- `anomaly.go`: Implement `anomaly.Repository` interface

**Example:**
```go
// internal/repository/postgres/user.go
type UserRepository struct {
    db *sql.DB
}

func (r *UserRepository) Create(ctx context.Context, user *user.User) error {
    query := `INSERT INTO users (email, username, role, plan_type) VALUES ($1, $2, $3, $4) RETURNING id`
    return r.db.QueryRowContext(ctx, query, user.Email, user.Username, user.Role, user.PlanType).Scan(&user.ID)
}
```

#### 2. Service Implementations
Create service implementations in `internal/services/`:

**Files needed:**
- `user_service.go`
- `resource_service.go`
- `provider_service.go`
- `alert_service.go`
- `recommendation_service.go`
- `drift_service.go`
- `anomaly_service.go`

**Example:**
```go
// internal/services/user_service.go
type UserService struct {
    repo   user.Repository
    logger *logger.Logger
}

func (s *UserService) Create(ctx context.Context, email string) (*user.User, error) {
    // Validation
    // Business logic
    // Call repository
    return s.repo.Create(ctx, newUser)
}
```

#### 3. Complete Handlers
Create handlers for each domain in `internal/api/handlers/`:

**Files needed:**
- `health.go`: Health check endpoints
- `resources.go`: Resource CRUD
- `providers.go`: Provider management
- `alerts.go`: Alert management
- `recommendations.go`: Recommendation management
- `drifts.go`: Drift management
- `anomalies.go`: Anomaly management

Use `auth.go` as a template for structure and patterns.

#### 4. Router Setup
Create `internal/api/router/router.go`:

```go
func Setup(
    authHandler *handlers.AuthHandler,
    resourceHandler *handlers.ResourceHandler,
    // ... other handlers
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
        r.Post("/api/auth/login", authHandler.Login)
        r.Post("/api/auth/register", authHandler.Register)
        r.Get("/healthz", healthHandler.Healthz)
    })

    // Protected routes
    r.Group(func(r chi.Router) {
        r.Use(middleware.AuthMiddleware(cfg.Auth.JWTSecret))

        r.Get("/api/auth/me", authHandler.Me)
        r.Post("/api/auth/logout", authHandler.Logout)

        r.Route("/api/resources", func(r chi.Router) {
            r.Get("/", resourceHandler.List)
            r.Post("/", resourceHandler.Create)
            r.Get("/{id}", resourceHandler.Get)
            r.Patch("/{id}", resourceHandler.Update)
            r.Delete("/{id}", resourceHandler.Delete)
        })

        // ... more routes
    })

    return r
}
```

#### 5. New main.go
Create `cmd/api/main.go` with dependency injection:

```go
package main

import (
    "context"
    "fmt"
    "log"
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
        log.Fatalf("Failed to load config: %v", err)
    }

    // Initialize logger
    appLogger := logger.New(logger.Config{
        Level:  cfg.Logging.Level,
        Format: cfg.Logging.Format,
    })

    // Initialize validator
    val := validator.New()

    // Setup database
    db, err := setupDatabase(cfg)
    if err != nil {
        appLogger.Fatalf("Failed to setup database: %v", err)
    }
    defer db.Close()

    // Initialize repositories
    userRepo := postgres.NewUserRepository(db)
    resourceRepo := postgres.NewResourceRepository(db)
    // ... other repos

    // Initialize services
    userService := services.NewUserService(userRepo, appLogger)
    resourceService := services.NewResourceService(resourceRepo, appLogger)
    // ... other services

    // Initialize handlers
    authHandler := handlers.NewAuthHandler(userService, cfg, appLogger, val)
    resourceHandler := handlers.NewResourceHandler(resourceService, appLogger, val)
    // ... other handlers

    // Setup router
    r := router.Setup(authHandler, resourceHandler, cfg, appLogger)

    // Create server
    server := &http.Server{
        Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
        Handler:      r,
        ReadTimeout:  cfg.Server.ReadTimeout,
        WriteTimeout: cfg.Server.WriteTimeout,
    }

    // Start server in goroutine
    go func() {
        appLogger.Infof("Server starting on %s", server.Addr)
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            appLogger.Fatalf("Server failed: %v", err)
        }
    }()

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    appLogger.Info("Shutting down server...")

    ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        appLogger.Fatalf("Server forced to shutdown: %v", err)
    }

    appLogger.Info("Server exited")
}

func setupDatabase(cfg *config.Config) (*sql.DB, error) {
    // Database setup logic
    // ...
    return db, nil
}
```

### Priority 2: Cloud Provider Integration

#### 6. Refactor Provider Code
Move and refactor existing provider code:

**From:** `internal/providers/{aws,gcp,azure}.go`
**To:** `internal/providers/{aws,gcp,azure}/client.go`

Split into service-specific files:
- `internal/providers/aws/ec2.go`
- `internal/providers/aws/s3.go`
- `internal/providers/gcp/compute.go`
- etc.

Create unified interface:
```go
// internal/providers/provider.go
type Provider interface {
    ListResources(ctx context.Context) ([]*resource.Resource, error)
    GetCredentials() Credentials
    TestConnection(ctx context.Context) error
}
```

### Priority 3: Database Migrations

#### 7. Create Migration System
Set up golang-migrate:

```bash
# Install
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Create initial migration
make migrate-create name=init_schema
```

**migrations/000001_init_schema.up.sql:**
```sql
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    username VARCHAR(100),
    full_name VARCHAR(255),
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    plan_type VARCHAR(50) NOT NULL DEFAULT 'free',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create other tables...
```

### Priority 4: Testing

#### 8. Add Unit Tests
Create tests for each layer:

```
tests/
├── unit/
│   ├── services/
│   │   └── user_service_test.go
│   ├── handlers/
│   │   └── auth_handler_test.go
│   └── middleware/
│       └── auth_test.go
├── integration/
│   ├── auth_test.go
│   └── resources_test.go
└── fixtures/
    └── test_data.go
```

**Example test:**
```go
// tests/unit/services/user_service_test.go
func TestUserService_Create(t *testing.T) {
    // Setup
    mockRepo := &MockUserRepository{}
    log := logger.New(logger.Config{Level: "debug"})
    service := services.NewUserService(mockRepo, log)

    // Test
    user, err := service.Create(context.Background(), "test@example.com")

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, "test@example.com", user.Email)
}
```

### Priority 5: Documentation

#### 9. Swagger/OpenAPI
Add Swagger documentation:

```bash
# Install swag
go install github.com/swaggo/swag/cmd/swag@latest

# Generate docs
make swagger
```

Swagger annotations are already in `auth.go` handler as examples.

#### 10. API Documentation
Create `docs/api.md` with:
- Endpoint list
- Request/response examples
- Authentication guide
- Error codes

---

## 🔄 Migration Steps

### Step-by-Step Guide

#### Phase 1: Prepare
1. ✅ Review this guide
2. ✅ Backup existing database
3. Copy `.env.example` to `.env` and configure
4. Review new structure

#### Phase 2: Implement Core (You are here)
1. Create repository implementations
2. Create service implementations
3. Create remaining handlers
4. Setup router
5. Create new main.go

#### Phase 3: Test
1. Start with docker-compose: `make docker-compose-up`
2. Test health endpoint
3. Test authentication
4. Test each API endpoint
5. Run unit tests: `make test`

#### Phase 4: Migrate Data
1. Create migrations
2. Run migrations: `make migrate-up`
3. Seed test data if needed
4. Verify data integrity

#### Phase 5: Replace Old Code
1. Backup old main.go: `mv cmd/api/main.go cmd/api/main.go.old`
2. Remove old settings service (if not needed)
3. Update imports across codebase
4. Remove unused code

#### Phase 6: Deploy
1. Build: `make build`
2. Test locally: `make run`
3. Build Docker: `make docker-build`
4. Deploy to staging
5. Run smoke tests
6. Deploy to production

---

## 📚 How to Use the New Structure

### Creating a New Feature

Let's say you want to add a "Projects" feature:

#### 1. Define Domain Model
```go
// internal/domain/project/model.go
package project

type Project struct {
    ID          int64
    UserID      int64
    Name        string
    Description string
    CreatedAt   time.Time
}
```

#### 2. Define Repository Interface
```go
// internal/domain/project/repository.go
package project

type Repository interface {
    Create(ctx context.Context, project *Project) error
    GetByID(ctx context.Context, id int64) (*Project, error)
    // ...
}
```

#### 3. Implement Repository
```go
// internal/repository/postgres/project.go
type ProjectRepository struct {
    db *sql.DB
}

func (r *ProjectRepository) Create(ctx context.Context, p *Project) error {
    // Implementation
}
```

#### 4. Define Service Interface
```go
// internal/domain/project/service.go
package project

type Service interface {
    Create(ctx context.Context, userID int64, name string) (*Project, error)
    // ...
}
```

#### 5. Implement Service
```go
// internal/services/project_service.go
type ProjectService struct {
    repo   project.Repository
    logger *logger.Logger
}

func (s *ProjectService) Create(ctx context.Context, userID int64, name string) (*Project, error) {
    // Validation
    // Business logic
    // Call repository
}
```

#### 6. Create DTOs
```go
// internal/api/dto/project.go
type CreateProjectRequest struct {
    Name        string `json:"name" validate:"required"`
    Description string `json:"description"`
}

type ProjectDTO struct {
    ID          int64  `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description"`
}
```

#### 7. Create Handler
```go
// internal/api/handlers/project.go
type ProjectHandler struct {
    service   project.Service
    logger    *logger.Logger
    validator *validator.Validator
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
    // Parse, validate, call service, respond
}
```

#### 8. Add Routes
```go
// internal/api/router/router.go
r.Route("/api/projects", func(r chi.Router) {
    r.Use(middleware.AuthMiddleware(cfg.Auth.JWTSecret))
    r.Get("/", projectHandler.List)
    r.Post("/", projectHandler.Create)
})
```

### Best Practices

#### Error Handling
```go
// In handlers
if err := someOperation(); err != nil {
    h.logger.ErrorWithErr(err, "Operation failed")
    utils.WriteError(w, errors.Internal("Operation failed", err))
    return
}
```

#### Logging
```go
// Structured logging
h.logger.WithFields(map[string]interface{}{
    "user_id":    userID,
    "resource_id": resourceID,
    "action":     "delete",
}).Info("Resource deleted")
```

#### Context Usage
```go
// Always pass context through layers
func (s *Service) DoSomething(ctx context.Context, id int64) error {
    // Check context cancellation
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }

    // Continue with operation
    return s.repo.Update(ctx, id)
}
```

#### Transaction Handling
```go
// In repository layer
func (r *Repository) UpdateWithRelations(ctx context.Context, item *Item) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // Multiple operations
    if err := r.updateItem(ctx, tx, item); err != nil {
        return err
    }

    if err := r.updateRelations(ctx, tx, item.ID); err != nil {
        return err
    }

    return tx.Commit()
}
```

---

## 🎓 Learning Resources

### Go Best Practices
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)

### Clean Architecture
- [The Clean Architecture by Uncle Bob](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Implementing Clean Architecture](https://8thlight.com/blog/uncle-bob/2011/11/22/Clean-Architecture.html)

### Domain-Driven Design
- [Domain-Driven Design Reference](https://www.domainlanguage.com/ddd/reference/)

---

## 📞 Support

If you have questions during the migration:

1. Check this guide first
2. Review the example `auth.go` handler
3. Look at existing implementations in other similar projects
4. Check Go documentation

---

## ✅ Completion Checklist

Use this to track your progress:

- [x] Foundation packages created
- [x] Domain models defined
- [x] Middleware layer complete
- [x] DTOs created
- [x] Example handler created
- [x] Docker setup complete
- [x] Makefile updated
- [ ] Repository implementations
- [ ] Service implementations
- [ ] All handlers created
- [ ] Router setup complete
- [ ] New main.go created
- [ ] OAuth handlers migrated
- [ ] Cloud provider code refactored
- [ ] Database migrations created
- [ ] Unit tests added
- [ ] Integration tests added
- [ ] Swagger documentation
- [ ] Old code removed
- [ ] Deployment tested

---

**Last Updated**: 2025-10-30
**Version**: 1.0
**Status**: Phase 1 Complete - Ready for Phase 2 Implementation
