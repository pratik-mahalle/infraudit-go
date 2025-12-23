package user

import "time"

// User represents a user in the system
type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	Username     string    `json:"username,omitempty"`
	FullName     *string   `json:"full_name,omitempty"`
	PasswordHash string    `json:"-"` // Not exposed in JSON
	Role         string    `json:"role"`
	PlanType     string    `json:"plan_type"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Plan types
const (
	PlanTypeFree       = "free"
	PlanTypeTrial      = "trial"
	PlanTypeStarter    = "starter"
	PlanTypePro        = "pro"
	PlanTypeEnterprise = "enterprise"
)

// User roles
const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

// TrialStatus represents trial information
type TrialStatus struct {
	Status        string `json:"status"`
	DaysRemaining int    `json:"days_remaining"`
}
