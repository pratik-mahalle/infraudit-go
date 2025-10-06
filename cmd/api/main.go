package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"infraaudit/backend/internal/auth"
	"infraaudit/backend/internal/db"
	providerspkg "infraaudit/backend/internal/providers"
	"infraaudit/backend/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type User struct {
	ID             int64   `json:"id"`
	Username       string  `json:"username"`
	Email          string  `json:"email"`
	FullName       *string `json:"fullName,omitempty"`
	Role           string  `json:"role"`
	OrganizationId *int64  `json:"organizationId,omitempty"`
	PlanType       string  `json:"planType"`
	TrialStatus    string  `json:"trialStatus"`
}

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterData struct {
	Username string  `json:"username"`
	Email    string  `json:"email"`
	Password string  `json:"password"`
	FullName *string `json:"fullName"`
	Role     *string `json:"role"`
}

// In-memory state for cloud integrations
var (
	stateMu   sync.Mutex
	providers = map[string]*CloudProviderAccount{
		"aws":   {ID: "aws", Name: "Amazon Web Services", IsConnected: false},
		"gcp":   {ID: "gcp", Name: "Google Cloud Platform", IsConnected: false},
		"azure": {ID: "azure", Name: "Microsoft Azure", IsConnected: false},
	}
	resources = []CloudResource{}
)

type CloudProviderAccount struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	IsConnected bool       `json:"isConnected"`
	LastSynced  *time.Time `json:"lastSynced,omitempty"`
	// AWS-specific simple credential storage for demo (do not use in production)
	AWSAccessKeyID     string `json:"-"`
	AWSSecretAccessKey string `json:"-"`
	AWSRegion          string `json:"region,omitempty"`
	// GCP
	GCPProjectID          string `json:"-"`
	GCPServiceAccountJSON string `json:"-"`
	GCPRegion             string `json:"-"`
	// Azure
	AzureTenantID       string `json:"-"`
	AzureClientID       string `json:"-"`
	AzureClientSecret   string `json:"-"`
	AzureSubscriptionID string `json:"-"`
	AzureLocation       string `json:"-"`
}

type CloudResource = services.CloudResource

var repo *db.DB
var slackService services.SlackService
var slackDefaultChannel string

