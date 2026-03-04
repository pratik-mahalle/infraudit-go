package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pratik-mahalle/infraudit/internal/api/handlers"
	"github.com/pratik-mahalle/infraudit/internal/api/middleware"
	"github.com/pratik-mahalle/infraudit/internal/auth"
	"github.com/pratik-mahalle/infraudit/internal/domain/user"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
	"github.com/pratik-mahalle/infraudit/internal/services"
	"github.com/pratik-mahalle/infraudit/internal/testutil"
)

const testJWTSecret = "test-supabase-jwt-secret-for-testing"

// setupAuthTestRouter creates a test router for auth integration tests
// using SupabaseAuthMiddleware with the mock resolver
func setupAuthTestRouter(t *testing.T) (*chi.Mux, *testutil.MockUserRepository) {
	userRepo := testutil.NewMockUserRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})

	userService := services.NewUserService(userRepo, log)
	authHandler := handlers.NewAuthHandler(userService, log)

	r := chi.NewRouter()

	// Public routes
	r.Post("/api/auth/logout", authHandler.Logout)

	// Protected routes - use SupabaseAuthMiddleware with HS256 test key
	kf := auth.NewJWKSKeyFunc("", testJWTSecret)
	r.Group(func(r chi.Router) {
		r.Use(middleware.SupabaseAuthMiddleware(kf, userRepo.ResolveAuthID))
		r.Get("/api/auth/me", authHandler.Me)
	})

	return r, userRepo
}

// seedProfile creates a user profile in the mock repository
func seedProfile(repo *testutil.MockUserRepository, authID, email string) *user.User {
	u := &user.User{
		AuthID:   authID,
		Email:    email,
		Username: email,
		Role:     user.RoleUser,
		PlanType: user.PlanTypeFree,
	}
	_ = repo.Create(context.Background(), u)
	return u
}

// mintTestSupabaseJWT creates a test JWT matching Supabase format
func mintTestSupabaseJWT(t *testing.T, sub, email string) string {
	t.Helper()
	claims := auth.SupabaseClaims{
		Sub:   sub,
		Email: email,
		Role:  "authenticated",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(testJWTSecret))
	if err != nil {
		t.Fatalf("Failed to sign test JWT: %v", err)
	}
	return signed
}

func TestAuth_ProtectedRouteWithoutToken(t *testing.T) {
	r, _ := setupAuthTestRouter(t)
	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/auth/me")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Protected route without token should return 401, got %d", resp.StatusCode)
	}
}

func TestAuth_ProtectedRouteWithValidToken(t *testing.T) {
	r, userRepo := setupAuthTestRouter(t)
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Seed a profile in the mock repo
	authID := "test-supabase-uuid-123"
	seedProfile(userRepo, authID, "protected@example.com")

	// Mint a Supabase-style JWT for testing
	token := mintTestSupabaseJWT(t, authID, "protected@example.com")

	// Make authenticated request
	req, _ := http.NewRequest("GET", ts.URL+"/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Authenticated request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Authenticated request should return 200, got %d", resp.StatusCode)
	}
}

func TestAuth_Logout(t *testing.T) {
	r, _ := setupAuthTestRouter(t)
	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/api/auth/logout", "application/json", nil)
	if err != nil {
		t.Fatalf("Logout request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Logout should return 200, got %d", resp.StatusCode)
	}
}
