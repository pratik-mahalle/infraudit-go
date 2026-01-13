package services

import "time"

// Core domain types kept minimal; expand as needed.

type User struct {
	ID          int64
	Username    string
	Email       string
	FullName    *string
	Role        string
	PlanType    string
	TrialStatus string
}

type TrialStatus struct {
	Status        string
	DaysRemaining int
}

type OAuthProvider string

const (
	OAuthGoogle OAuthProvider = "google"
	OAuthGitHub OAuthProvider = "github"
)

type CloudProvider string

const (
	ProviderAWS          CloudProvider = "aws"
	ProviderGCP          CloudProvider = "gcp"
	ProviderAzure        CloudProvider = "azure"
	ProviderDigitalOcean CloudProvider = "digitalocean"
)

type TimeRange struct {
	From time.Time
	To   time.Time
}

// CloudResource is a normalized representation of a resource discovered from any provider.
// It is intentionally small and UI-friendly.
type CloudResource struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	Provider      string `json:"provider"`
	Region        string `json:"region"`
	Status        string `json:"status"`
	Configuration string `json:"configuration"`
}