func main() {
	r := chi.NewRouter()

	// Initialize DB (file path via env DB_PATH, default to data.db)
	dbPath := env("DB_PATH", "data.db")
	var err error
	repo, err = db.Open(dbPath)
	if err != nil {
		log.Fatalf("failed opening db: %v", err)
	}
	defer repo.Close()

	initOAuth()

	// Initialize Slack notifications
	slackURL := env("SLACK_WEBHOOK_URL", "")
	slackDefaultChannel = env("SLACK_CHANNEL", "")
	if slackURL != "" {
		slackService = &services.WebhookSlack{WebhookURL: slackURL}
		log.Printf("slack webhook configured")
	} else {
		slackService = &services.InMemorySlack{}
		log.Printf("slack webhook not configured; using noop")
	}

	frontend := env("FRONTEND_URL", "http://localhost:5173")
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{frontend},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	// Log all requests for debugging
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	// Friendly NotFound and MethodNotAllowed to reveal paths during integration
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusNotFound, map[string]any{
			"message":   "not found",
			"method":    r.Method,
			"path":      r.URL.Path,
			"requestId": r.Context().Value(middleware.RequestIDKey),
		})
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{
			"message":   "method not allowed",
			"method":    r.Method,
			"path":      r.URL.Path,
			"requestId": r.Context().Value(middleware.RequestIDKey),
		})
	})

	// Health and integrations status (structured for frontend expectations)
	r.Get("/api/status", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()
		dbOK := repo.Ping(ctx) == nil
		out := map[string]any{
			"status": map[string]any{
				"slack":  map[string]any{"configured": false},
				"openai": map[string]any{"configured": os.Getenv("OPENAI_API_KEY") != ""},
				"stripe": map[string]any{"configured": false},
				"oauth": map[string]any{
					"google": map[string]any{"configured": oauthGoogleConfig != nil},
					"github": map[string]any{"configured": oauthGitHubConfig != nil},
				},
				"db": map[string]any{"ok": dbOK},
			},
			"time": time.Now().Format(time.RFC3339),
			"ok":   true,
		}
		writeJSON(w, http.StatusOK, out)
	})
	// Alias used by some frontends
	r.Get("/api/integrations/status", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()
		dbOK := repo.Ping(ctx) == nil
		out := map[string]any{
			"status": map[string]any{
				"slack":  map[string]any{"configured": false},
				"openai": map[string]any{"configured": os.Getenv("OPENAI_API_KEY") != ""},
				"stripe": map[string]any{"configured": false},
				"oauth": map[string]any{
					"google": map[string]any{"configured": oauthGoogleConfig != nil},
					"github": map[string]any{"configured": oauthGitHubConfig != nil},
				},
				"db": map[string]any{"ok": dbOK},
			},
			"time": time.Now().Format(time.RFC3339),
			"ok":   true,
		}
		writeJSON(w, http.StatusOK, out)
	})

	// Liveness & Readiness
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := repo.Ping(ctx); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unready", "db": "down"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready", "db": "ok"})
		_, _ = w.Write([]byte("ok"))
	})

	// Auth stubs compatible with current frontend flows
	r.Post("/api/login", func(w http.ResponseWriter, r *http.Request) {
		var c Credentials
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
			return
		}
		user := demoUser(c.Email)
		_, _ = mintAndSetTokens(w, user.ID, user.Email)
		writeJSON(w, http.StatusOK, user)
	})
	// Alias to support frontend variants
	r.Post("/api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		var c Credentials
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
			return
		}
		user := demoUser(c.Email)
		_, _ = mintAndSetTokens(w, user.ID, user.Email)
		writeJSON(w, http.StatusOK, user)
	})
	// Additional aliases commonly used by frontends
	r.Post("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		var c Credentials
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
			return
		}
		user := demoUser(c.Email)
		_, _ = mintAndSetTokens(w, user.ID, user.Email)
		writeJSON(w, http.StatusOK, user)
	})
	r.Post("/api/v1/login", func(w http.ResponseWriter, r *http.Request) {
		var c Credentials
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
			return
		}
		user := demoUser(c.Email)
		_, _ = mintAndSetTokens(w, user.ID, user.Email)
		writeJSON(w, http.StatusOK, user)
	})
	r.Post("/api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		var c Credentials
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
			return
		}
		user := demoUser(c.Email)
		_, _ = mintAndSetTokens(w, user.ID, user.Email)
		writeJSON(w, http.StatusOK, user)
	})

	r.Post("/api/register", func(w http.ResponseWriter, r *http.Request) {
		var d RegisterData
		if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
			return
		}
		user := demoUser(d.Email)
		user.Username = d.Username
		_, _ = mintAndSetTokens(w, user.ID, user.Email)
		writeJSON(w, http.StatusCreated, user)
	})

	r.Post("/api/logout", func(w http.ResponseWriter, r *http.Request) {
		clearAuthCookies(w)
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.With(requireAuth).Get("/api/user", func(w http.ResponseWriter, r *http.Request) {
		c := claimsFromContext(r.Context())
		email := env("DEMO_USER_EMAIL", "demo@example.com")
		if c != nil && c.Email != "" {
			email = c.Email
		}
		writeJSON(w, http.StatusOK, demoUser(email))
	})

	r.Post("/api/start-trial", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "started"})
	})

	r.Get("/api/trial-status", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"status": "active", "daysRemaining": 7})
	})

	// OAuth: use real handlers
	r.Get("/api/auth/{provider}", func(w http.ResponseWriter, r *http.Request) {
		p := chi.URLParam(r, "provider")
		oauthLoginHandler(p)(w, r)
	})
	r.Get("/api/auth/{provider}/callback", func(w http.ResponseWriter, r *http.Request) {
		p := chi.URLParam(r, "provider")
		oauthCallbackHandler(p)(w, r)
	})
	// Aliases to support alternate oauth prefix
	r.Get("/api/oauth/{provider}", func(w http.ResponseWriter, r *http.Request) {
		p := chi.URLParam(r, "provider")
		oauthLoginHandler(p)(w, r)
	})
	r.Get("/api/oauth/{provider}/callback", func(w http.ResponseWriter, r *http.Request) {
		p := chi.URLParam(r, "provider")
		oauthCallbackHandler(p)(w, r)
	})

	// Cloud Providers API (persisted)
	r.Get("/api/cloud-providers", func(w http.ResponseWriter, r *http.Request) {
		uid := int64(1)
		rows, err := repo.GetProviderAccounts(r.Context(), uid)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "db error"})
			return
		}
		list := []*CloudProviderAccount{}
		for _, p := range rows {
			var ls *time.Time
			if p.LastSynced.Valid {
				v := p.LastSynced.Time
				ls = &v
			}
			list = append(list, &CloudProviderAccount{
				ID:                    p.Provider,
				Name:                  map[string]string{"aws": "Amazon Web Services", "gcp": "Google Cloud Platform", "azure": "Microsoft Azure"}[p.Provider],
				IsConnected:           p.IsConnected,
				LastSynced:            ls,
				AWSAccessKeyID:        p.AWSAccessKeyID,
				AWSSecretAccessKey:    p.AWSSecretAccessKey,
				AWSRegion:             p.AWSRegion,
				GCPProjectID:          p.GCPProjectID,
				GCPServiceAccountJSON: p.GCPServiceAccount,
				GCPRegion:             p.GCPRegion,
				AzureTenantID:         p.AzureTenantID,
				AzureClientID:         p.AzureClientID,
				AzureClientSecret:     p.AzureClientSecret,
				AzureSubscriptionID:   p.AzureSubscriptionID,
				AzureLocation:         p.AzureLocation,
			})
		}
		writeJSON(w, http.StatusOK, list)
	})

	// Upsert provider account (persisted)
	r.With(requirePlanAtLeast("pro"), adminOnly).Post("/api/provider-accounts/{provider}", func(w http.ResponseWriter, r *http.Request) {
		uid := int64(1)
		provider := chi.URLParam(r, "provider")
		var body struct {
			IsConnected         *bool   `json:"isConnected"`
			LastSynced          *string `json:"lastSynced"`
			AWSAccessKeyID      *string `json:"awsAccessKeyId"`
			AWSSecretAccessKey  *string `json:"awsSecretAccessKey"`
			AWSRegion           *string `json:"awsRegion"`
			GCPProjectID        *string `json:"gcpProjectId"`
			GCPServiceAccount   *string `json:"gcpServiceAccountJson"`
			GCPRegion           *string `json:"gcpRegion"`
			AzureTenantID       *string `json:"azureTenantId"`
			AzureClientID       *string `json:"azureClientId"`
			AzureClientSecret   *string `json:"azureClientSecret"`
			AzureSubscriptionID *string `json:"azureSubscriptionId"`
			AzureLocation       *string `json:"azureLocation"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		isConn := true
		if body.IsConnected != nil {
			isConn = *body.IsConnected
		}
		row := db.ProviderAccountRow{
			UserID:      uid,
			Provider:    provider,
			IsConnected: isConn,
		}
		if body.AWSAccessKeyID != nil {
			row.AWSAccessKeyID = *body.AWSAccessKeyID
		}
		if body.AWSSecretAccessKey != nil {
			row.AWSSecretAccessKey = *body.AWSSecretAccessKey
		}
		if body.AWSRegion != nil {
			row.AWSRegion = *body.AWSRegion
		}
		if body.GCPProjectID != nil {
			row.GCPProjectID = *body.GCPProjectID
		}
		if body.GCPServiceAccount != nil {
			row.GCPServiceAccount = *body.GCPServiceAccount
		}
		if body.GCPRegion != nil {
			row.GCPRegion = *body.GCPRegion
		}
		if body.AzureTenantID != nil {
			row.AzureTenantID = *body.AzureTenantID
		}
		if body.AzureClientID != nil {
			row.AzureClientID = *body.AzureClientID
		}
		if body.AzureClientSecret != nil {
			row.AzureClientSecret = *body.AzureClientSecret
		}
		if body.AzureSubscriptionID != nil {
			row.AzureSubscriptionID = *body.AzureSubscriptionID
		}
		if body.AzureLocation != nil {
			row.AzureLocation = *body.AzureLocation
		}
		if err := repo.UpsertProviderAccount(r.Context(), row); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "db error"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"status": "upserted", "provider": provider})
	})

	// Delete provider account (persisted)
	r.With(requirePlanAtLeast("pro"), adminOnly).Delete("/api/provider-accounts/{provider}", func(w http.ResponseWriter, r *http.Request) {
		uid := int64(1)
		provider := chi.URLParam(r, "provider")
		if err := repo.DeleteProviderAccount(r.Context(), uid, provider); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "db error"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
	})
	r.With(requirePlanAtLeast("pro"), adminOnly).Post("/api/cloud-providers/aws", handleConnect("aws"))
	r.With(requirePlanAtLeast("pro"), adminOnly).Post("/api/cloud-providers/gcp", handleConnect("gcp"))
	r.With(requirePlanAtLeast("pro"), adminOnly).Post("/api/cloud-providers/azure", handleConnect("azure"))
	r.With(requirePlanAtLeast("pro")).Post("/api/cloud-providers/{id}/sync", handleSyncProvider)
	r.With(requirePlanAtLeast("pro"), adminOnly).Delete("/api/cloud-providers/{id}", handleDisconnect)

	// Cloud Resources API (persisted + pagination)
	r.Get("/api/cloud-resources", func(w http.ResponseWriter, r *http.Request) {
		uid := int64(1)
		provider := r.URL.Query().Get("provider")
		pageSize := 50
		page := 1
		if v := r.URL.Query().Get("pageSize"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 500 {
				pageSize = n
			}
		}
		if v := r.URL.Query().Get("page"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				page = n
			}
		}
		offset := (page - 1) * pageSize
		rows, total, err := repo.ListResourcesPage(r.Context(), uid, provider, pageSize, offset)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "db error"})
			return
		}
		list := make([]CloudResource, 0, len(rows))
		for _, rr := range rows {
			list = append(list, CloudResource{ID: rr.ResourceID, Name: rr.Name, Type: rr.Type, Provider: rr.Provider, Region: rr.Region, Status: rr.Status})
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"items":     list,
			"resources": list, // compatibility for UIs expecting `resources`
			"total":     total,
			"page":      page,
			"pageSize":  pageSize,
			"provider":  provider,
		})
	})

	// Alias: list resources with pagination using DB
	r.Get("/api/resources", func(w http.ResponseWriter, r *http.Request) {
		uid := int64(1)
		provider := r.URL.Query().Get("provider")
		pageSize := 50
		page := 1
		if v := r.URL.Query().Get("pageSize"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 500 {
				pageSize = n
			}
		}
		if v := r.URL.Query().Get("page"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				page = n
			}
		}
		offset := (page - 1) * pageSize
		rows, total, err := repo.ListResourcesPage(r.Context(), uid, provider, pageSize, offset)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "db error"})
			return
		}
		list := make([]CloudResource, 0, len(rows))
		for _, rr := range rows {
			list = append(list, CloudResource{ID: rr.ResourceID, Name: rr.Name, Type: rr.Type, Provider: rr.Provider, Region: rr.Region, Status: rr.Status})
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"items":     list,
			"resources": list, // compatibility for UIs expecting `resources`
			"total":     total,
			"page":      page,
			"pageSize":  pageSize,
			"provider":  provider,
		})
	})

	// ---- Cost Prediction (Pro+) ----
	r.With(requirePlanAtLeast("pro")).Get("/api/cost-prediction/history", func(w http.ResponseWriter, r *http.Request) {
		uid := int64(1)
		days := 90
		if v := r.URL.Query().Get("days"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 3650 {
				days = n
			}
		}
		series, err := buildDailyCostSeries(r.Context(), uid, days)
		if err != nil {
			writeJSON(w, http.StatusPreconditionRequired, map[string]any{"message": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, series)
	})

	r.With(requirePlanAtLeast("pro")).Post("/api/cost-prediction", func(w http.ResponseWriter, r *http.Request) {
		uid := int64(1)
		model := r.URL.Query().Get("model")
		if model == "" {
			model = "linear"
		}
		days := 30
		if v := r.URL.Query().Get("days"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 3650 {
				days = n
			}
		}
		series, err := buildDailyCostSeries(r.Context(), uid, 60)
		if err != nil {
			writeJSON(w, http.StatusPreconditionRequired, map[string]any{"message": err.Error()})
			return
		}
		pred, weekly, ciLow, ciHigh := predictCostLinear(series, days)
		// Optional AI explanation
		expl := "Predicted using linear trend over last 60 days of derived daily costs from synced resources."
		ai := r.URL.Query().Get("explain")
		if ai == "1" {
			prompt := "Given a recent cost history (daily totals) and a linear-regression forecast, explain in 2-3 bullets the drivers and risks. Keep it concise and non-marketing. History points: "
			for i, p := range series {
				if i < 10 {
					prompt += p.Date + ":" + strconv.FormatFloat(p.Amount, 'f', 2, 64) + ", "
				}
			}
			resp := providerspkg.AskOpenAI(r.Context(), prompt, []string{"Forecast assumes stable inventory; seasonality not modeled.", "Confidence interval reflects variance in daily costs."})
			if len(resp) > 0 {
				expl = resp[0]
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"model":              model,
			"days":               days,
			"predictedMonthly":   pred,
			"confidenceInterval": map[string]any{"low": ciLow, "high": ciHigh},
			"weeklyBreakdown":    weekly,
			"explanation":        expl,
		})
	})

	// Enterprise-only optimization suggestions for cost prediction UI
	r.With(requirePlanAtLeast("enterprise")).Get("/api/cost-prediction/optimization-suggestions", func(w http.ResponseWriter, r *http.Request) {
		uid := int64(1)
		series, err := buildDailyCostSeries(r.Context(), uid, 30)
		if err != nil {
			writeJSON(w, http.StatusPreconditionRequired, map[string]any{"message": err.Error()})
			return
		}
		prompt := "You are a FinOps assistant. Given this short daily cost history, propose 5 concrete optimization actions with estimated monthly savings per action and effort level. Reply as concise bullet points. History: "
		for _, p := range series {
			prompt += p.Date + ":" + strconv.FormatFloat(p.Amount, 'f', 2, 64) + ", "
		}
		sugg := providerspkg.AskOpenAI(r.Context(), prompt, []string{
			"Rightsize compute instances based on CPU/memory headroom.",
			"Schedule dev/test environments to stop outside business hours.",
			"Enable lifecycle policies to transition cold object storage.",
			"Review idle load balancers and unattached volumes.",
			"Prefer committed use discounts or savings plans where applicable.",
		})
		writeJSON(w, http.StatusOK, map[string]any{"suggestions": sugg})
	})

	// Create resource (admin only)
	r.With(requireAuth, adminOnly).Post("/api/resources", func(w http.ResponseWriter, r *http.Request) {
		uid := int64(1)
		var p struct{ ID, Name, Type, Provider, Region, Status string }
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
			return
		}
		if p.ID == "" || p.Provider == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "missing id/provider"})
			return
		}
		if err := repo.UpsertResource(r.Context(), uid, db.ResourceRow{Provider: p.Provider, ResourceID: p.ID, Name: p.Name, Type: p.Type, Region: p.Region, Status: p.Status}); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "db error"})
			return
		}
		rr, ok, _ := repo.GetResourceByID(r.Context(), uid, p.ID)
		if !ok {
			writeJSON(w, http.StatusCreated, map[string]string{"status": "created"})
			return
		}
		writeJSON(w, http.StatusCreated, CloudResource{ID: rr.ResourceID, Name: rr.Name, Type: rr.Type, Provider: rr.Provider, Region: rr.Region, Status: rr.Status})
	})

	// Update resource requires auth
	r.With(requireAuth).Patch("/api/resources/{id}", func(w http.ResponseWriter, r *http.Request) {
		uid := int64(1)
		id := chi.URLParam(r, "id")
		type payload struct {
			Name   *string `json:"name"`
			Type   *string `json:"type"`
			Region *string `json:"region"`
			Status *string `json:"status"`
		}
		var p payload
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
			return
		}
		if err := repo.UpdateResource(r.Context(), uid, id, p.Name, p.Type, p.Region, p.Status); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "db error"})
			return
		}
		rr, ok, err := repo.GetResourceByID(r.Context(), uid, id)
		if err != nil || !ok {
			writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
			return
		}
		writeJSON(w, http.StatusOK, CloudResource{ID: rr.ResourceID, Name: rr.Name, Type: rr.Type, Provider: rr.Provider, Region: rr.Region, Status: rr.Status})
	})

	// Resource by id (string id from CloudResource.ID)
	r.Get("/api/resources/{id}", func(w http.ResponseWriter, r *http.Request) {
		uid := int64(1)
		id := chi.URLParam(r, "id")
		rr, ok, err := repo.GetResourceByID(r.Context(), uid, id)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "db error"})
			return
		}
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"message": "resource not found"})
			return
		}
		writeJSON(w, http.StatusOK, CloudResource{ID: rr.ResourceID, Name: rr.Name, Type: rr.Type, Provider: rr.Provider, Region: rr.Region, Status: rr.Status})
	})

	// Update resource metadata (name/type/region/status)
	r.Patch("/api/resources/{id}", func(w http.ResponseWriter, r *http.Request) {
		uid := int64(1)
		id := chi.URLParam(r, "id")
		type payload struct {
			Name   *string `json:"name"`
			Type   *string `json:"type"`
			Region *string `json:"region"`
			Status *string `json:"status"`
		}
		var p payload
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
			return
		}
		if err := repo.UpdateResource(r.Context(), uid, id, p.Name, p.Type, p.Region, p.Status); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "db error"})
			return
		}
		rr, ok, err := repo.GetResourceByID(r.Context(), uid, id)
		if err != nil || !ok {
			writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
			return
		}
		writeJSON(w, http.StatusOK, CloudResource{ID: rr.ResourceID, Name: rr.Name, Type: rr.Type, Provider: rr.Provider, Region: rr.Region, Status: rr.Status})
	})

	// Delete resource
	r.With(requireAuth).Delete("/api/resources/{id}", func(w http.ResponseWriter, r *http.Request) {
		uid := int64(1)
		id := chi.URLParam(r, "id")
		if err := repo.DeleteResource(r.Context(), uid, id); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "db error"})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	// Analytics & detections
	r.Get("/api/cost-analysis", handleCostAnalysis)
	r.Get("/api/cost-anomalies", handleCostAnomalies)
	r.Get("/api/security-drifts", handleSecurityDrifts)

	// Subscription & Billing stubs
	r.Get("/api/subscriptions/plans", handleSubscriptionPlans)
	r.Get("/api/billing/history", handleBillingHistory)
	r.Get("/api/subscriptions/status", handleSubscriptionStatus)
	r.Post("/api/subscriptions/checkout", handleCreateCheckout)

	// Scanning and dashboard
	r.Post("/api/scan/run", handleRunScan)
	r.Get("/api/scan/status", handleScanStatus)
	r.Get("/api/dashboard/overview", handleDashboardOverview)
	r.Post("/api/dashboard/refresh", handleRefreshDashboard)
	r.Post("/api/dashboard/export", handleExportDashboard)
	r.Post("/api/dashboard/widgets", handleAddWidget)
	r.Get("/api/dashboard/widgets", handleListWidgets)
	r.Delete("/api/dashboard/widgets/{id}", handleDeleteWidget)

	// Cloud provider configuration
	r.Post("/api/dashboard/configure-providers", handleConfigureProviders)

	// Dashboard filters and analysis
	r.Get("/api/dashboard/cost-analysis/periods", handleCostAnalysisPeriods)
	r.Post("/api/dashboard/anomalies/toggle", handleToggleAnomalies)
	r.Post("/api/dashboard/compare-periods", handleComparePeriods)

	// Missing dashboard endpoints
	r.Get("/api/alerts", func(w http.ResponseWriter, r *http.Request) {
		uid := int64(1)
		typ := r.URL.Query().Get("type")
		sev := r.URL.Query().Get("severity")
		rows, err := repo.ListAlerts(r.Context(), uid, typ, sev)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "db error"})
			return
		}
		writeJSON(w, http.StatusOK, rows)
	})

	// Delete alert
	r.Delete("/api/alerts/{id}", func(w http.ResponseWriter, r *http.Request) {
		uid := int64(1)
		idStr := chi.URLParam(r, "id")
		id, _ := strconv.ParseInt(idStr, 10, 64)
		if err := repo.DeleteAlert(r.Context(), uid, id); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "db error"})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	r.Get("/api/alerts/{id}", func(w http.ResponseWriter, r *http.Request) {
		uid := int64(1)
		idStr := chi.URLParam(r, "id")
		id, _ := strconv.ParseInt(idStr, 10, 64)
		row, ok, err := repo.GetAlertByID(r.Context(), uid, id)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "db error"})
			return
		}
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"message": "not found"})
			return
		}
		writeJSON(w, http.StatusOK, row)
	})
	r.Post("/api/alerts/test", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"message": "test"})
	})
	r.Get("/api/alerts/test", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"message": "test"})
	})
	r.Patch("/api/alerts/test", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"message": "test"})
	})
	r.Delete("/api/alerts/test", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"message": "test"})
	})
	r.Post("/api/alerts", func(w http.ResponseWriter, r *http.Request) {
		uid := int64(1)
		var p struct {
			Type        string `json:"type"`
			Severity    string `json:"severity"`
			Title       string `json:"title"`
			Description string `json:"description"`
			Resource    string `json:"resource"`
			Status      string `json:"status"`
			Timestamp   string `json:"timestamp"`
		}
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
			return
		}
		id, err := repo.CreateAlert(r.Context(), db.AlertRow{UserID: uid, Type: p.Type, Severity: p.Severity, Title: p.Title, Description: p.Description, Resource: p.Resource, Status: p.Status, Timestamp: p.Timestamp})
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "db error"})
			return
		}
		row, _, _ := repo.GetAlertByID(r.Context(), uid, id)
		writeJSON(w, http.StatusCreated, row)
	})
	r.Patch("/api/alerts/{id}", func(w http.ResponseWriter, r *http.Request) {
		uid := int64(1)
		idStr := chi.URLParam(r, "id")
		id, _ := strconv.ParseInt(idStr, 10, 64)
		var p struct{ Type, Severity, Title, Description, Resource, Status, Timestamp *string }
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
			return
		}
		if err := repo.UpdateAlert(r.Context(), uid, id, p.Type, p.Severity, p.Title, p.Description, p.Resource, p.Status, p.Timestamp); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "db error"})
			return
		}
		row, ok, _ := repo.GetAlertByID(r.Context(), uid, id)
		if !ok {
			writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
			return
		}
		writeJSON(w, http.StatusOK, row)
	})

	r.Get("/api/recommendations", func(w http.ResponseWriter, r *http.Request) {
		uid := int64(1)
		typ := r.URL.Query().Get("type")
		rows, err := repo.ListRecommendations(r.Context(), uid, typ)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "db error"})
			return
		}
		writeJSON(w, http.StatusOK, rows)
	})

	// Delete recommendation
	r.Delete("/api/recommendations/{id}", func(w http.ResponseWriter, r *http.Request) {
		uid := int64(1)
		idStr := chi.URLParam(r, "id")
		id, _ := strconv.ParseInt(idStr, 10, 64)
		if err := repo.DeleteRecommendation(r.Context(), uid, id); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "db error"})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	// Security drifts CRUD
	r.Get("/api/security-drifts", func(w http.ResponseWriter, r *http.Request) {
		uid := int64(1)
		resID := r.URL.Query().Get("resourceId")
		sev := r.URL.Query().Get("severity")
		rows, err := repo.ListSecurityDrifts(r.Context(), uid, resID, sev)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "db error"})
			return
		}
		writeJSON(w, http.StatusOK, rows)
	})

	// Delete security drift
	r.Delete("/api/security-drifts/{id}", func(w http.ResponseWriter, r *http.Request) {
		uid := int64(1)
		idStr := chi.URLParam(r, "id")
		id, _ := strconv.ParseInt(idStr, 10, 64)
		if err := repo.DeleteSecurityDrift(r.Context(), uid, id); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "db error"})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	r.Get("/api/security-drifts/{id}", func(w http.ResponseWriter, r *http.Request) {
		uid := int64(1)
		idStr := chi.URLParam(r, "id")
		id, _ := strconv.ParseInt(idStr, 10, 64)
		row, ok, err := repo.GetSecurityDriftByID(r.Context(), uid, id)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "db error"})
			return
		}
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"message": "not found"})
			return
		}
		writeJSON(w, http.StatusOK, row)
	})
	r.Post("/api/security-drifts", func(w http.ResponseWriter, r *http.Request) {
		uid := int64(1)
		var p struct{ ResourceID, DriftType, Severity, Details, DetectedAt, Status string }
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
			return
		}
		id, err := repo.CreateSecurityDrift(r.Context(), db.SecurityDriftRow{UserID: uid, ResourceID: p.ResourceID, DriftType: p.DriftType, Severity: p.Severity, Details: p.Details, DetectedAt: p.DetectedAt, Status: p.Status})
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "db error"})
			return
		}
		row, _, _ := repo.GetSecurityDriftByID(r.Context(), uid, id)
		writeJSON(w, http.StatusCreated, row)
	})
	r.Patch("/api/security-drifts/{id}", func(w http.ResponseWriter, r *http.Request) {
		uid := int64(1)
		idStr := chi.URLParam(r, "id")
		id, _ := strconv.ParseInt(idStr, 10, 64)
		var p struct{ ResourceID, DriftType, Severity, Details, DetectedAt, Status *string }
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
			return
		}
		if err := repo.UpdateSecurityDrift(r.Context(), uid, id, p.ResourceID, p.DriftType, p.Severity, p.Details, p.DetectedAt, p.Status); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "db error"})
			return
		}
		row, ok, _ := repo.GetSecurityDriftByID(r.Context(), uid, id)
		if !ok {
			writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
			return
		}
		writeJSON(w, http.StatusOK, row)
	})

	// Cost anomalies CRUD
	r.Get("/api/cost-anomalies", func(w http.ResponseWriter, r *http.Request) {
		uid := int64(1)
		resID := r.URL.Query().Get("resourceId")
		sev := r.URL.Query().Get("severity")
		rows, err := repo.ListCostAnomalies(r.Context(), uid, resID, sev)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "db error"})
			return
		}
		writeJSON(w, http.StatusOK, rows)
	})

	// Delete cost anomaly
	r.Delete("/api/cost-anomalies/{id}", func(w http.ResponseWriter, r *http.Request) {
		uid := int64(1)
		idStr := chi.URLParam(r, "id")
		id, _ := strconv.ParseInt(idStr, 10, 64)
		if err := repo.DeleteCostAnomaly(r.Context(), uid, id); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "db error"})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	r.Get("/api/cost-anomalies/{id}", func(w http.ResponseWriter, r *http.Request) {
		uid := int64(1)
		idStr := chi.URLParam(r, "id")
		id, _ := strconv.ParseInt(idStr, 10, 64)
		row, ok, err := repo.GetCostAnomalyByID(r.Context(), uid, id)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "db error"})
			return
		}
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"message": "not found"})
			return
		}
		writeJSON(w, http.StatusOK, row)
	})
	r.Post("/api/cost-anomalies", func(w http.ResponseWriter, r *http.Request) {
		uid := int64(1)
		var p struct {
			ResourceID, AnomalyType, Severity, DetectedAt, Status string
			Percentage, PreviousCost, CurrentCost                 int
		}
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
			return
		}
		id, err := repo.CreateCostAnomaly(r.Context(), db.CostAnomalyRow{UserID: uid, ResourceID: p.ResourceID, AnomalyType: p.AnomalyType, Severity: p.Severity, Percentage: p.Percentage, PreviousCost: p.PreviousCost, CurrentCost: p.CurrentCost, DetectedAt: p.DetectedAt, Status: p.Status})
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "db error"})
			return
		}
		row, _, _ := repo.GetCostAnomalyByID(r.Context(), uid, id)
		writeJSON(w, http.StatusCreated, row)
	})
	r.Patch("/api/cost-anomalies/{id}", func(w http.ResponseWriter, r *http.Request) {
		uid := int64(1)
		idStr := chi.URLParam(r, "id")
		id, _ := strconv.ParseInt(idStr, 10, 64)
		var p struct {
			ResourceID, AnomalyType, Severity, DetectedAt, Status *string
			Percentage, PreviousCost, CurrentCost                 *int
		}
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
			return
		}
		if err := repo.UpdateCostAnomaly(r.Context(), uid, id, p.ResourceID, p.AnomalyType, p.Severity, p.DetectedAt, p.Status, p.Percentage, p.PreviousCost, p.CurrentCost); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "db error"})
			return
		}
		row, ok, _ := repo.GetCostAnomalyByID(r.Context(), uid, id)
		if !ok {
			writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
			return
		}
		writeJSON(w, http.StatusOK, row)
	})
	r.Get("/api/recommendations/{id}", func(w http.ResponseWriter, r *http.Request) {
		uid := int64(1)
		idStr := chi.URLParam(r, "id")
		id, _ := strconv.ParseInt(idStr, 10, 64)
		row, ok, err := repo.GetRecommendationByID(r.Context(), uid, id)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "db error"})
			return
		}
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"message": "not found"})
			return
		}
		writeJSON(w, http.StatusOK, row)
	})
	r.Post("/api/recommendations", func(w http.ResponseWriter, r *http.Request) {
		uid := int64(1)
		var p struct {
			Type, Priority, Title, Description, Effort, Impact, Category, Resources string
			Savings                                                                 float64
		}
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
			return
		}
		id, err := repo.CreateRecommendation(r.Context(), db.RecommendationRow{UserID: uid, Type: p.Type, Priority: p.Priority, Title: p.Title, Description: p.Description, Savings: p.Savings, Effort: p.Effort, Impact: p.Impact, Category: p.Category, Resources: p.Resources})
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "db error"})
			return
		}
		row, _, _ := repo.GetRecommendationByID(r.Context(), uid, id)
		writeJSON(w, http.StatusCreated, row)
	})
	r.Patch("/api/recommendations/{id}", func(w http.ResponseWriter, r *http.Request) {
		uid := int64(1)
		idStr := chi.URLParam(r, "id")
		id, _ := strconv.ParseInt(idStr, 10, 64)
		var p struct {
			Type, Priority, Title, Description, Effort, Impact, Category, Resources *string
			Savings                                                                 *float64
		}
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
			return
		}
		if err := repo.UpdateRecommendation(r.Context(), uid, id, p.Type, p.Priority, p.Title, p.Description, p.Effort, p.Impact, p.Category, p.Resources, p.Savings); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "db error"})
			return
		}
		row, ok, _ := repo.GetRecommendationByID(r.Context(), uid, id)
		if !ok {
			writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
			return
		}
		writeJSON(w, http.StatusOK, row)
	})

	// AI analysis and recommendations
	r.With(requirePlanAtLeast("enterprise")).Post("/api/ai-analysis/analyze/cost/{resourceId}", handleAICostAnalysis)
	r.With(requirePlanAtLeast("enterprise")).Post("/api/ai-analysis/analyze/security/{resourceId}", handleAISecurityAnalysis)
	r.With(requirePlanAtLeast("enterprise")).Post("/api/ai-analysis/recommendations/{resourceId}", handleAIRecommendations)

	// ---- Settings minimal routes to satisfy frontend ----
	r.Get("/profile", func(w http.ResponseWriter, r *http.Request) {
		email := env("DEMO_USER_EMAIL", "demo@example.com")
		writeJSON(w, http.StatusOK, demoUser(email))
	})
	r.Put("/profile", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
	})
	r.Post("/profile/avatar", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusCreated, map[string]any{"url": "https://example.com/avatar.png"})
	})
	r.Get("/api/profile", func(w http.ResponseWriter, r *http.Request) {
		email := env("DEMO_USER_EMAIL", "demo@example.com")
		writeJSON(w, http.StatusOK, demoUser(email))
	})
	r.Put("/api/profile", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
	})
	r.Post("/api/profile/avatar", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusCreated, map[string]any{"url": "https://example.com/avatar.png"})
	})

	// Account settings
	r.Get("/account/settings", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"language": "en", "timezone": "UTC", "theme": "system"})
	})
	r.Put("/account/settings", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
	})
	r.Post("/account/password", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "changed"})
	})
	r.Get("/api/account/settings", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"language": "en", "timezone": "UTC", "theme": "system"})
	})
	r.Put("/api/account/settings", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
	})
	r.Post("/api/account/password", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "changed"})
	})

	addr := env("API_ADDR", ":8080")
	log.Printf("Go API serving on %s (CORS allow: %s)", addr, frontend)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}

func handleListProviders(w http.ResponseWriter, r *http.Request) {
	stateMu.Lock()
	defer stateMu.Unlock()
	list := []*CloudProviderAccount{}
	for _, p := range providers {
		// Make a copy to avoid race
		cp := *p
		list = append(list, &cp)
	}
	writeJSON(w, http.StatusOK, list)
}

type awsConnectRequest struct {
	AccessKeyID     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	Region          string `json:"region"`
}

type gcpConnectRequest struct {
	ProjectID          string `json:"projectId"`
	ServiceAccountJSON string `json:"serviceAccountJson"`
	Region             string `json:"region"`
}

type azureConnectRequest struct {
	TenantID       string `json:"tenantId"`
	ClientID       string `json:"clientId"`
	ClientSecret   string `json:"clientSecret"`
	SubscriptionID string `json:"subscriptionId"`
	Location       string `json:"location"`
}

func handleConnect(id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stateMu.Lock()
		defer stateMu.Unlock()
		p, ok := providers[id]
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"message": "unknown provider"})
			return
		}
		switch id {
		case "aws":
			var req awsConnectRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			// Minimal validation; allow empty to support env-based creds if desired
			p.AWSAccessKeyID = req.AccessKeyID
			p.AWSSecretAccessKey = req.SecretAccessKey
			if req.Region != "" {
				p.AWSRegion = req.Region
			}
			p.IsConnected = true
			p.LastSynced = nil
		case "gcp":
			var req gcpConnectRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			p.GCPProjectID = req.ProjectID
			p.GCPServiceAccountJSON = req.ServiceAccountJSON
			p.GCPRegion = req.Region
			p.IsConnected = true
			p.LastSynced = nil
		case "azure":
			var req azureConnectRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			p.AzureTenantID = req.TenantID
			p.AzureClientID = req.ClientID
			p.AzureClientSecret = req.ClientSecret
			p.AzureSubscriptionID = req.SubscriptionID
			p.AzureLocation = req.Location
			p.IsConnected = true
			p.LastSynced = nil
		default:
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "unsupported provider"})
			return
		}
		writeJSON(w, http.StatusOK, p)
	}
}

func handleSyncProvider(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	stateMu.Lock()
	defer stateMu.Unlock()
	p, ok := providers[id]
	if !ok || !p.IsConnected {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "provider not connected"})
		return
	}
	// Replace previous resources for this provider
	resources = filter(resources, func(cr CloudResource) bool { return cr.Provider != id })
	switch id {
	case "aws":
		awsCreds := providerspkg.AWSCredentials{AccessKeyID: p.AWSAccessKeyID, SecretAccessKey: p.AWSSecretAccessKey, Region: p.AWSRegion}
		fetched, err := providerspkg.AWSListResources(r.Context(), awsCreds)
		if err != nil {
			log.Printf("aws sync error: %v", err)
			writeJSON(w, http.StatusBadGateway, map[string]string{"message": "aws sync failed"})
			return
		}
		resources = append(resources, fetched...)
	case "gcp":
		gcpCreds := providerspkg.GCPCredentials{ProjectID: p.GCPProjectID, ServiceAccountJSON: p.GCPServiceAccountJSON, Region: p.GCPRegion}
		fetched, err := providerspkg.GCPListResources(r.Context(), gcpCreds)
		if err != nil {
			log.Printf("gcp sync error: %v", err)
			writeJSON(w, http.StatusBadGateway, map[string]string{"message": "gcp sync failed"})
			return
		}
		resources = append(resources, fetched...)
	case "azure":
		azCreds := providerspkg.AzureCredentials{TenantID: p.AzureTenantID, ClientID: p.AzureClientID, ClientSecret: p.AzureClientSecret, SubscriptionID: p.AzureSubscriptionID, Location: p.AzureLocation}
		fetched, err := providerspkg.AzureListResources(r.Context(), azCreds)
		if err != nil {
			log.Printf("azure sync error: %v", err)
			writeJSON(w, http.StatusBadGateway, map[string]string{"message": "azure sync failed"})
			return
		}
		resources = append(resources, fetched...)
	default:
		// Fallback to mock for other providers
		for i := 0; i < 5+rand.Intn(6); i++ {
			resources = append(resources, CloudResource{
				ID:       id + "-res-" + randID(),
				Name:     id + "-resource-" + randID(),
				Type:     []string{"EC2", "S3", "RDS", "GKE", "AKS"}[rand.Intn(5)],
				Provider: id,
				Region:   []string{"us-east-1", "us-west-2", "eu-west-1"}[rand.Intn(3)],
				Status:   []string{"running", "stopped"}[rand.Intn(2)],
			})
		}
	}
	now := time.Now()
	p.LastSynced = &now
	writeJSON(w, http.StatusOK, map[string]string{"status": "synced"})
}

func handleDisconnect(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	stateMu.Lock()
	defer stateMu.Unlock()
	p, ok := providers[id]
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"message": "unknown provider"})
		return
	}
	p.IsConnected = false
	p.LastSynced = nil
	resources = filter(resources, func(cr CloudResource) bool { return cr.Provider != id })
	writeJSON(w, http.StatusOK, map[string]string{"status": "disconnected"})
}

func handleListResources(w http.ResponseWriter, r *http.Request) {
	stateMu.Lock()
	defer stateMu.Unlock()
	list := make([]CloudResource, len(resources))
	copy(list, resources)
	writeJSON(w, http.StatusOK, list)
}

// ---- Analytics & Detection stubs ----
// Cost analysis returns a simple time series grouped by day
func handleCostAnalysis(w http.ResponseWriter, r *http.Request) {
	// Require at least one connected provider; otherwise instruct to connect
	uid := int64(1)
	provs, err := repo.GetProviderAccounts(r.Context(), uid)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "db error"})
		return
	}
	providersConnected := map[string]bool{"aws": false, "gcp": false, "azure": false}
	anyConnected := false
	for _, p := range provs {
		providersConnected[p.Provider] = p.IsConnected
		if p.IsConnected {
			anyConnected = true
		}
	}
	if !anyConnected {
		writeJSON(w, http.StatusPreconditionRequired, map[string]any{
			"message":            "connect a cloud provider first",
			"providersConnected": providersConnected,
		})
		return
	}

	// Period and filters
	type point struct {
		Date    string  `json:"date"`
		Amount  float64 `json:"amount"`
		Service string  `json:"service"`
		Region  string  `json:"region"`
	}
	q := r.URL.Query()
	period := q.Get("period")
	days := 30
	switch period {
	case "7d":
		days = 7
	case "30d", "":
		days = 30
	case "90d":
		days = 90
	case "1y":
		days = 365
	}
	if v := q.Get("days"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 3650 {
			days = n
		}
	}
	filterService := q.Get("service")
	filterRegion := q.Get("region")

	rows, err := repo.ListResources(r.Context(), uid)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "db error"})
		return
	}
	if len(rows) == 0 {
		writeJSON(w, http.StatusPreconditionRequired, map[string]any{
			"message":            "no resources found - run a sync first",
			"providersConnected": providersConnected,
		})
		return
	}

	rate := map[string]float64{"EC2": 5.0, "VM": 4.5, "VMSS": 6.0, "RDS": 4.0, "EKS": 3.0, "GKE": 3.0, "AKS": 3.0, "S3": 0.05, "GCS": 0.05, "Storage": 0.08}

	end := time.Now()
	start := end.AddDate(0, 0, -days)
	var out []point
	for d := start; !d.After(end); d = d.Add(24 * time.Hour) {
		for _, rr := range rows {
			amt := rate[rr.Type]
			if amt <= 0 {
				amt = 0.02
			}
			svc := rr.Type
			reg := rr.Region
			if filterService != "" && filterService != svc {
				continue
			}
			if filterRegion != "" && filterRegion != reg {
				continue
			}
			out = append(out, point{Date: d.Format("2006-01-02"), Amount: amt, Service: svc, Region: reg})
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"period": map[string]any{"days": days}, "points": out})
}

// buildDailyCostSeries produces daily totals derived from persisted resources.
// Returns error if no providers connected or no resources.
type costPoint struct {
	Date   string
	Amount float64
}

func buildDailyCostSeries(ctx context.Context, userID int64, days int) ([]costPoint, error) {
	provs, err := repo.GetProviderAccounts(ctx, userID)
	if err != nil {
		return nil, err
	}
	any := false
	for _, p := range provs {
		if p.IsConnected {
			any = true
			break
		}
	}
	if !any {
		return nil, fmt.Errorf("connect a cloud provider first")
	}
	rows, err := repo.ListResources(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("no resources found - run a sync first")
	}
	rate := map[string]float64{"EC2": 5.0, "VM": 4.5, "VMSS": 6.0, "RDS": 4.0, "EKS": 3.0, "GKE": 3.0, "AKS": 3.0, "S3": 0.05, "GCS": 0.05, "Storage": 0.08}
	end := time.Now()
	start := end.AddDate(0, 0, -days)
	var out []costPoint
	for d := start; !d.After(end); d = d.Add(24 * time.Hour) {
		var total float64
		for _, rr := range rows {
			v := rate[rr.Type]
			if v <= 0 {
				v = 0.02
			}
			total += v
		}
		out = append(out, costPoint{Date: d.Format("2006-01-02"), Amount: total})
	}
	return out, nil
}

// predictCostLinear: simple linear regression on daily series; returns monthly sum, weekly breakdown, and naive CI.
func predictCostLinear(series []costPoint, days int) (predictedMonthly float64, weekly []map[string]any, ciLow, ciHigh float64) {
	n := len(series)
	if n == 0 {
		return 0, nil, 0, 0
	}
	// Fit y = a + b*x where x is day index
	var sumX, sumY, sumXY, sumXX float64
	for i, p := range series {
		x := float64(i)
		sumX += x
		sumY += p.Amount
		sumXY += x * p.Amount
		sumXX += x * x
	}
	denom := float64(n)*sumXX - sumX*sumX
	var a, b float64
	if denom != 0 {
		b = (float64(n)*sumXY - sumX*sumY) / denom
		a = (sumY - b*sumX) / float64(n)
	} else {
		a = sumY / float64(n)
	}
	// Forecast next days
	startIdx := float64(n)
	var forecast []float64
	for i := 0; i < days; i++ {
		x := startIdx + float64(i)
		y := a + b*x
		if y < 0 {
			y = 0
		}
		forecast = append(forecast, y)
	}
	// Aggregate to monthly sum and weekly buckets (approx 7-day weeks)
	for i, v := range forecast {
		predictedMonthly += v
		if i%7 == 0 {
			weekly = append(weekly, map[string]any{"week": len(weekly) + 1, "amount": 0.0})
		}
		weekly[len(weekly)-1]["amount"] = weekly[len(weekly)-1]["amount"].(float64) + v
	}
	// Naive CI: +/-10% of predicted
	ciLow = predictedMonthly * 0.9
	ciHigh = predictedMonthly * 1.1
	return
}

// Cost anomalies flags spikes over a threshold
func handleCostAnomalies(w http.ResponseWriter, r *http.Request) {
	type anomaly struct {
		ID           string `json:"id"`
		ResourceID   string `json:"resourceId"`
		AnomalyType  string `json:"anomalyType"`
		Severity     string `json:"severity"`
		Percentage   int    `json:"percentage"`
		PreviousCost int    `json:"previousCost"`
		CurrentCost  int    `json:"currentCost"`
		DetectedAt   string `json:"detectedAt"`
		Status       string `json:"status"`
	}
	// Optional filters: provider, severity
	q := r.URL.Query()
	filterProvider := q.Get("provider")
	filterSeverity := q.Get("severity")

	stateMu.Lock()
	defer stateMu.Unlock()
	out := []anomaly{}
	for _, r := range resources {
		if filterProvider != "" && r.Provider != filterProvider {
			continue
		}
		// Randomly assign some anomalies
		if rand.Intn(6) == 0 {
			prev := 100 + rand.Intn(200)
			curr := int(float64(prev) * (1.5 + rand.Float64()))
			pct := int(float64(curr-prev) * 100 / float64(prev))
			sev := "medium"
			if pct > 120 {
				sev = "high"
			}
			if filterSeverity != "" && filterSeverity != sev {
				continue
			}
			out = append(out, anomaly{
				ID: "an-" + randID(), ResourceID: r.ID, AnomalyType: "spike", Severity: sev,
				Percentage: pct, PreviousCost: prev, CurrentCost: curr,
				DetectedAt: time.Now().Format(time.RFC3339), Status: "open",
			})
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items":    out,
		"total":    len(out),
		"provider": filterProvider,
	})
}

// Security drifts flags config changes or risky settings
func handleSecurityDrifts(w http.ResponseWriter, r *http.Request) {
	type drift struct {
		ID         string         `json:"id"`
		ResourceID string         `json:"resourceId"`
		DriftType  string         `json:"driftType"`
		Severity   string         `json:"severity"`
		Details    map[string]any `json:"details"`
		DetectedAt string         `json:"detectedAt"`
		Status     string         `json:"status"`
	}
	stateMu.Lock()
	defer stateMu.Unlock()
	out := []drift{}
	for _, r := range resources {
		if rand.Intn(8) == 0 {
			kinds := []string{"IAM policy change", "Security group open to world", "Public S3 bucket"}
			sev := []string{"low", "medium", "high"}[rand.Intn(3)]
			out = append(out, drift{
				ID: "dr-" + randID(), ResourceID: r.ID, DriftType: kinds[rand.Intn(len(kinds))],
				Severity: sev, Details: map[string]any{"resource": r.Name, "region": r.Region},
				DetectedAt: time.Now().Format(time.RFC3339), Status: "open",
			})
		}
	}
	writeJSON(w, http.StatusOK, out)
}

// ---- Subscriptions & Billing stubs ----

type SubscriptionPlan struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	PriceMonthly int      `json:"priceMonthly"`
	Currency     string   `json:"currency"`
	Description  string   `json:"description"`
	Features     []string `json:"features"`
	TrialDays    int      `json:"trialDays"`
	MostPopular  bool     `json:"mostPopular"`
	// Compatibility fields for frontends expecting simple price fields
	Price       int    `json:"price"`    // same as monthly price in major units
	Interval    string `json:"interval"` // e.g. "month"
	YearlyPrice int    `json:"yearlyPrice"`
}

func handleSubscriptionPlans(w http.ResponseWriter, r *http.Request) {
	plans := []SubscriptionPlan{
		{
			ID:           "free",
			Name:         "Free",
			PriceMonthly: 0,
			Currency:     "USD",
			Description:  "Get started with core features",
			Features:     []string{"Up to 1 cloud account", "Basic cost charts", "Weekly reports"},
			TrialDays:    0,
			MostPopular:  false,
			Price:        0,
			Interval:     "month",
			YearlyPrice:  0,
		},
		{
			ID:           "pro",
			Name:         "Pro",
			PriceMonthly: 49,
			Currency:     "USD",
			Description:  "Advanced monitoring for teams",
			Features:     []string{"Unlimited accounts", "Anomaly detection", "Slack alerts"},
			TrialDays:    14,
			MostPopular:  true,
			Price:        49,
			Interval:     "month",
			YearlyPrice:  490,
		},
		{
			ID:           "enterprise",
			Name:         "Enterprise",
			PriceMonthly: 199,
			Currency:     "USD",
			Description:  "Security & scale for large orgs",
			Features:     []string{"SSO/SAML", "RBAC", "Premium support"},
			TrialDays:    30,
			MostPopular:  false,
			Price:        199,
			Interval:     "month",
			YearlyPrice:  1990,
		},
	}
	writeJSON(w, http.StatusOK, plans)
}

type Invoice struct {
	ID          string `json:"id"`
	Date        string `json:"date"`
	Amount      int    `json:"amount"`
	Currency    string `json:"currency"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

func handleBillingHistory(w http.ResponseWriter, r *http.Request) {
	// Generate a small synthetic invoice history
	now := time.Now()
	invoices := []Invoice{
		{ID: "inv-" + randID(), Date: now.AddDate(0, -2, 0).Format("2006-01-02"), Amount: 49, Currency: "USD", Status: "paid", Description: "Pro plan - 2 months ago"},
		{ID: "inv-" + randID(), Date: now.AddDate(0, -1, 0).Format("2006-01-02"), Amount: 49, Currency: "USD", Status: "paid", Description: "Pro plan - last month"},
	}
	writeJSON(w, http.StatusOK, invoices)
}

// Subscription status shape kept simple for UI integration
func handleSubscriptionStatus(w http.ResponseWriter, r *http.Request) {
	// Demo: pretend user is on free plan with active trial
	status := map[string]any{
		"plan": map[string]any{
			"id":   "free",
			"name": "Free",
		},
		"status":            "active",
		"cancelAtPeriodEnd": false,
		"currentPeriodEnd":  time.Now().AddDate(0, 1, 0).Format(time.RFC3339),
		"trial": map[string]any{
			"status":        "active",
			"daysRemaining": 7,
		},
	}
	writeJSON(w, http.StatusOK, status)
}

// Create Stripe checkout session (stubbed)
func handleCreateCheckout(w http.ResponseWriter, r *http.Request) {
	type checkoutRequest struct {
		Plan string `json:"plan"`
	}
	var req checkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Plan == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid plan"})
		return
	}
	// In real implementation, look up authenticated user
	userID := int64(1)
	var stripe services.StripeService = &services.InMemoryStripe{}
	url, err := stripe.CreateCheckout(r.Context(), userID, req.Plan)
	if err != nil {
		log.Printf("stripe checkout error: %v", err)
		writeJSON(w, http.StatusBadGateway, map[string]string{"message": "checkout failed"})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"checkoutUrl": url})
}

// ---- Scanning and Dashboard ----

type ScanStatus struct {
	JobID          string         `json:"jobId"`
	Running        bool           `json:"running"`
	StartedAt      *time.Time     `json:"startedAt,omitempty"`
	CompletedAt    *time.Time     `json:"completedAt,omitempty"`
	Progress       int            `json:"progress"`
	NumResources   int            `json:"numResources"`
	ProviderCounts map[string]int `json:"providerCounts"`
	Message        string         `json:"message,omitempty"`
}

var currentScan ScanStatus

func handleRunScan(w http.ResponseWriter, r *http.Request) {
	stateMu.Lock()
	if currentScan.Running {
		out := currentScan
		stateMu.Unlock()
		writeJSON(w, http.StatusAccepted, out)
		return
	}
	job := "scan-" + randID()
	now := time.Now()
	currentScan = ScanStatus{JobID: job, Running: true, StartedAt: &now, Progress: 0, ProviderCounts: map[string]int{}}
	stateMu.Unlock()

	go func() {
		ctx := context.Background()
		var aggregated []CloudResource
		providersToScan := []string{"aws", "gcp", "azure"}
		step := 100
		if len(providersToScan) > 0 {
			step = 100 / len(providersToScan)
		}

		for _, id := range providersToScan {
			stateMu.Lock()
			p := providers[id]
			stateMu.Unlock()

			if p == nil || !p.IsConnected {
				stateMu.Lock()
				if currentScan.Progress+step <= 100 {
					currentScan.Progress += step
				}
				stateMu.Unlock()
				continue
			}

			switch id {
			case "aws":
				awsCreds := providerspkg.AWSCredentials{AccessKeyID: p.AWSAccessKeyID, SecretAccessKey: p.AWSSecretAccessKey, Region: p.AWSRegion}
				if items, err := providerspkg.AWSListResources(ctx, awsCreds); err == nil {
					aggregated = append(aggregated, items...)
				} else {
					log.Printf("scan aws error: %v", err)
				}
			case "gcp":
				gcpCreds := providerspkg.GCPCredentials{ProjectID: p.GCPProjectID, ServiceAccountJSON: p.GCPServiceAccountJSON, Region: p.GCPRegion}
				if items, err := providerspkg.GCPListResources(ctx, gcpCreds); err == nil {
					aggregated = append(aggregated, items...)
				} else {
					log.Printf("scan gcp error: %v", err)
				}
			case "azure":
				azCreds := providerspkg.AzureCredentials{TenantID: p.AzureTenantID, ClientID: p.AzureClientID, ClientSecret: p.AzureClientSecret, SubscriptionID: p.AzureSubscriptionID, Location: p.AzureLocation}
				if items, err := providerspkg.AzureListResources(ctx, azCreds); err == nil {
					aggregated = append(aggregated, items...)
				} else {
					log.Printf("scan azure error: %v", err)
				}
			}

			stateMu.Lock()
			if currentScan.Progress+step <= 100 {
				currentScan.Progress += step
			}
			stateMu.Unlock()
		}

		stateMu.Lock()
		resources = aggregated
		counts := map[string]int{}
		for _, cr := range resources {
			counts[cr.Provider]++
		}
		nowDone := time.Now()
		currentScan.ProviderCounts = counts
		currentScan.NumResources = len(resources)
		currentScan.CompletedAt = &nowDone
		currentScan.Running = false
		currentScan.Progress = 100
		stateMu.Unlock()
	}()

	writeJSON(w, http.StatusAccepted, map[string]any{"jobId": job, "status": "started"})
}

func handleScanStatus(w http.ResponseWriter, r *http.Request) {
	stateMu.Lock()
	defer stateMu.Unlock()
	writeJSON(w, http.StatusOK, currentScan)
}

func handleDashboardOverview(w http.ResponseWriter, r *http.Request) {
	stateMu.Lock()
	defer stateMu.Unlock()
	providerCounts := map[string]int{}
	typeCounts := map[string]int{}
	regionCounts := map[string]int{}
	for _, cr := range resources {
		providerCounts[cr.Provider]++
		typeCounts[cr.Type]++
		regionCounts[cr.Region]++
	}
	// Include resources with a simple synthetic cost for frontend aggregations
	resourceRates := map[string]float64{"EC2": 5.0, "VM": 4.5, "VMSS": 6.0, "RDS": 4.0, "EKS": 3.0, "GKE": 3.0, "AKS": 3.0, "S3": 0.05, "GCS": 0.05, "Storage": 0.08}
	resWithCost := make([]map[string]any, 0, len(resources))
	for _, cr := range resources {
		cost := resourceRates[cr.Type]
		if cost <= 0 {
			cost = 0.02
		}
		resWithCost = append(resWithCost, map[string]any{
			"id":       cr.ID,
			"name":     cr.Name,
			"type":     cr.Type,
			"provider": cr.Provider,
			"region":   cr.Region,
			"status":   cr.Status,
			"cost":     cost,
		})
	}
	out := map[string]any{
		"resourcesTotal": len(resources),
		"providersConnected": map[string]bool{
			"aws":   providers["aws"].IsConnected,
			"gcp":   providers["gcp"].IsConnected,
			"azure": providers["azure"].IsConnected,
		},
		"providerCounts": providerCounts,
		"typeCounts":     typeCounts,
		"regionCounts":   regionCounts,
		"lastScan":       currentScan,
		"resources":      resWithCost,
	}
	writeJSON(w, http.StatusOK, out)
}

// Dashboard management endpoints
func handleRefreshDashboard(w http.ResponseWriter, r *http.Request) {
	// Trigger a fresh scan and return updated overview
	stateMu.Lock()
	if currentScan.Running {
		stateMu.Unlock()
		writeJSON(w, http.StatusConflict, map[string]string{"message": "scan already running"})
		return
	}
	stateMu.Unlock()

	// Start a new scan
	handleRunScan(w, r)
}

func handleExportDashboard(w http.ResponseWriter, r *http.Request) {
	stateMu.Lock()
	defer stateMu.Unlock()

	// Check if PDF format is requested
	format := r.URL.Query().Get("format")

	// Build resources with a synthetic cost for compatibility with UIs that sum resource.cost
	resourceRates := map[string]float64{"EC2": 5.0, "VM": 4.5, "VMSS": 6.0, "RDS": 4.0, "EKS": 3.0, "GKE": 3.0, "AKS": 3.0, "S3": 0.05, "GCS": 0.05, "Storage": 0.08}
	resWithCost := make([]map[string]any, 0, len(resources))
	for _, cr := range resources {
		cost := resourceRates[cr.Type]
		if cost <= 0 {
			cost = 0.02
		}
		resWithCost = append(resWithCost, map[string]any{
			"id":       cr.ID,
			"name":     cr.Name,
			"type":     cr.Type,
			"provider": cr.Provider,
			"region":   cr.Region,
			"status":   cr.Status,
			"cost":     cost,
		})
	}

	export := map[string]any{
		"timestamp": time.Now().Format(time.RFC3339),
		"resources": resWithCost,
		"providers": map[string]bool{
			"aws":   providers["aws"].IsConnected,
			"gcp":   providers["gcp"].IsConnected,
			"azure": providers["azure"].IsConnected,
		},
		"lastScan": currentScan,
		"summary": map[string]any{
			"totalResources": len(resources),
			"providerCounts": func() map[string]int {
				counts := map[string]int{}
				for _, r := range resources {
					counts[r.Provider]++
				}
				return counts
			}(),
		},
	}

	if format == "pdf" {
		// For PDF export, set appropriate headers and return a simple PDF indicator
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", "attachment; filename=dashboard-export.pdf")

		// In a real implementation, you'd generate actual PDF content
		// For now, return a mock PDF response
		pdfContent := "PDF export functionality - Dashboard data exported on " + time.Now().Format("2006-01-02 15:04:05")
		w.Write([]byte(pdfContent))
		return
	}

	// Default JSON export
	w.Header().Set("Content-Disposition", "attachment; filename=dashboard-export.json")
	writeJSON(w, http.StatusOK, export)
}

type Widget struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Title    string         `json:"title"`
	Config   map[string]any `json:"config"`
	Position map[string]int `json:"position"`
}

