package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/pratik-mahalle/infraudit/internal/auth"
	"github.com/pratik-mahalle/infraudit/internal/pkg/errors"
	"github.com/pratik-mahalle/infraudit/internal/pkg/utils"
)

// ContextKey is a custom type for context keys
type ContextKey string

const (
	// UserIDKey is the context key for user ID
	UserIDKey ContextKey = "userID"
	// UserEmailKey is the context key for user email
	UserEmailKey ContextKey = "email"
)

// AuthMiddleware returns a middleware that validates JWT tokens
func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try to get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			var tokenStr string

			if authHeader != "" {
				// Bearer token
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					tokenStr = parts[1]
				}
			} else {
				// Try to get from cookie
				cookie, err := r.Cookie("accessToken")
				if err == nil {
					tokenStr = cookie.Value
				}
			}

			if tokenStr == "" {
				utils.WriteError(w, errors.Unauthorized("Missing authentication token"))
				return
			}

			// Parse and validate token
			claims, err := auth.ParseClaims(tokenStr, jwtSecret)
			if err != nil {
				utils.WriteError(w, errors.Unauthorized("Invalid or expired token"))
				return
			}

			// Add user info to context
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)

			// Add audit info to logs
			AddLogField(w, "user_id", claims.UserID)
			AddLogField(w, "email", claims.Email)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuthMiddleware is like AuthMiddleware but doesn't reject requests without tokens
func OptionalAuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try to get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			var tokenStr string

			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					tokenStr = parts[1]
				}
			} else {
				cookie, err := r.Cookie("accessToken")
				if err == nil {
					tokenStr = cookie.Value
				}
			}

			if tokenStr != "" {
				claims, err := auth.ParseClaims(tokenStr, jwtSecret)
				if err == nil {
					ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
					ctx = context.WithValue(ctx, UserEmailKey, claims.Email)

					// Add audit info to logs
					AddLogField(w, "user_id", claims.UserID)
					AddLogField(w, "email", claims.Email)

					r = r.WithContext(ctx)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserID extracts the user ID from the request context
func GetUserID(r *http.Request) (int64, bool) {
	userID, ok := r.Context().Value(UserIDKey).(int64)
	return userID, ok
}

// GetUserEmail extracts the user email from the request context
func GetUserEmail(r *http.Request) (string, bool) {
	email, ok := r.Context().Value(UserEmailKey).(string)
	return email, ok
}
