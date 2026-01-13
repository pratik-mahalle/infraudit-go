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
	fields     map[string]interface{}
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

// AddLogField adds a field to the request log
func AddLogField(w http.ResponseWriter, key string, value interface{}) {
	if rw, ok := w.(*responseWriter); ok {
		rw.fields[key] = value
	}
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
				fields:         make(map[string]interface{}),
			}

			// Get request ID from context if available
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				if val := r.Context().Value(RequestIDKey); val != nil {
					requestID = val.(string)
				}
			}

			// Process request
			next.ServeHTTP(wrapped, r)

			// Log request details
			duration := time.Since(start)

			fields := map[string]interface{}{
				"method":     r.Method,
				"path":       r.URL.Path,
				"query":      r.URL.RawQuery,
				"status":     wrapped.statusCode,
				"duration":   duration.Milliseconds(),
				"bytes":      wrapped.written,
				"ip":         r.RemoteAddr,
				"user_agent": r.UserAgent(),
				"request_id": requestID,
			}

			// Merge custom fields
			for k, v := range wrapped.fields {
				fields[k] = v
			}

			log.WithFields(fields).Info("HTTP request")
		})
	}
}