var dashboardWidgets = []Widget{}

func handleAddWidget(w http.ResponseWriter, r *http.Request) {
	var widget Widget
	if err := json.NewDecoder(r.Body).Decode(&widget); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid widget data"})
		return
	}

	widget.ID = "widget-" + randID()
	dashboardWidgets = append(dashboardWidgets, widget)
	writeJSON(w, http.StatusCreated, widget)
}

func handleListWidgets(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, dashboardWidgets)
}

func handleDeleteWidget(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	for i, widget := range dashboardWidgets {
		if widget.ID == id {
			dashboardWidgets = append(dashboardWidgets[:i], dashboardWidgets[i+1:]...)
			writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
			return
		}
	}
	writeJSON(w, http.StatusNotFound, map[string]string{"message": "widget not found"})
}

func handleConfigureProviders(w http.ResponseWriter, r *http.Request) {
	type providerConfig struct {
		Provider string            `json:"provider"`
		Enabled  bool              `json:"enabled"`
		Config   map[string]string `json:"config"`
	}

	var configs []providerConfig
	if err := json.NewDecoder(r.Body).Decode(&configs); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid config data"})
		return
	}

	stateMu.Lock()
	defer stateMu.Unlock()

	for _, config := range configs {
		if p, exists := providers[config.Provider]; exists {
			if !config.Enabled {
				p.IsConnected = false
			}
			// In a real implementation, you'd update credentials here
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "configured"})
}

