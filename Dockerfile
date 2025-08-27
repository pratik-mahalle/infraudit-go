# Multi-stage build for InfraAudit Go API

FROM golang:1.24 AS builder
WORKDIR /app

# Enable static build (no CGO) so we can use a minimal runtime image
ENV CGO_ENABLED=0 \
    GO111MODULE=on

# Pre-cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build the API binary
COPY . .
RUN go build -o /out/api ./cmd/api

# Minimal runtime
FROM gcr.io/distroless/static:nonroot
WORKDIR /

# Runtime envs
ENV API_ADDR=":5000" \
    DB_PATH="/data/data.db"

# App data dir
USER nonroot:nonroot
VOLUME ["/data"]

COPY --from=builder /out/api /api

EXPOSE 5000
ENTRYPOINT ["/api"]

