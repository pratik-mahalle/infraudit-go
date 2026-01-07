package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"github.com/pratik-mahalle/infraudit/internal/pkg/errors"
	"github.com/pratik-mahalle/infraudit/internal/pkg/utils"
)

// RateLimiter represents a rate limiter for HTTP requests
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerSecond float64, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rate.Limit(requestsPerSecond),
		burst:    burst,
	}
}

// getLimiter returns a rate limiter for the given key (usually IP address)
func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[key] = limiter
	}

	return limiter
}

// Cleanup removes old rate limiters (should be called periodically)
func (rl *RateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Remove limiters that haven't been used recently
	for key, limiter := range rl.limiters {
		if limiter.Tokens() == float64(rl.burst) {
			delete(rl.limiters, key)
		}
	}
}

// RateLimit returns a middleware that rate limits requests by IP
func RateLimit(requestsPerSecond float64, burst int) func(http.Handler) http.Handler {
	limiter := NewRateLimiter(requestsPerSecond, burst)

	// Start cleanup goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			limiter.Cleanup()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Use IP address as key
			key := r.RemoteAddr

			// Check if request is allowed
			if !limiter.getLimiter(key).Allow() {
				utils.WriteError(w, errors.RateLimited("Too many requests. Please try again later."))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// UserRateLimit returns a middleware that rate limits requests per user
func UserRateLimit(requestsPerSecond float64, burst int) func(http.Handler) http.Handler {
	limiter := NewRateLimiter(requestsPerSecond, burst)

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			limiter.Cleanup()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := GetUserID(r)
			if !ok {
				// If no user ID, fall back to IP-based limiting
				key := r.RemoteAddr
				if !limiter.getLimiter(key).Allow() {
					utils.WriteError(w, errors.RateLimited("Too many requests. Please try again later."))
					return
				}
			} else {
				key := fmt.Sprintf("user:%d", userID)
				if !limiter.getLimiter(key).Allow() {
					utils.WriteError(w, errors.RateLimited("Too many requests. Please try again later."))
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
