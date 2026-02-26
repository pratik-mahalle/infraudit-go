package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/pratik-mahalle/infraudit/internal/api/middleware"
)

// getUserIDFromContext extracts the user ID from the request context.
// When auth is disabled (no JWT in context), falls back to user ID 1
// so all API endpoints remain functional during development.
func getUserIDFromContext(ctx context.Context) int64 {
	if userID, ok := ctx.Value(middleware.UserIDKey).(int64); ok {
		return userID
	}
	// AUTH DISABLED FALLBACK: return default dev user so endpoints work without login.
	// To re-enable strict auth, change this back to `return 0`.
	return 1
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// respondError sends an error response
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]interface{}{
		"success": false,
		"error":   message,
	})
}
