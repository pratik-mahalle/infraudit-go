# InfraAudit Go Backend (Scaffold)

This is a minimal Go HTTP API scaffold to support the InfraAudit frontend.
It exposes the key auth endpoints used by the UI and sets up CORS.

## Endpoints (stubs)
- GET `/api/status`
- POST `/api/login`
- POST `/api/register`
- POST `/api/logout`
- GET `/api/user`
- POST `/api/start-trial`
- GET `/api/trial-status`
- GET `/api/auth/{provider}`
- GET `/api/auth/{provider}/callback`

These return demo data for now so the frontend can integrate while the full backend is implemented.

## Run locally
```bash
cd backend-go
cp .env.example .env  # adjust as needed
# Use Go 1.22+
go mod tidy
go run ./cmd/api
```
The API will listen on `:5000` by default and allow CORS from `FRONTEND_URL`.

## Docker
```bash
docker build -t infraaudit-backend-go ./backend-go
# Note: for real config, mount envs or use a proper runtime config loader
```

## Next steps (production-ready)
- Replace stubs with real implementations:
  - Postgres connection via `pgxpool`
  - Migrations via `golang-migrate`
  - Password hashing with scrypt (compatible with current format)
  - JWT mint/refresh/revoke using `github.com/golang-jwt/jwt/v5`
  - OAuth via `goth` or `x/oauth2` for Google/GitHub
- Mirror all remaining endpoints used by the frontend
- Add structured logging, request IDs, metrics
- CI/CD and containerization 