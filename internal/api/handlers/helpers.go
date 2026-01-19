package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/pratik-mahalle/infraudit/internal/api/middleware"
)

// getUserIDFromContext extracts the user ID from the request context
func getUserIDFromContext(ctx context.Context) int64 {
	// Create a minimal request to use middleware.GetUserID
	// This is a helper for when we already have a context
	if userID, ok := ctx.Value(middleware.UserIDKey).(int64); ok {
		return userID
	}
	return 0
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
