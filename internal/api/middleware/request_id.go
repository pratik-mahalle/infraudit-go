package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const (
	// RequestIDKey is the context key for request ID
	RequestIDKey ContextKey = "requestID"
	// RequestIDHeader is the HTTP header for request ID
	RequestIDHeader = "X-Request-ID"
)

// RequestID returns a middleware that adds a request ID to each request
func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if request ID already exists in header
			requestID := r.Header.Get(RequestIDHeader)
			if requestID == "" {
				// Generate new UUID
				requestID = uuid.New().String()
			}

			// Add to context
			ctx := context.WithValue(r.Context(), RequestIDKey, requestID)

			// Add to response header
			w.Header().Set(RequestIDHeader, requestID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetRequestID extracts the request ID from the request context
func GetRequestID(r *http.Request) string {
	if requestID, ok := r.Context().Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}
