package compliance

import (
	"encoding/json"
	"time"
)

// Framework represents a compliance framework (CIS, NIST, SOC2, etc.)
type Framework struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	Provider    string    `json:"provider,omitempty"` // aws, gcp, azure, or empty for multi-cloud
	IsEnabled   bool      `json:"is_enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Control represents a specific compliance control within a framework
type Control struct {
	ID           string    `json:"id"`
	FrameworkID  string    `json:"framework_id"`
	ControlID    string    `json:"control_id"` // e.g., "2.1.1" for CIS
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Category     string    `json:"category"` // e.g., "Identity and Access Management"
	Severity     string    `json:"severity"` // critical, high, medium, low
	Remediation  string    `json:"remediation"`
	ReferenceURL string    `json:"reference_url,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// ControlMapping maps security rules/drift types to compliance controls
type ControlMapping struct {
	ID                string `json:"id"`
	ControlID         string `json:"control_id"`
	SecurityRuleType  string `json:"security_rule_type"` // drift_type or vuln_type
	ResourceType      string `json:"resource_type"`
	Provider          string `json:"provider,omitempty"`
	MappingConfidence string `json:"mapping_confidence"` // high, medium, low
	CheckQuery        string `json:"check_query,omitempty"`
}

// Assessment represents a compliance assessment run
type Assessment struct {
	ID                    string          `json:"id"`
	UserID                int64           `json:"user_id"`
	FrameworkID           string          `json:"framework_id"`
	FrameworkName         string          `json:"framework_name"`
	AssessmentDate        time.Time       `json:"assessment_date"`
	TotalControls         int             `json:"total_controls"`
	PassedControls        int             `json:"passed_controls"`
	FailedControls        int             `json:"failed_controls"`
	NotApplicableControls int             `json:"not_applicable_controls"`
	CompliancePercent     float64         `json:"compliance_percent"`
	Findings              json.RawMessage `json:"findings,omitempty"`
	Status                string          `json:"status"` // running, completed, failed
	CreatedAt             time.Time       `json:"created_at"`
}

// AssessmentFinding represents a single finding in an assessment
type AssessmentFinding struct {
	ControlID         string   `json:"control_id"`
	ControlTitle      string   `json:"control_title"`
	Category          string   `json:"category"`
	Severity          string   `json:"severity"`
	Status            string   `json:"status"` // passed, failed, not_applicable
	AffectedCount     int      `json:"affected_count"`
	AffectedResources []string `json:"affected_resources,omitempty"`
	Remediation       string   `json:"remediation"`
	Evidence          string   `json:"evidence,omitempty"`
}

// ComplianceStatus represents compliance status for a resource
type ComplianceStatus struct {
	ResourceID      string          `json:"resource_id"`
	ResourceType    string          `json:"resource_type"`
	Provider        string          `json:"provider"`
	ControlStatuses []ControlStatus `json:"control_statuses"`
	OverallStatus   string          `json:"overall_status"` // compliant, non_compliant, partial
	LastChecked     time.Time       `json:"last_checked"`
}

// ControlStatus represents the status of a single control for a resource
type ControlStatus struct {
	FrameworkID string    `json:"framework_id"`
	ControlID   string    `json:"control_id"`
	Title       string    `json:"title"`
	Status      string    `json:"status"` // passed, failed, not_checked
	Remediation string    `json:"remediation,omitempty"`
	LastChecked time.Time `json:"last_checked"`
}

// Framework constants
const (
	FrameworkCISAWS    = "cis-aws"
	FrameworkCISGCP    = "cis-gcp"
	FrameworkCISAzure  = "cis-azure"
	FrameworkNIST80053 = "nist-800-53"
	FrameworkSOC2      = "soc2"
	FrameworkPCIDSS    = "pci-dss"
	FrameworkHIPAA     = "hipaa"
	FrameworkISO27001  = "iso-27001"
)

// Control severity constants
const (
	SeverityCritical = "critical"
	SeverityHigh     = "high"
	SeverityMedium   = "medium"
	SeverityLow      = "low"
)

// Assessment status constants
const (
	AssessmentStatusRunning   = "running"
	AssessmentStatusCompleted = "completed"
	AssessmentStatusFailed    = "failed"
)

// Control status constants
const (
	ControlStatusPassed        = "passed"
	ControlStatusFailed        = "failed"
	ControlStatusNotApplicable = "not_applicable"
	ControlStatusNotChecked    = "not_checked"
)

// Compliance status constants
const (
	StatusCompliant    = "compliant"
	StatusNonCompliant = "non_compliant"
	StatusPartial      = "partial"
)

// Filter for compliance queries
type Filter struct {
	FrameworkID string
	Category    string
	Severity    string
	Status      string
	Provider    string
}