// Additional dashboard endpoints
func handleCostAnalysisPeriods(w http.ResponseWriter, r *http.Request) {
	// Return available time periods for cost analysis
	periods := []map[string]any{
		{"id": "7d", "name": "Last 7 days", "days": 7},
		{"id": "30d", "name": "Last 30 days", "days": 30},
		{"id": "90d", "name": "Last 90 days", "days": 90},
		{"id": "1y", "name": "Last year", "days": 365},
	}
	writeJSON(w, http.StatusOK, periods)
}

func handleToggleAnomalies(w http.ResponseWriter, r *http.Request) {
	type toggleRequest struct {
		Show bool `json:"show"`
	}
	var req toggleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid request"})
		return
	}

	// In a real implementation, this would update user preferences
	writeJSON(w, http.StatusOK, map[string]any{
		"anomaliesVisible": req.Show,
		"status":           "updated",
	})
}

func handleComparePeriods(w http.ResponseWriter, r *http.Request) {
	type compareRequest struct {
		Period1 string `json:"period1"`
		Period2 string `json:"period2"`
	}
	var req compareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid request"})
		return
	}

	// Mock comparison data
	comparison := map[string]any{
		"period1": map[string]any{
			"name":      req.Period1,
			"totalCost": 1250.75,
			"topServices": []map[string]any{
				{"name": "EC2", "cost": 450.25},
				{"name": "S3", "cost": 320.50},
			},
		},
		"period2": map[string]any{
			"name":      req.Period2,
			"totalCost": 980.25,
			"topServices": []map[string]any{
				{"name": "EC2", "cost": 380.15},
				{"name": "S3", "cost": 280.10},
			},
		},
		"difference": map[string]any{
			"amount":     270.50,
			"percentage": 27.6,
			"trend":      "increase",
		},
	}

	writeJSON(w, http.StatusOK, comparison)
}

