package remediation

import (
	"encoding/json"
	"time"
)

// Action represents a remediation action
type Action struct {
	ID               string          `json:"id"`
	UserID           int64           `json:"user_id"`
	DriftID          *string         `json:"drift_id,omitempty"`
	VulnerabilityID  *string         `json:"vulnerability_id,omitempty"`
	RemediationType  RemediationType `json:"remediation_type"`
	Status           ActionStatus    `json:"status"`
	Strategy         *Strategy       `json:"strategy"`
	ApprovalRequired bool            `json:"approval_required"`
	ApprovedBy       *int64          `json:"approved_by,omitempty"`
	ApprovedAt       *time.Time      `json:"approved_at,omitempty"`
	StartedAt        *time.Time      `json:"started_at,omitempty"`
	CompletedAt      *time.Time      `json:"completed_at,omitempty"`
	Result           json.RawMessage `json:"result,omitempty"`
	RollbackData     json.RawMessage `json:"rollback_data,omitempty"`
	ErrorMessage     string          `json:"error_message,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

// RemediationType represents the type of remediation
type RemediationType string

const (
	RemediationTypeIaCPR    RemediationType = "iac_pr"
	RemediationTypeCloudAPI RemediationType = "cloud_api"
	RemediationTypePolicy   RemediationType = "policy"
	RemediationTypeManual   RemediationType = "manual"
)

// ActionStatus represents the status of a remediation action
type ActionStatus string

const (
	ActionStatusPending    ActionStatus = "pending"
	ActionStatusApproved   ActionStatus = "approved"
	ActionStatusInProgress ActionStatus = "in_progress"
	ActionStatusCompleted  ActionStatus = "completed"
	ActionStatusFailed     ActionStatus = "failed"
	ActionStatusRejected   ActionStatus = "rejected"
	ActionStatusRolledBack ActionStatus = "rolled_back"
)

// Strategy contains the remediation strategy details
type Strategy struct {
	Type        RemediationType        `json:"type"`
	Description string                 `json:"description"`
	Steps       []RemediationStep      `json:"steps"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`

	// IaC PR specific
	Repository   string `json:"repository,omitempty"`
	Branch       string `json:"branch,omitempty"`
	FilePath     string `json:"file_path,omitempty"`
	PatchContent string `json:"patch_content,omitempty"`

	// Cloud API specific
	Provider  string `json:"provider,omitempty"`
	APIAction string `json:"api_action,omitempty"`
	APIParams string `json:"api_params,omitempty"`

	// Policy specific
	PolicyName string `json:"policy_name,omitempty"`
	PolicyRule string `json:"policy_rule,omitempty"`
}

// RemediationStep represents a single step in the remediation process
type RemediationStep struct {
	Order       int                    `json:"order"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Status      string                 `json:"status"` // pending, running, completed, failed, skipped
	Command     string                 `json:"command,omitempty"`
	Result      map[string]interface{} `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

// Suggestion represents a remediation suggestion
type Suggestion struct {
	ID              string          `json:"id"`
	IssueType       string          `json:"issue_type"` // drift, vulnerability
	IssueID         string          `json:"issue_id"`
	Title           string          `json:"title"`
	Description     string          `json:"description"`
	Severity        string          `json:"severity"`
	RemediationType RemediationType `json:"remediation_type"`
	Strategy        *Strategy       `json:"strategy"`
	Risk            string          `json:"risk"` // low, medium, high
	Impact          string          `json:"impact"`
	EstimatedTime   string          `json:"estimated_time"`
}

// Filter contains remediation action filtering options
type Filter struct {
	UserID          int64
	Status          ActionStatus
	RemediationType RemediationType
	DriftID         *string
	VulnerabilityID *string
	From            *time.Time
	To              *time.Time
}

// Result represents the result of a remediation action
type Result struct {
	Success          bool                   `json:"success"`
	Message          string                 `json:"message"`
	ChangesMade      []string               `json:"changes_made,omitempty"`
	PullRequestURL   string                 `json:"pull_request_url,omitempty"`
	RollbackPossible bool                   `json:"rollback_possible"`
	Details          map[string]interface{} `json:"details,omitempty"`
}

// RollbackData contains data needed to rollback a remediation
type RollbackData struct {
	OriginalState    map[string]interface{} `json:"original_state"`
	RollbackSteps    []RemediationStep      `json:"rollback_steps"`
	CanAutoRollback  bool                   `json:"can_auto_rollback"`
	RollbackDeadline *time.Time             `json:"rollback_deadline,omitempty"`
}

// IsValid checks if the remediation type is valid
func (rt RemediationType) IsValid() bool {
	switch rt {
	case RemediationTypeIaCPR, RemediationTypeCloudAPI, RemediationTypePolicy, RemediationTypeManual:
		return true
	default:
		return false
	}
}

// IsTerminal checks if the action status is terminal
func (as ActionStatus) IsTerminal() bool {
	return as == ActionStatusCompleted || as == ActionStatusFailed ||
		as == ActionStatusRejected || as == ActionStatusRolledBack
}

// RequiresApproval determines if the remediation type typically requires approval
func (rt RemediationType) RequiresApproval() bool {
	return rt == RemediationTypeCloudAPI || rt == RemediationTypeIaCPR
}
