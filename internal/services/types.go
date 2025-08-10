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
	ProviderAWS   CloudProvider = "aws"
	ProviderGCP   CloudProvider = "gcp"
	ProviderAzure CloudProvider = "azure"
)

type TimeRange struct {
	From time.Time
	To   time.Time
}
