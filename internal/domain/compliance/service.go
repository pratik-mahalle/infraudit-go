package compliance

import (
	"context"
)

// Service defines the compliance service interface
type Service interface {
	// Framework Management
	ListFrameworks(ctx context.Context) ([]*Framework, error)
	GetFramework(ctx context.Context, id string) (*Framework, error)
	EnableFramework(ctx context.Context, id string) error
	DisableFramework(ctx context.Context, id string) error

	// Controls
	GetControls(ctx context.Context, frameworkID string, category string) ([]*Control, error)
	GetControlDetails(ctx context.Context, controlID string) (*Control, []*ControlMapping, error)
	GetFailingControls(ctx context.Context, userID int64, frameworkID string) ([]*AssessmentFinding, error)

	// Assessments
	RunAssessment(ctx context.Context, userID int64, frameworkID string) (*Assessment, error)
	GetAssessment(ctx context.Context, id string) (*Assessment, error)
	ListAssessments(ctx context.Context, userID int64, frameworkID string, limit, offset int) ([]*Assessment, int64, error)
	GetLatestAssessment(ctx context.Context, userID int64, frameworkID string) (*Assessment, error)

	// Compliance Status
	GetResourceCompliance(ctx context.Context, userID int64, resourceID string) (*ComplianceStatus, error)
	GetComplianceOverview(ctx context.Context, userID int64) (*ComplianceOverview, error)
	GetComplianceTrend(ctx context.Context, userID int64, frameworkID string, days int) (*ComplianceTrend, error)

	// Reports
	GenerateReport(ctx context.Context, assessmentID string, format string) ([]byte, error)
	ExportAssessment(ctx context.Context, assessmentID string) (*AssessmentExport, error)

	// Initialization
	InitializeFrameworks(ctx context.Context) error
}

// ComplianceOverview represents a high-level compliance summary
type ComplianceOverview struct {
	TotalControls      int                   `json:"total_controls"`
	PassedControls     int                   `json:"passed_controls"`
	FailedControls     int                   `json:"failed_controls"`
	CompliancePercent  float64               `json:"compliance_percent"`
	ByFramework        []FrameworkCompliance `json:"by_framework"`
	TopFailingControls []AssessmentFinding   `json:"top_failing_controls"`
	BySeverity         map[string]int        `json:"by_severity"`
}

// FrameworkCompliance represents compliance status for a specific framework
type FrameworkCompliance struct {
	FrameworkID       string  `json:"framework_id"`
	FrameworkName     string  `json:"framework_name"`
	TotalControls     int     `json:"total_controls"`
	PassedControls    int     `json:"passed_controls"`
	FailedControls    int     `json:"failed_controls"`
	CompliancePercent float64 `json:"compliance_percent"`
	LastAssessment    string  `json:"last_assessment,omitempty"`
}

// ComplianceTrend represents compliance changes over time
type ComplianceTrend struct {
	FrameworkID   string           `json:"framework_id"`
	FrameworkName string           `json:"framework_name"`
	DataPoints    []TrendDataPoint `json:"data_points"`
	CurrentScore  float64          `json:"current_score"`
	PreviousScore float64          `json:"previous_score"`
	ChangePercent float64          `json:"change_percent"`
	Trend         string           `json:"trend"` // improving, declining, stable
}

// TrendDataPoint represents a single compliance data point
type TrendDataPoint struct {
	Date            string  `json:"date"`
	ComplianceScore float64 `json:"compliance_score"`
	PassedControls  int     `json:"passed_controls"`
	TotalControls   int     `json:"total_controls"`
}

// AssessmentExport represents an exportable assessment report
type AssessmentExport struct {
	Assessment  *Assessment         `json:"assessment"`
	Framework   *Framework          `json:"framework"`
	Findings    []AssessmentFinding `json:"findings"`
	Summary     *ExportSummary      `json:"summary"`
	GeneratedAt string              `json:"generated_at"`
}

// ExportSummary provides a summary for exports
type ExportSummary struct {
	TotalControls int     `json:"total_controls"`
	Passed        int     `json:"passed"`
	Failed        int     `json:"failed"`
	Score         float64 `json:"score"`
	CriticalCount int     `json:"critical_count"`
	HighCount     int     `json:"high_count"`
	MediumCount   int     `json:"medium_count"`
	LowCount      int     `json:"low_count"`
}