// Missing dashboard endpoints
func handleAlerts(w http.ResponseWriter, r *http.Request) {
	// Generate mock alerts data
	alerts := []map[string]any{
		{
			"id":          "alert-" + randID(),
			"type":        "cost",
			"severity":    "high",
			"title":       "High cost spike detected",
			"description": "EC2 costs increased by 150% in the last 24 hours",
			"timestamp":   time.Now().AddDate(0, 0, -1).Format(time.RFC3339),
			"status":      "active",
			"resource":    "i-" + randID(),
		},
		{
			"id":          "alert-" + randID(),
			"type":        "security",
			"severity":    "medium",
			"title":       "Security group misconfiguration",
			"description": "Security group allows unrestricted access from 0.0.0.0/0",
			"timestamp":   time.Now().AddDate(0, 0, -2).Format(time.RFC3339),
			"status":      "active",
			"resource":    "sg-" + randID(),
		},
		{
			"id":          "alert-" + randID(),
			"type":        "performance",
			"severity":    "low",
			"title":       "Underutilized resource",
			"description": "RDS instance showing low CPU utilization for 7 days",
			"timestamp":   time.Now().AddDate(0, 0, -3).Format(time.RFC3339),
			"status":      "acknowledged",
			"resource":    "db-" + randID(),
		},
	}

	writeJSON(w, http.StatusOK, alerts)
}

