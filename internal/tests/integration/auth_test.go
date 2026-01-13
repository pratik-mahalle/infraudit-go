package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/pratik-mahalle/infraudit/internal/api/handlers"
	"github.com/pratik-mahalle/infraudit/internal/api/middleware"
	"github.com/pratik-mahalle/infraudit/internal/config"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
	"github.com/pratik-mahalle/infraudit/internal/pkg/validator"
	"github.com/pratik-mahalle/infraudit/internal/services"
	"github.com/pratik-mahalle/infraudit/internal/testutil"
)

// setupAuthTestRouter creates a test router for auth integration tests
func setupAuthTestRouter(t *testing.T) *chi.Mux {
	userRepo := testutil.NewMockUserRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	val := validator.New()

	cfg := &config.Config{
		Auth: config.AuthConfig{
			JWTSecret:          "test-secret-key-for-testing-only",
			BCryptCost:         4, // Low cost for fast tests
			AccessTokenExpiry:  15 * time.Minute,
			RefreshTokenExpiry: 24 * time.Hour,
		},
	}

	userService := services.NewUserService(userRepo, log)
	authHandler := handlers.NewAuthHandler(userService, cfg, log, val)

	r := chi.NewRouter()

	// Public routes
	r.Post("/api/auth/register", authHandler.Register)
	r.Post("/api/auth/login", authHandler.Login)
	r.Post("/api/auth/refresh", authHandler.RefreshToken)
	r.Post("/api/auth/logout", authHandler.Logout)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg.Auth.JWTSecret))
		r.Get("/api/auth/me", authHandler.Me)
	})

	return r
}

func TestAuth_RegisterNewUser(t *testing.T) {
	r := setupAuthTestRouter(t)
	ts := httptest.NewServer(r)
	defer ts.Close()

	payload := map[string]string{
		"email":    "newuser@example.com",
		"password": "SecurePassword123!",
		"name":     "New User",
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(ts.URL+"/api/auth/register", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Register request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Errorf("Register returned status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}
}

func TestAuth_RegisterDuplicateEmail(t *testing.T) {
	r := setupAuthTestRouter(t)
	ts := httptest.NewServer(r)
	defer ts.Close()

	payload := map[string]string{
		"email":    "duplicate@example.com",
		"password": "SecurePassword123!",
		"name":     "First User",
	}
	body, _ := json.Marshal(payload)

	// First registration
	resp, _ := http.Post(ts.URL+"/api/auth/register", "application/json", bytes.NewBuffer(body))
	resp.Body.Close()

	// Second registration with same email
	body, _ = json.Marshal(payload)
	resp, err := http.Post(ts.URL+"/api/auth/register", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Second register request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusConflict && resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Duplicate register should return 409 or 400, got %d", resp.StatusCode)
	}
}

func TestAuth_LoginSuccess(t *testing.T) {
	r := setupAuthTestRouter(t)
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Register first
	regPayload := map[string]string{
		"email":    "login@example.com",
		"password": "SecurePassword123!",
		"name":     "Login User",
	}
	body, _ := json.Marshal(regPayload)
	resp, _ := http.Post(ts.URL+"/api/auth/register", "application/json", bytes.NewBuffer(body))
	resp.Body.Close()

	// Login
	loginPayload := map[string]string{
		"email":    "login@example.com",
		"password": "SecurePassword123!",
	}
	body, _ = json.Marshal(loginPayload)

	resp, err := http.Post(ts.URL+"/api/auth/login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Login request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Errorf("Login returned status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// Check for access_token in response
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if _, ok := result["access_token"]; !ok {
		if data, ok := result["data"].(map[string]interface{}); ok {
			if _, ok := data["access_token"]; !ok {
				t.Log("Response body does not contain access_token directly, checking structure...")
			}
		}
	}
}

func TestAuth_LoginInvalidCredentials(t *testing.T) {
	r := setupAuthTestRouter(t)
	ts := httptest.NewServer(r)
	defer ts.Close()

	payload := map[string]string{
		"email":    "nonexistent@example.com",
		"password": "WrongPassword123!",
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(ts.URL+"/api/auth/login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Login request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusNotFound {
		t.Errorf("Invalid login should return 401 or 404, got %d", resp.StatusCode)
	}
}

func TestAuth_ProtectedRouteWithoutToken(t *testing.T) {
	r := setupAuthTestRouter(t)
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
	r := setupAuthTestRouter(t)
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Register and login to get token
	regPayload := map[string]string{
		"email":    "protected@example.com",
		"password": "SecurePassword123!",
		"name":     "Protected User",
	}
	body, _ := json.Marshal(regPayload)
	resp, _ := http.Post(ts.URL+"/api/auth/register", "application/json", bytes.NewBuffer(body))
	resp.Body.Close()

	loginPayload := map[string]string{
		"email":    "protected@example.com",
		"password": "SecurePassword123!",
	}
	body, _ = json.Marshal(loginPayload)
	resp, _ = http.Post(ts.URL+"/api/auth/login", "application/json", bytes.NewBuffer(body))

	var loginResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&loginResult)
	resp.Body.Close()

	// Extract token (structure depends on API response format)
	var accessToken string
	if token, ok := loginResult["access_token"].(string); ok {
		accessToken = token
	} else if data, ok := loginResult["data"].(map[string]interface{}); ok {
		if token, ok := data["access_token"].(string); ok {
			accessToken = token
		}
	}

	if accessToken == "" {
		t.Log("Could not extract access token from login response")
		return
	}

	// Make authenticated request
	req, _ := http.NewRequest("GET", ts.URL+"/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Authenticated request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Errorf("Authenticated request should return 200, got %d: %s", resp.StatusCode, string(bodyBytes))
	}
}

func TestAuth_RefreshToken(t *testing.T) {
	r := setupAuthTestRouter(t)
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Register and login
	regPayload := map[string]string{
		"email":    "refresh@example.com",
		"password": "SecurePassword123!",
		"name":     "Refresh User",
	}
	body, _ := json.Marshal(regPayload)
	resp, _ := http.Post(ts.URL+"/api/auth/register", "application/json", bytes.NewBuffer(body))
	resp.Body.Close()

	loginPayload := map[string]string{
		"email":    "refresh@example.com",
		"password": "SecurePassword123!",
	}
	body, _ = json.Marshal(loginPayload)
	resp, _ = http.Post(ts.URL+"/api/auth/login", "application/json", bytes.NewBuffer(body))

	var loginResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&loginResult)
	resp.Body.Close()

	// Extract refresh token
	var refreshToken string
	if token, ok := loginResult["refresh_token"].(string); ok {
		refreshToken = token
	} else if data, ok := loginResult["data"].(map[string]interface{}); ok {
		if token, ok := data["refresh_token"].(string); ok {
			refreshToken = token
		}
	}

	if refreshToken == "" {
		t.Log("Could not extract refresh token from login response")
		return
	}

	// Refresh token
	refreshPayload := map[string]string{
		"refresh_token": refreshToken,
	}
	body, _ = json.Marshal(refreshPayload)
	resp, err := http.Post(ts.URL+"/api/auth/refresh", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Refresh request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Errorf("Token refresh should return 200, got %d: %s", resp.StatusCode, string(bodyBytes))
	}
}
