package middleware

import (
	"net/http"
	"time"

	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.written += int64(n)
	return n, err
}

// Logger returns a middleware that logs HTTP requests
func Logger(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Get request ID from context if available
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = r.Context().Value(RequestIDKey).(string)
			}

			// Process request
			next.ServeHTTP(wrapped, r)

			// Log request details
			duration := time.Since(start)
			log.WithFields(map[string]interface{}{
				"method":     r.Method,
				"path":       r.URL.Path,
				"query":      r.URL.RawQuery,
				"status":     wrapped.statusCode,
				"duration":   duration.Milliseconds(),
				"bytes":      wrapped.written,
				"ip":         r.RemoteAddr,
				"user_agent": r.UserAgent(),
				"request_id": requestID,
			}).Info("HTTP request")
		})
	}
}
