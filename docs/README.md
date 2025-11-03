# API Documentation

This directory contains auto-generated Swagger/OpenAPI documentation for the InfraAudit API.

## Files

- **`docs.go`** - Go package with embedded Swagger documentation
- **`swagger.json`** - OpenAPI 3.0 specification in JSON format
- **`swagger.yaml`** - OpenAPI 3.0 specification in YAML format

## ⚠️ Do Not Edit Manually

These files are **automatically generated** from code annotations in the handler files.
Any manual changes will be overwritten.

## How to Regenerate

### Locally

```bash
# Generate Swagger docs
make swagger

# Or manually:
swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal
```

### CI/CD

Swagger documentation is automatically:
1. **Generated** during CI builds
2. **Validated** to ensure it's up to date
3. **Rejected** if docs are outdated (you must run `make swagger` and commit)

## Adding/Updating Endpoints

To document a new endpoint or update existing documentation:

1. Add Swagger annotations to your handler function:

```go
// List returns all alerts
// @Summary List alerts
// @Description Get a paginated list of security alerts
// @Tags Alerts
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} utils.PaginatedResponse{data=[]dto.AlertDTO}
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /alerts [get]
func (h *AlertHandler) List(w http.ResponseWriter, r *http.Request) {
    // handler code...
}
```

2. Regenerate documentation:

```bash
make swagger
```

3. Commit the changes:

```bash
git add docs/
git commit -m "docs: update API documentation"
```

## Viewing Documentation

### Option 1: Swagger UI (When Server is Running)

Start the server:
```bash
make run
```

Open in browser:
```
http://localhost:8080/swagger/index.html
```

### Option 2: Export for Frontend

```bash
# Get OpenAPI JSON
curl http://localhost:8080/swagger/doc.json > openapi.json

# Or use the generated file directly
cp docs/swagger.json openapi.json
```

### Option 3: Use with Tools

Import `swagger.json` or `swagger.yaml` into:
- [Postman](https://www.postman.com/) - API testing
- [Swagger Editor](https://editor.swagger.io/) - Online viewer/editor
- [OpenAPI Generator](https://openapi-generator.tech/) - Client code generation
- [Stoplight](https://stoplight.io/) - API documentation platform

## Auto-Generation Setup (Optional)

To automatically generate docs before every commit:

```bash
./scripts/setup-swagger-hook.sh
```

This installs a git pre-commit hook that runs `make swagger` automatically.

## Annotation Reference

Common Swagger annotations:

- `@Summary` - Brief endpoint description
- `@Description` - Detailed description
- `@Tags` - Group endpoints by tag
- `@Accept` - Request content type (e.g., `json`)
- `@Produce` - Response content type (e.g., `json`)
- `@Param` - Parameter definition (path, query, body)
- `@Success` - Success response with status code and type
- `@Failure` - Error response with status code and type
- `@Security` - Authentication requirement (e.g., `BearerAuth`)
- `@Router` - Route path and HTTP method

For more details, see: https://github.com/swaggo/swag#declarative-comments-format