func handleRecommendations(w http.ResponseWriter, r *http.Request) {
	// Generate mock recommendations data
	recommendations := []map[string]any{
		{
			"id":          "rec-" + randID(),
			"type":        "cost-optimization",
			"priority":    "high",
			"title":       "Right-size EC2 instances",
			"description": "Reduce instance types for underutilized resources to save 30% on compute costs",
			"savings":     450.75,
			"effort":      "medium",
			"impact":      "high",
			"category":    "compute",
			"resources":   []string{"i-" + randID(), "i-" + randID()},
		},
		{
			"id":          "rec-" + randID(),
			"type":        "security",
			"priority":    "high",
			"title":       "Enable MFA for root accounts",
			"description": "Multi-factor authentication is not enabled for AWS root accounts",
			"savings":     0,
			"effort":      "low",
			"impact":      "high",
			"category":    "security",
			"resources":   []string{"account-root"},
		},
		{
			"id":          "rec-" + randID(),
			"type":        "cost-optimization",
			"priority":    "medium",
			"title":       "Use Reserved Instances",
			"description": "Purchase Reserved Instances for consistent workloads to save up to 72%",
			"savings":     1200.50,
			"effort":      "low",
			"impact":      "medium",
			"category":    "compute",
			"resources":   []string{"i-" + randID(), "i-" + randID(), "i-" + randID()},
		},
		{
			"id":          "rec-" + randID(),
			"type":        "performance",
			"priority":    "medium",
			"title":       "Enable auto-scaling",
			"description": "Configure auto-scaling groups to handle traffic spikes efficiently",
			"savings":     200.25,
			"effort":      "medium",
			"impact":      "medium",
			"category":    "compute",
			"resources":   []string{"asg-" + randID()},
		},
	}

	writeJSON(w, http.StatusOK, recommendations)
}

