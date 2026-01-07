package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/pratik-mahalle/infraudit/internal/pkg/errors"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
	"github.com/pratik-mahalle/infraudit/internal/pkg/utils"
)

// Recovery returns a middleware that recovers from panics
func Recovery(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log the panic with stack trace
					stack := debug.Stack()
					log.WithFields(map[string]interface{}{
						"error":      err,
						"stack":      string(stack),
						"method":     r.Method,
						"path":       r.URL.Path,
						"request_id": GetRequestID(r),
					}).Error("Panic recovered")

					// Return 500 error to client
					appErr := errors.Internal(
						fmt.Sprintf("Internal server error: %v", err),
						fmt.Errorf("panic: %v", err),
					)
					utils.WriteError(w, appErr)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
