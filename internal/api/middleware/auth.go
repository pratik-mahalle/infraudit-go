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
	// AuthIDKey is the context key for Supabase auth UUID
	AuthIDKey ContextKey = "authID"
)

// UserResolver resolves a Supabase auth UUID to an internal user ID (int64)
type UserResolver func(ctx context.Context, authID string) (int64, error)

// SupabaseAuthMiddleware returns a middleware that validates Supabase JWT tokens
// and resolves the Supabase user UUID to an internal integer user ID.
// It supports both HS256 and ES256 (JWKS) token verification.
func SupabaseAuthMiddleware(kf *auth.JWKSKeyFunc, resolveUser UserResolver) func(http.Handler) http.Handler {
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

			// Parse and validate Supabase JWT (supports HS256 + ES256)
			claims, err := auth.ParseSupabaseClaims(tokenStr, kf)
			if err != nil {
				utils.WriteError(w, errors.Unauthorized("Invalid or expired token"))
				return
			}

			// Resolve Supabase UUID to internal user ID
			userID, err := resolveUser(r.Context(), claims.Sub)
			if err != nil {
				utils.WriteError(w, errors.Unauthorized("User profile not found"))
				return
			}

			// Add user info to context (int64 userID for backward compatibility)
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
			ctx = context.WithValue(ctx, AuthIDKey, claims.Sub)

			// Add audit info to logs
			AddLogField(w, "user_id", userID)
			AddLogField(w, "email", claims.Email)
			AddLogField(w, "auth_id", claims.Sub)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalSupabaseAuthMiddleware is like SupabaseAuthMiddleware but doesn't reject requests without tokens
func OptionalSupabaseAuthMiddleware(kf *auth.JWKSKeyFunc, resolveUser UserResolver) func(http.Handler) http.Handler {
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
				claims, err := auth.ParseSupabaseClaims(tokenStr, kf)
				if err == nil {
					userID, err := resolveUser(r.Context(), claims.Sub)
					if err == nil {
						ctx := context.WithValue(r.Context(), UserIDKey, userID)
						ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
						ctx = context.WithValue(ctx, AuthIDKey, claims.Sub)

						AddLogField(w, "user_id", userID)
						AddLogField(w, "email", claims.Email)
						AddLogField(w, "auth_id", claims.Sub)

						r = r.WithContext(ctx)
					}
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

// GetAuthID extracts the Supabase auth UUID from the request context
func GetAuthID(r *http.Request) (string, bool) {
	authID, ok := r.Context().Value(AuthIDKey).(string)
	return authID, ok
}
