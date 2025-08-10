package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

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

type CloudResource struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Provider string `json:"provider"`
	Region   string `json:"region"`
	Status   string `json:"status"`
}

func main() {
	r := chi.NewRouter()

	initOAuth()

	frontend := env("FRONTEND_URL", "http://localhost:5173")
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{frontend},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	// Log all requests for debugging
	r.Use(middleware.Logger)
	// Friendly NotFound and MethodNotAllowed to reveal paths during integration
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("404 %s %s", r.Method, r.URL.Path)
		writeJSON(w, http.StatusNotFound, map[string]string{"message": "not found"})
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("405 %s %s", r.Method, r.URL.Path)
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"message": "method not allowed"})
	})

	// Health
	r.Get("/api/status", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"status": "OK"})
	})

	// Auth stubs compatible with current frontend flows
	r.Post("/api/login", func(w http.ResponseWriter, r *http.Request) {
		var c Credentials
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
			return
		}
		user := demoUser(c.Email)
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
		writeJSON(w, http.StatusOK, user)
	})
	r.Post("/api/v1/login", func(w http.ResponseWriter, r *http.Request) {
		var c Credentials
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
			return
		}
		user := demoUser(c.Email)
		writeJSON(w, http.StatusOK, user)
	})
	r.Post("/api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		var c Credentials
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
			return
		}
		user := demoUser(c.Email)
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
		writeJSON(w, http.StatusCreated, user)
	})

	r.Post("/api/logout", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Get("/api/user", func(w http.ResponseWriter, r *http.Request) {
		// In real impl, verify JWT/cookie and fetch user
		email := env("DEMO_USER_EMAIL", "demo@example.com")
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

	// Cloud Providers API
	r.Get("/api/cloud-providers", handleListProviders)
	r.Post("/api/cloud-providers/aws", handleConnect("aws"))
	r.Post("/api/cloud-providers/gcp", handleConnect("gcp"))
	r.Post("/api/cloud-providers/azure", handleConnect("azure"))
	r.Post("/api/cloud-providers/{id}/sync", handleSyncProvider)
	r.Delete("/api/cloud-providers/{id}", handleDisconnect)

	// Cloud Resources API
	r.Get("/api/cloud-resources", handleListResources)

	// Analytics & detections
	r.Get("/api/cost-analysis", handleCostAnalysis)
	r.Get("/api/cost-anomalies", handleCostAnomalies)
	r.Get("/api/security-drifts", handleSecurityDrifts)

	// AI analysis and recommendations
	r.Post("/api/ai-analysis/analyze/cost/{resourceId}", handleAICostAnalysis)
	r.Post("/api/ai-analysis/analyze/security/{resourceId}", handleAISecurityAnalysis)
	r.Post("/api/ai-analysis/recommendations/{resourceId}", handleAIRecommendations)

	addr := env("API_ADDR", ":5000")
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
		awsCreds := AWSCredentials{AccessKeyID: p.AWSAccessKeyID, SecretAccessKey: p.AWSSecretAccessKey, Region: p.AWSRegion}
		fetched, err := awsListResources(r.Context(), awsCreds)
		if err != nil {
			log.Printf("aws sync error: %v", err)
			writeJSON(w, http.StatusBadGateway, map[string]string{"message": "aws sync failed"})
			return
		}
		resources = append(resources, fetched...)
	case "gcp":
		gcpCreds := GCPCredentials{ProjectID: p.GCPProjectID, ServiceAccountJSON: p.GCPServiceAccountJSON, Region: p.GCPRegion}
		fetched, err := gcpListResources(r.Context(), gcpCreds)
		if err != nil {
			log.Printf("gcp sync error: %v", err)
			writeJSON(w, http.StatusBadGateway, map[string]string{"message": "gcp sync failed"})
			return
		}
		resources = append(resources, fetched...)
	case "azure":
		azCreds := AzureCredentials{TenantID: p.AzureTenantID, ClientID: p.AzureClientID, ClientSecret: p.AzureClientSecret, SubscriptionID: p.AzureSubscriptionID, Location: p.AzureLocation}
		fetched, err := azureListResources(r.Context(), azCreds)
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
	// Generate 30-day mock cost data with mild variance
	type point struct {
		Date    string  `json:"date"`
		Amount  float64 `json:"amount"`
		Service string  `json:"service"`
		Region  string  `json:"region"`
	}
	end := time.Now()
	start := end.AddDate(0, 0, -30)
	var out []point
	services := []string{"EC2", "S3", "RDS", "EKS", "Lambda"}
	regions := []string{"us-east-1", "us-west-2", "eu-west-1"}
	for d := start; !d.After(end); d = d.Add(24 * time.Hour) {
		for i := 0; i < 3; i++ {
			amt := 20 + rand.Float64()*50
			// add occasional spike
			if rand.Intn(20) == 0 {
				amt *= 3
			}
			out = append(out, point{Date: d.Format("2006-01-02"), Amount: amt, Service: services[rand.Intn(len(services))], Region: regions[rand.Intn(len(regions))]})
		}
	}
	writeJSON(w, http.StatusOK, out)
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
	stateMu.Lock()
	defer stateMu.Unlock()
	out := []anomaly{}
	for _, r := range resources {
		// Randomly assign some anomalies
		if rand.Intn(6) == 0 {
			prev := 100 + rand.Intn(200)
			curr := int(float64(prev) * (1.5 + rand.Float64()))
			pct := int(float64(curr-prev) * 100 / float64(prev))
			sev := "medium"
			if pct > 120 {
				sev = "high"
			}
			out = append(out, anomaly{
				ID: "an-" + randID(), ResourceID: r.ID, AnomalyType: "spike", Severity: sev,
				Percentage: pct, PreviousCost: prev, CurrentCost: curr,
				DetectedAt: time.Now().Format(time.RFC3339), Status: "open",
			})
		}
	}
	writeJSON(w, http.StatusOK, out)
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
	suggestions := askOpenAI(r.Context(), prompt, []string{
		"Right-size compute instances based on recent CPU/Memory usage.",
		"Schedule non-production resources to stop outside business hours.",
		"Move infrequently accessed data to cheaper storage tiers.",
	})
	writeJSON(w, http.StatusOK, map[string]any{"resourceId": rid, "suggestions": suggestions})
}

func handleAISecurityAnalysis(w http.ResponseWriter, r *http.Request) {
	rid := chi.URLParam(r, "resourceId")
	prompt := "You are a cloud security assistant. Provide 3 hardening steps for this resource focusing on least privilege, network access, and encryption. Resource ID: " + rid
	suggestions := askOpenAI(r.Context(), prompt, []string{
		"Audit IAM permissions and restrict to least privilege for this resource.",
		"Ensure network access is restricted to required CIDR ranges and ports.",
		"Enforce encryption at rest and in transit; rotate keys regularly.",
	})
	writeJSON(w, http.StatusOK, map[string]any{"resourceId": rid, "suggestions": suggestions})
}

func handleAIRecommendations(w http.ResponseWriter, r *http.Request) {
	rid := chi.URLParam(r, "resourceId")
	prompt := "Summarize the top 5 prioritized actions to reduce cost and improve security posture for the given resource. Resource ID: " + rid
	suggestions := askOpenAI(r.Context(), prompt, []string{
		"Consolidate underutilized instances and adopt autoscaling.",
		"Enable lifecycle policies to transition cold data to archival tiers.",
		"Tighten security groups; remove 0.0.0.0/0 where unnecessary.",
		"Enable detailed monitoring and set anomaly alerts.",
		"Use savings plans or committed use discounts where applicable.",
	})
	writeJSON(w, http.StatusOK, map[string]any{"resourceId": rid, "recommendations": suggestions})
}