func env(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// ---- Auth helpers & middleware ----

type contextKey string

const ctxClaimsKey contextKey = "claims"

func mintAndSetTokens(w http.ResponseWriter, userID int64, email string) (auth.TokenPair, error) {
	secret := env("JWT_SECRET", "dev-secret")
	pair, err := auth.MintTokens(userID, email, secret, 15*time.Minute, 30*24*time.Hour)
	if err != nil {
		return auth.TokenPair{}, err
	}
	secure := env("COOKIE_SECURE", "0") == "1"
	// Access token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    pair.AccessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((15 * time.Minute).Seconds()),
	})
	// Refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    pair.RefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((30 * 24 * time.Hour).Seconds()),
	})
	return pair, nil
}

func clearAuthCookies(w http.ResponseWriter) {
	for _, name := range []string{"access_token", "refresh_token"} {
		http.SetCookie(w, &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Expires:  time.Unix(0, 0),
			MaxAge:   -1,
		})
	}
}

func requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secret := env("JWT_SECRET", "dev-secret")
		var token string
		if ah := r.Header.Get("Authorization"); strings.HasPrefix(strings.ToLower(ah), "bearer ") {
			token = strings.TrimSpace(ah[len("Bearer "):])
		}
		if token == "" {
			if c, err := r.Cookie("access_token"); err == nil {
				token = c.Value
			}
		}
		if token == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "missing token"})
			return
		}
		claims, err := auth.ParseClaims(token, secret)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "invalid token"})
			return
		}
		ctx := context.WithValue(r.Context(), ctxClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func claimsFromContext(ctx context.Context) *auth.Claims {
	if v := ctx.Value(ctxClaimsKey); v != nil {
		if c, ok := v.(*auth.Claims); ok {
			return c
		}
	}
	return nil
}

func requirePlanAtLeast(minPlan string) func(http.Handler) http.Handler {
	order := map[string]int{"free": 0, "pro": 1, "enterprise": 2}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Demo: read plan from env, default to pro
			plan := env("DEMO_PLAN", "pro")
			if ae := env("ADMIN_EMAIL", ""); ae != "" {
				if c := claimsFromContext(r.Context()); c != nil && c.Email == ae {
					next.ServeHTTP(w, r)
					return
				}
			}
			if order[plan] < order[minPlan] {
				writeJSON(w, http.StatusPaymentRequired, map[string]any{"message": "upgrade required", "required": minPlan, "current": plan})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func adminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ae := env("ADMIN_EMAIL", "")
		if ae == "" {
			// If not configured, allow all in demo
			next.ServeHTTP(w, r)
			return
		}
		c := claimsFromContext(r.Context())
		if c != nil && c.Email == ae {
			next.ServeHTTP(w, r)
			return
		}
		writeJSON(w, http.StatusForbidden, map[string]string{"message": "admin only"})
	})
}

