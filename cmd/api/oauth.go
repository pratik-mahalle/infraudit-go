package main

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

var (
	oauthGoogleConfig *oauth2.Config
	oauthGitHubConfig *oauth2.Config
)

func initOAuth() {
	baseURL := env("API_BASE_URL", "http://localhost:8080")
	frontend := env("FRONTEND_URL", "http://localhost:5173")
	_ = frontend

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

func oauthLoginHandler(provider string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := randomState()
		var url string
		switch provider {
		case "google":
			if oauthGoogleConfig == nil {
				http.Error(w, "google oauth not configured", 503)
				return
			}
			url = oauthGoogleConfig.AuthCodeURL(state, oauth2.AccessTypeOnline)
		case "github":
			if oauthGitHubConfig == nil {
				http.Error(w, "github oauth not configured", 503)
				return
			}
			url = oauthGitHubConfig.AuthCodeURL(state, oauth2.AccessTypeOnline)
		default:
			http.NotFound(w, r)
			return
		}
		// TODO: store state server-side/session; for demo we skip and redirect directly
		http.Redirect(w, r, url, http.StatusFound)
	}
}

func oauthCallbackHandler(provider string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
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
			http.Error(w, "oauth not configured", 503)
			return
		}
		tok, err := conf.Exchange(r.Context(), code)
		if err != nil {
			log.Println("oauth exchange error:", err)
			http.Error(w, "oauth failed", 500)
			return
		}
		_ = tok
		// In real impl: fetch userinfo; for now mint demo JWT and redirect
		_, _ = mintAndSetTokens(w, 1, "demo@example.com")
		frontendURL := env("FRONTEND_URL", "http://localhost:5173")
		http.Redirect(w, r, frontendURL+"/dashboard", http.StatusFound)
	}
}

func randomState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
