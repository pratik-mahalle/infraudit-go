# Guide: Converting the InfraAudit Backend from Node.js to Go

This document outlines a step-by-step plan for converting the InfraAudit backend from its current Node.js/TypeScript implementation to a Go (Golang) backend. It covers architecture, feature mapping, migration strategies, and recommended Go libraries.

---

## 1. Current Backend Overview

**Node.js Technologies Used:**
- **HTTP Server:** Express.js
- **Database:** PostgreSQL, accessed via Drizzle ORM
- **Authentication:** Passport.js (local strategy), JWT, session support (express-session, connect-pg-simple)
- **Containerization & Orchestration:** Docker, Kubernetes
- **Other:** dotenv, custom service classes (e.g., AWS integration), shared TypeScript types

**Core Features:**
- RESTful APIs for resources, users, orgs, security drifts, cost anomalies, recommendations
- JWT-based and session-based authentication
- PostgreSQL-backed data storage
- Cloud provider integrations (AWS, Azure, GCP)
- Multi-tenancy (organizations, users, RBAC)
- Automated database seeding and migrations

---

## 2. High-Level Conversion Strategy

### 2.1. Choose Go Libraries/Frameworks

| Node.js Dependency           | Go Equivalent / Recommendation                        |
|------------------------------|------------------------------------------------------|
| Express.js                   | net/http, gorilla/mux, or chi                        |
| Drizzle ORM                  | GORM, sqlc, ent, or pgx for PostgreSQL               |
| dotenv/config                | github.com/joho/godotenv, os.Getenv                  |
| Passport.js (auth)           | github.com/golang-jwt/jwt/v5, gorilla/sessions, bcrypt|
| express-session              | gorilla/sessions, securecookie                       |
| connect-pg-simple            | gorilla/sessions (PostgreSQL backend)                |
| TypeScript types             | Go structs                                           |
| Docker, Kubernetes           | Same, with updated Dockerfile for Go                 |

### 2.2. Project Structure Proposal

```
InfraAudit/
├── api/                # Go server (replace server/)
│   ├── main.go
│   ├── handlers/
│   ├── models/
│   ├── db/
│   ├── middleware/
│   └── services/
├── client/             # React frontend (unchanged)
├── shared/             # Shared OpenAPI specs or protobuf (optional)
├── k8s/                # Kubernetes configs (adapt as needed)
├── Dockerfile          # Update for Go build
└── docker-compose.yml  # Update services if necessary
```

---

## 3. Detailed Migration Steps

### 3.1. API Layer: Express.js → Go HTTP Router

- **Define all REST endpoints** in Go using gorilla/mux or chi.
- Map each Express route/middleware to its Go equivalent.
- Use Go's context to pass request/user info.

### 3.2. Database Access: Drizzle ORM → GORM/sqlc/pgx

- **Define Go structs** for each table (`users`, `organizations`, `resources`, etc.).
- Implement CRUD operations as Go methods.
- Use database migration tool (e.g., golang-migrate or goose).

### 3.3. Authentication

- **Local auth:** Use bcrypt to hash and verify passwords.
- **JWT support:** Use `github.com/golang-jwt/jwt/v5`.
- **Session management:** Use gorilla/sessions with a PostgreSQL backend.

### 3.4. Business Logic & Services

- Re-implement Node.js service classes (e.g., AWS integration) in Go, using the AWS SDK for Go.
- Port core logic for resource monitoring, cost analysis, drift detection, etc.

### 3.5. Error Handling & Logging

- Use Go idioms for error return values.
- Use a structured logger (e.g., zap, logrus).

### 3.6. Configuration

- Read `.env` or use environment variables directly via `os.Getenv`.
- Implement config struct and validation.

### 3.7. Testing

- Rewrite API and unit tests using Go's `testing` package.
- Optionally, use Postman or k6 for integration tests.

---

## 4. Feature-by-Feature Mapping

| Feature                    | Node.js Implementation                   | Go Migration Plan                                         |
|----------------------------|------------------------------------------|----------------------------------------------------------|
| User CRUD, auth            | Express routes, Passport.js, JWT         | REST endpoints, bcrypt, JWT, gorilla/sessions            |
| Organization CRUD          | Express routes, ORM                      | REST endpoints, Go structs, GORM/sqlc                    |
| Resource CRUD/queries      | Express, ORM                             | REST endpoints, Go structs, GORM/sqlc                    |
| Security drifts, anomalies | Express, ORM, custom logic               | REST endpoints, Go structs, business logic in Go         |
| AWS/Cloud integration      | TypeScript services                      | Use AWS SDK for Go, implement service interfaces         |
| Database migrations        | Drizzle migration scripts                | Use golang-migrate or goose with SQL/Go migration files  |
| Seeding                    | Custom TypeScript seed scripts           | Implement `main.go` seeder with CLI flags                |
| Testing                    | Jest, k6                                 | Go `testing`, k6 (unchanged)                             |

---

## 5. Example: Mapping a Node.js Route to Go

**Node.js/Express:**
```typescript
app.get('/api/resources', async (req, res) => {
  const resources = await storage.getResources();
  res.json(resources);
});
```
**Go/gorilla-mux:**
```go
func GetResources(w http.ResponseWriter, r *http.Request) {
    resources, err := db.GetResources()
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(resources)
}
```

---

## 6. Migration Tips

- Start with data models and DB migrations.
- Scaffold all endpoints as stubs, then incrementally port business logic.
- Write integration tests for Go endpoints to match current Node.js behavior.
- Replace TypeScript interfaces with Go structs. Use `json` tags for marshaling.
- Use Go interfaces to abstract storage and services for testability.
- Port only backend logic—frontend React app remains unchanged.

---

## 7. Recommended Go Libraries

- Routing: [chi](https://github.com/go-chi/chi) or [gorilla/mux](https://github.com/gorilla/mux)
- Database: [GORM](https://gorm.io/), [sqlc](https://sqlc.dev/), [pgx](https://github.com/jackc/pgx)
- Auth: [golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt), [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)
- Sessions: [gorilla/sessions](https://github.com/gorilla/sessions)
- AWS SDK: [aws-sdk-go-v2](https://github.com/aws/aws-sdk-go-v2)
- Config: [joho/godotenv](https://github.com/joho/godotenv)
- Logging: [uber-go/zap](https://github.com/uber-go/zap) or [sirupsen/logrus](https://github.com/sirupsen/logrus)
- Migrations: [golang-migrate/migrate](https://github.com/golang-migrate/migrate)

---

## 8. References

- [InfraAudit README](./README.md)
- [GORM Documentation](https://gorm.io/docs/)
- [Go REST API Example](https://github.com/golang-standards/project-layout)
- [Migrating from Node.js to Go](https://medium.com/@attilacsordas/migrating-from-nodejs-to-golang-959dcd6e8d33)

---

## 9. Next Steps

1. Scaffold the Go project structure.
2. Define Go structs for all models and migrate schema.
3. Implement authentication and core endpoints.
4. Port cloud provider integrations.
5. Continuously test and validate against the existing API.

---

*This guide should serve as a comprehensive starting point for a systematic migration from Node.js to Go for the InfraAudit backend.*