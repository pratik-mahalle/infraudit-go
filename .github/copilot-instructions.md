# GitHub Copilot Instructions for InfraAudit

This is a Go-based cloud infrastructure auditing and security platform that helps organizations monitor and secure their cloud resources across AWS, Azure, and GCP. Please follow these guidelines when contributing:

## Code Standards

### Required Before Each Commit
- Run `make fmt` before committing any changes to ensure proper code formatting
- This will run gofmt on all Go files to maintain consistent style

### Development Flow (Current Makefile)
- Build: `make build`
- Run: `make run`
- Format: `make fmt` (format all Go files with gofmt)
- Lint: `make lint` (run go vet)
- Swagger docs: `make swagger` (regenerate API documentation)
- Clean: `make clean`

### Additional Targets (Makefile.new - Extended Build System)
The repository includes an extended Makefile (`Makefile.new`) with more comprehensive targets. 
This is a newer, more complete build configuration that may eventually replace the main Makefile.
You can use it by specifying `-f Makefile.new`:
- Test: `make -f Makefile.new test` (run all tests with race detector)
- Test with coverage: `make -f Makefile.new test-coverage`
- Pre-commit check: `make -f Makefile.new pre-commit` (runs fmt, lint, and test)
- Install dev tools: `make -f Makefile.new install-tools`
- Security scan: `make -f Makefile.new security-scan`

### Manual Testing
If the test target is not available in your Makefile:
- Run tests manually: `go test -v -race ./...`
- Run with coverage: `go test -v -race -coverprofile=coverage.out ./...`

## Repository Structure
- `cmd/`: Main service entry points and executables
  - `cmd/api/`: Main API server application
  - `cmd/migrate/`: Database migration tool
- `internal/`: Private application code
  - `internal/auth/`: Authentication and authorization
  - `internal/config/`: Configuration management
  - `internal/db/`: Database connection and utilities
  - `internal/detector/`: Drift detection logic
  - `internal/domain/`: Domain models and business logic (alerts, anomalies, resources, etc.)
  - `internal/integrations/`: Third-party service integrations
  - `internal/pkg/`: Internal shared packages (logger, validator, errors, utils)
  - `internal/providers/`: Cloud provider implementations (AWS, Azure, GCP)
  - `internal/repository/`: Data access layer (PostgreSQL)
  - `internal/scanners/`: Security and compliance scanners
  - `internal/services/`: Business logic services
- `pkg/`: Public packages that can be imported by external projects
  - `pkg/client/`: Go client library for the API
- `docs/`: Auto-generated Swagger/OpenAPI documentation (DO NOT edit manually)
- `migrations/`: Database migration files
- `deployments/`: Deployment configurations (Docker, etc.)
- `scripts/`: Utility scripts
- `settings/`: Settings management UI/tool

## Key Guidelines

### Go Best Practices
1. Follow Go best practices and idiomatic patterns
2. Use standard Go project layout conventions
3. Maintain existing code structure and organization
4. Use dependency injection patterns where appropriate
5. Keep functions small and focused
6. Use meaningful variable and function names

### Testing
1. Write unit tests for new functionality
2. Use table-driven tests when possible
3. Use the race detector (`-race` flag) to catch concurrency issues
4. Aim for good test coverage on business logic
5. Test files should be in the same package as the code they test (use `_test.go` suffix)

### Documentation
1. Document all public APIs using godoc conventions
2. Add comments for complex logic or non-obvious code
3. Use Swagger annotations for API endpoints
4. Run `make swagger` after updating API handlers to regenerate documentation
5. Suggest updates to `docs/` folder for architectural changes

### API Development
1. Use Swagger annotations for all HTTP endpoints
2. Follow RESTful conventions
3. Always validate input using the internal validator package
4. Return consistent error responses
5. Use proper HTTP status codes
6. Implement proper authentication and authorization

### Cloud Provider Integration
1. Implement provider-specific logic in `internal/providers/`
2. Use interfaces to abstract provider differences
3. Handle provider-specific errors gracefully
4. Test with mock providers when possible
5. Document provider-specific requirements and limitations

### Database
1. Use migrations for schema changes (see `migrations/` directory)
2. Create new migrations with descriptive names
3. Test migrations both up and down
4. Use prepared statements to prevent SQL injection
5. Handle database errors appropriately

### Security
1. Never commit secrets or credentials
2. Use environment variables for sensitive configuration
3. Validate and sanitize all user input
4. Follow OWASP security guidelines
5. Use the auth middleware for protected endpoints
6. Implement proper RBAC (Role-Based Access Control)

### Error Handling
1. Use structured errors with context
2. Log errors with appropriate severity levels
3. Return user-friendly error messages
4. Don't expose internal implementation details in errors
5. Use the internal logger package consistently

### Code Style
1. Use `gofmt` for formatting (automatically done by `make fmt`)
2. Follow standard Go naming conventions
3. Group imports logically (standard library, external, internal)
4. Keep line length reasonable (aim for 80-120 characters)
5. Use early returns to reduce nesting

## Development Workflow

1. Create a feature branch from `main`
2. Make changes following the guidelines above
3. Run `make fmt` to format your code
4. Run `make lint` to check for issues
5. Run tests manually with `go test -v -race ./...` (or use `make -f Makefile.new test`)
6. If you modified API handlers with Swagger annotations, run `make swagger` to update documentation
7. Commit your changes with a descriptive message
8. Push and create a pull request

## Common Tasks

### Adding a New API Endpoint
1. Add handler function following the existing patterns in the codebase
2. Add Swagger annotations to the handler (see Swagger tag definitions in `cmd/api/main.go`)
3. Register the route in the router setup
4. Run `make swagger` to regenerate API documentation
5. Add tests for the new endpoint

### Adding Support for a New Cloud Provider
1. Create provider implementation in `internal/providers/`
2. Implement the provider interface
3. Add provider-specific scanners in `internal/scanners/`
4. Update integrations in `internal/integrations/`
5. Add tests and documentation

### Creating a Database Migration
1. Create migration files in `migrations/` directory
2. Name files with timestamp and description
3. Test both up and down migrations
4. Update relevant repository code if needed

## Tips
- Use the internal logger package for all logging
- Follow existing patterns in the codebase
- Ask for clarification if requirements are unclear
- Consider backward compatibility when making changes
- Document breaking changes clearly
