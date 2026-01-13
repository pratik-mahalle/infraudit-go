package middleware

import (
	"net/http"
	"strings"

	"github.com/go-chi/cors"
)

// CORS returns a CORS middleware with the given allowed origins
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
			"X-Request-ID",
		},
		ExposedHeaders: []string{
			"X-Request-ID",
		},
		AllowCredentials: true,
		MaxAge:           300, // 5 minutes
	})
}

// DefaultCORS returns a CORS middleware with default settings
func DefaultCORS(frontendURL string) func(http.Handler) http.Handler {
	allowedOrigins := []string{frontendURL}

	// Add localhost origins for development
	if strings.Contains(frontendURL, "localhost") || strings.Contains(frontendURL, "127.0.0.1") {
		allowedOrigins = append(allowedOrigins,
			"http://localhost:3000",
			"http://localhost:5173",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:5173",
		)
	}

	return CORS(allowedOrigins)
}
