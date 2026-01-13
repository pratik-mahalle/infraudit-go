package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"

	"github.com/golang-jwt/jwt/v5"
)

var (
	oauthGoogleConfig *oauth2.Config
	oauthGitHubConfig *oauth2.Config
	jwtSecret         = []byte(env("JWT_SECRET", "supersecretkey"))
)

// Initialize OAuth configs
func initOAuth() {
	baseURL := env("API_BASE_URL", "http://localhost:8080")

	if cid := os.Getenv("GOOGLE_CLIENT_ID"); cid != "" {
		oauthGoogleConfig = &oauth2.Config{
			ClientID:     cid,
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  baseURL + "/api/auth/google/callback",
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		}
	}

	if cid := os.Getenv("GITHUB_CLIENT_ID"); cid != "" {
		oauthGitHubConfig = &oauth2.Config{
			ClientID:     cid,
			ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
			RedirectURL:  baseURL + "/api/auth/github/callback",
			Scopes:       []string{"read:user", "user:email"},
			Endpoint:     github.Endpoint,
		}
	}
}

// Generates a random state string for OAuth
func randomState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// Login handler
func oauthLoginHandler(provider string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := randomState()

		// Store state in cookie for CSRF protection
		http.SetCookie(w, &http.Cookie{
			Name:     "oauth_state",
			Value:    state,
			Path:     "/",
			HttpOnly: true,
			Secure:   false, // Set true in production with HTTPS
			MaxAge:   300,   // 5 minutes
		})

		var url string
		switch provider {
		case "google":
			if oauthGoogleConfig == nil {
				http.Error(w, "google oauth not configured", http.StatusServiceUnavailable)
				return
			}
			url = oauthGoogleConfig.AuthCodeURL(state, oauth2.AccessTypeOnline)
		case "github":
			if oauthGitHubConfig == nil {
				http.Error(w, "github oauth not configured", http.StatusServiceUnavailable)
				return
			}
			url = oauthGitHubConfig.AuthCodeURL(state, oauth2.AccessTypeOnline)
		default:
			http.NotFound(w, r)
			return
		}

		http.Redirect(w, r, url, http.StatusFound)
	}
}

// Callback handler
func oauthCallbackHandler(provider string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			return
		}

		// Validate state
		cookie, err := r.Cookie("oauth_state")
		if err != nil || r.URL.Query().Get("state") != cookie.Value {
			http.Error(w, "invalid state", http.StatusBadRequest)
			return
		}

		var conf *oauth2.Config
		switch provider {
		case "google":
			conf = oauthGoogleConfig
		case "github":
			conf = oauthGitHubConfig
		}
		if conf == nil {
			http.Error(w, "oauth not configured", http.StatusServiceUnavailable)
			return
		}

		tok, err := conf.Exchange(r.Context(), code)
		if err != nil {
			log.Println("oauth exchange error:", err)
			http.Error(w, "oauth failed", http.StatusInternalServerError)
			return
		}

		// Fetch user info
		userEmail, err := fetchUserEmail(provider, tok)
		if err != nil {
			log.Println("failed to fetch user info:", err)
			http.Error(w, "failed to fetch user info", http.StatusInternalServerError)
			return
		}

		// Mint JWT
		tokenString, err := mintJWT(userEmail)
		if err != nil {
			log.Println("failed to mint JWT:", err)
			http.Error(w, "failed to generate token", http.StatusInternalServerError)
			return
		}

		// Set JWT in cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "auth_token",
			Value:    tokenString,
			Path:     "/",
			HttpOnly: true,
			Secure:   false, // set true in production
			MaxAge:   3600,  // 1 hour
		})

		frontendURL := env("FRONTEND_URL", "http://localhost:5173")
		http.Redirect(w, r, frontendURL+"/dashboard", http.StatusFound)
	}
}

// Fetch user email
func fetchUserEmail(provider string, tok *oauth2.Token) (string, error) {
	client := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(tok))

	switch provider {
	case "google":
		resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		var data struct {
			Email string `json:"email"`
		}
		err = json.NewDecoder(resp.Body).Decode(&data)
		return data.Email, err

	case "github":
		// Fetch primary email
		resp, err := client.Get("https://api.github.com/user/emails")
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		var emails []struct {
			Email    string `json:"email"`
			Primary  bool   `json:"primary"`
			Verified bool   `json:"verified"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
			return "", err
		}
		for _, e := range emails {
			if e.Primary && e.Verified {
				return e.Email, nil
			}
		}
		return "", fmt.Errorf("no verified primary email found")

	default:
		return "", fmt.Errorf("unsupported provider")
	}
}

// Mint JWT
func mintJWT(email string) (string, error) {
	claims := jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(time.Hour * 1).Unix(),
		"iat":   time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// Helper: get env variable with default
// uses env from main.go