func demoUser(email string) User {
	full := "Demo User"
	return User{
		ID:          1,
		Username:    "demo",
		Email:       email,
		FullName:    &full,
		Role:        "user",
		PlanType:    "free",
		TrialStatus: "active",
	}
}

func randID() string {
	return randomString(6)
}

const letters = "abcdefghijklmnopqrstuvwxyz0123456789"

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func filter[T any](in []T, keep func(T) bool) []T {
	out := make([]T, 0, len(in))
	for _, v := range in {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}

// ---- AI integration (OpenAI) ----
// Uses environment variable OPENAI_API_KEY. If not set, returns heuristic suggestions.
func handleAICostAnalysis(w http.ResponseWriter, r *http.Request) {
	rid := chi.URLParam(r, "resourceId")
	prompt := "You are a FinOps assistant. Given a cloud resource, provide 3 actionable, provider-agnostic cost optimization steps. Resource ID: " + rid
	suggestions := providerspkg.AskOpenAI(r.Context(), prompt, []string{
		"Right-size compute instances based on recent CPU/Memory usage.",
		"Schedule non-production resources to stop outside business hours.",
		"Move infrequently accessed data to cheaper storage tiers.",
	})
	writeJSON(w, http.StatusOK, map[string]any{"resourceId": rid, "suggestions": suggestions})
}

func handleAISecurityAnalysis(w http.ResponseWriter, r *http.Request) {
	rid := chi.URLParam(r, "resourceId")
	prompt := "You are a cloud security assistant. Provide 3 hardening steps for this resource focusing on least privilege, network access, and encryption. Resource ID: " + rid
	suggestions := providerspkg.AskOpenAI(r.Context(), prompt, []string{
		"Audit IAM permissions and restrict to least privilege for this resource.",
		"Ensure network access is restricted to required CIDR ranges and ports.",
		"Enforce encryption at rest and in transit; rotate keys regularly.",
	})
	writeJSON(w, http.StatusOK, map[string]any{"resourceId": rid, "suggestions": suggestions})
}

func handleAIRecommendations(w http.ResponseWriter, r *http.Request) {
	rid := chi.URLParam(r, "resourceId")
	prompt := "Summarize the top 5 prioritized actions to reduce cost and improve security posture for the given resource. Resource ID: " + rid
	suggestions := providerspkg.AskOpenAI(r.Context(), prompt, []string{
		"Consolidate underutilized instances and adopt autoscaling.",
		"Enable lifecycle policies to transition cold data to archival tiers.",
		"Tighten security groups; remove 0.0.0.0/0 where unnecessary.",
		"Enable detailed monitoring and set anomaly alerts.",
		"Use savings plans or committed use discounts where applicable.",
	})
	writeJSON(w, http.StatusOK, map[string]any{"resourceId": rid, "recommendations": suggestions})
}
