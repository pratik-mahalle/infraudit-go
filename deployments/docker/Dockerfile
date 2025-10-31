# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/api ./cmd/api

# Final stage
FROM gcr.io/distroless/static:nonroot

# Copy binary from builder
COPY --from=builder /app/bin/api /api

# Expose port
EXPOSE 8080

# Create data directory
VOLUME /data

# Run as non-root user
USER nonroot:nonroot

# Run the application
ENTRYPOINT ["/api"]
