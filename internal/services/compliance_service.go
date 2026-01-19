package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/pratik-mahalle/infraudit/internal/domain/compliance"
	"github.com/pratik-mahalle/infraudit/internal/domain/drift"
	"github.com/pratik-mahalle/infraudit/internal/domain/vulnerability"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
)

// ComplianceService implements compliance.Service
type ComplianceServiceImpl struct {
	repo      compliance.Repository
	driftRepo drift.Repository
	vulnRepo  vulnerability.Repository
	logger    *logger.Logger
}

// NewComplianceService creates a new compliance service
func NewComplianceService(
	repo compliance.Repository,
	driftRepo drift.Repository,
	vulnRepo vulnerability.Repository,
	log *logger.Logger,
) compliance.Service {
	return &ComplianceServiceImpl{
		repo:      repo,
		driftRepo: driftRepo,
		vulnRepo:  vulnRepo,
		logger:    log,
	}
}

// ListFrameworks lists all available compliance frameworks
func (s *ComplianceServiceImpl) ListFrameworks(ctx context.Context) ([]*compliance.Framework, error) {
	return s.repo.ListFrameworks(ctx)
}

// GetFramework retrieves a specific framework
func (s *ComplianceServiceImpl) GetFramework(ctx context.Context, id string) (*compliance.Framework, error) {
	return s.repo.GetFramework(ctx, id)
}

// EnableFramework enables a compliance framework
func (s *ComplianceServiceImpl) EnableFramework(ctx context.Context, id string) error {
	framework, err := s.repo.GetFramework(ctx, id)
	if err != nil {
		return err
	}
	framework.IsEnabled = true
	return s.repo.UpdateFramework(ctx, framework)
}

// DisableFramework disables a compliance framework
func (s *ComplianceServiceImpl) DisableFramework(ctx context.Context, id string) error {
	framework, err := s.repo.GetFramework(ctx, id)
	if err != nil {
		return err
	}
	framework.IsEnabled = false
	return s.repo.UpdateFramework(ctx, framework)
}

// GetControls retrieves controls for a framework
func (s *ComplianceServiceImpl) GetControls(ctx context.Context, frameworkID string, category string) ([]*compliance.Control, error) {
	return s.repo.ListControls(ctx, frameworkID, category)
}

// GetControlDetails retrieves a control with its mappings
func (s *ComplianceServiceImpl) GetControlDetails(ctx context.Context, controlID string) (*compliance.Control, []*compliance.ControlMapping, error) {
	control, err := s.repo.GetControl(ctx, controlID)
	if err != nil {
		return nil, nil, err
	}

	mappings, err := s.repo.GetMappingsForControl(ctx, controlID)
	if err != nil {
		return nil, nil, err
	}

	return control, mappings, nil
}

// GetFailingControls retrieves failing controls for a user
func (s *ComplianceServiceImpl) GetFailingControls(ctx context.Context, userID int64, frameworkID string) ([]*compliance.AssessmentFinding, error) {
	assessment, err := s.repo.GetLatestAssessment(ctx, userID, frameworkID)
	if err != nil {
		return nil, err
	}

	var findings []compliance.AssessmentFinding
	if err := json.Unmarshal(assessment.Findings, &findings); err != nil {
		return nil, err
	}

	var failingControls []*compliance.AssessmentFinding
	for i := range findings {
		if findings[i].Status == compliance.ControlStatusFailed {
			failingControls = append(failingControls, &findings[i])
		}
	}

	return failingControls, nil
}

// RunAssessment runs a compliance assessment
func (s *ComplianceServiceImpl) RunAssessment(ctx context.Context, userID int64, frameworkID string) (*compliance.Assessment, error) {
	framework, err := s.repo.GetFramework(ctx, frameworkID)
	if err != nil {
		return nil, err
	}

	// Create assessment record
	assessment := &compliance.Assessment{
		ID:             uuid.New().String(),
		UserID:         userID,
		FrameworkID:    frameworkID,
		FrameworkName:  framework.Name,
		AssessmentDate: time.Now(),
		Status:         compliance.AssessmentStatusRunning,
	}

	if err := s.repo.CreateAssessment(ctx, assessment); err != nil {
		return nil, err
	}

	// Run assessment asynchronously
	go func() {
		execCtx := context.Background()
		s.executeAssessment(execCtx, assessment, framework)
	}()

	return assessment, nil
}

// executeAssessment performs the actual assessment
func (s *ComplianceServiceImpl) executeAssessment(ctx context.Context, assessment *compliance.Assessment, framework *compliance.Framework) {
	// Get all controls for the framework
	controls, err := s.repo.ListControls(ctx, framework.ID, "")
	if err != nil {
		assessment.Status = compliance.AssessmentStatusFailed
		s.repo.UpdateAssessment(ctx, assessment)
		return
	}

	assessment.TotalControls = len(controls)
	var findings []compliance.AssessmentFinding

	for _, control := range controls {
		finding := s.evaluateControl(ctx, assessment.UserID, control)
		findings = append(findings, finding)

		switch finding.Status {
		case compliance.ControlStatusPassed:
			assessment.PassedControls++
		case compliance.ControlStatusFailed:
			assessment.FailedControls++
		case compliance.ControlStatusNotApplicable:
			assessment.NotApplicableControls++
		}
	}

	// Calculate compliance percentage
	applicable := assessment.TotalControls - assessment.NotApplicableControls
	if applicable > 0 {
		assessment.CompliancePercent = (float64(assessment.PassedControls) / float64(applicable)) * 100
	}

	// Serialize findings
	findingsJSON, _ := json.Marshal(findings)
	assessment.Findings = findingsJSON
	assessment.Status = compliance.AssessmentStatusCompleted

	s.repo.UpdateAssessment(ctx, assessment)

	s.logger.WithFields(map[string]interface{}{
		"assessment_id":      assessment.ID,
		"framework":          assessment.FrameworkName,
		"compliance_percent": assessment.CompliancePercent,
	}).Info("Compliance assessment completed")
}

// evaluateControl evaluates a single control
func (s *ComplianceServiceImpl) evaluateControl(ctx context.Context, userID int64, control *compliance.Control) compliance.AssessmentFinding {
	finding := compliance.AssessmentFinding{
		ControlID:    control.ControlID,
		ControlTitle: control.Title,
		Category:     control.Category,
		Severity:     control.Severity,
		Remediation:  control.Remediation,
		Status:       compliance.ControlStatusPassed, // Default to passed
	}

	// Get mappings for this control
	mappings, _ := s.repo.GetMappingsForControl(ctx, control.ID)
	if len(mappings) == 0 {
		finding.Status = compliance.ControlStatusNotApplicable
		return finding
	}

	// Check for violations based on mappings
	for _, mapping := range mappings {
		violations := s.checkForViolations(ctx, userID, mapping)
		if len(violations) > 0 {
			finding.Status = compliance.ControlStatusFailed
			finding.AffectedCount = len(violations)
			finding.AffectedResources = violations
			break
		}
	}

	return finding
}

// checkForViolations checks for violations based on a control mapping
func (s *ComplianceServiceImpl) checkForViolations(ctx context.Context, userID int64, mapping *compliance.ControlMapping) []string {
	var violations []string

	// Check drifts
	if mapping.SecurityRuleType != "" {
		drifts, _ := s.driftRepo.List(ctx, userID, drift.Filter{DriftType: mapping.SecurityRuleType})
		for _, d := range drifts {
			if d.Status == drift.StatusDetected {
				violations = append(violations, d.ResourceID)
			}
		}
	}

	return violations
}

// GetAssessment retrieves an assessment
func (s *ComplianceServiceImpl) GetAssessment(ctx context.Context, id string) (*compliance.Assessment, error) {
	return s.repo.GetAssessment(ctx, id)
}

// ListAssessments lists assessments
func (s *ComplianceServiceImpl) ListAssessments(ctx context.Context, userID int64, frameworkID string, limit, offset int) ([]*compliance.Assessment, int64, error) {
	return s.repo.ListAssessments(ctx, userID, frameworkID, limit, offset)
}

// GetLatestAssessment retrieves the latest assessment
func (s *ComplianceServiceImpl) GetLatestAssessment(ctx context.Context, userID int64, frameworkID string) (*compliance.Assessment, error) {
	return s.repo.GetLatestAssessment(ctx, userID, frameworkID)
}

// GetResourceCompliance retrieves compliance status for a resource
func (s *ComplianceServiceImpl) GetResourceCompliance(ctx context.Context, userID int64, resourceID string) (*compliance.ComplianceStatus, error) {
	// TODO: Implement resource-level compliance checking
	return &compliance.ComplianceStatus{
		ResourceID:    resourceID,
		OverallStatus: compliance.StatusCompliant,
		LastChecked:   time.Now(),
	}, nil
}

// GetComplianceOverview returns a high-level compliance summary
func (s *ComplianceServiceImpl) GetComplianceOverview(ctx context.Context, userID int64) (*compliance.ComplianceOverview, error) {
	frameworks, err := s.repo.ListFrameworks(ctx)
	if err != nil {
		return nil, err
	}

	overview := &compliance.ComplianceOverview{
		ByFramework: make([]compliance.FrameworkCompliance, 0),
		BySeverity:  make(map[string]int),
	}

	for _, framework := range frameworks {
		if !framework.IsEnabled {
			continue
		}

		latest, err := s.repo.GetLatestAssessment(ctx, userID, framework.ID)
		if err != nil {
			continue
		}

		fc := compliance.FrameworkCompliance{
			FrameworkID:       framework.ID,
			FrameworkName:     framework.Name,
			TotalControls:     latest.TotalControls,
			PassedControls:    latest.PassedControls,
			FailedControls:    latest.FailedControls,
			CompliancePercent: latest.CompliancePercent,
			LastAssessment:    latest.AssessmentDate.Format(time.RFC3339),
		}

		overview.ByFramework = append(overview.ByFramework, fc)
		overview.TotalControls += latest.TotalControls
		overview.PassedControls += latest.PassedControls
		overview.FailedControls += latest.FailedControls
	}

	if overview.TotalControls > 0 {
		overview.CompliancePercent = (float64(overview.PassedControls) / float64(overview.TotalControls)) * 100
	}

	return overview, nil
}

// GetComplianceTrend returns compliance trend over time
func (s *ComplianceServiceImpl) GetComplianceTrend(ctx context.Context, userID int64, frameworkID string, days int) (*compliance.ComplianceTrend, error) {
	assessments, _, err := s.repo.ListAssessments(ctx, userID, frameworkID, days, 0)
	if err != nil {
		return nil, err
	}

	trend := &compliance.ComplianceTrend{
		FrameworkID: frameworkID,
		DataPoints:  make([]compliance.TrendDataPoint, 0, len(assessments)),
	}

	for _, a := range assessments {
		trend.DataPoints = append(trend.DataPoints, compliance.TrendDataPoint{
			Date:            a.AssessmentDate.Format("2006-01-02"),
			ComplianceScore: a.CompliancePercent,
			PassedControls:  a.PassedControls,
			TotalControls:   a.TotalControls,
		})
	}

	if len(assessments) > 0 {
		trend.CurrentScore = assessments[0].CompliancePercent
		if len(assessments) > 1 {
			trend.PreviousScore = assessments[1].CompliancePercent
			trend.ChangePercent = trend.CurrentScore - trend.PreviousScore

			if trend.ChangePercent > 0 {
				trend.Trend = "improving"
			} else if trend.ChangePercent < 0 {
				trend.Trend = "declining"
			} else {
				trend.Trend = "stable"
			}
		}
	}

	return trend, nil
}

// GenerateReport generates a compliance report
func (s *ComplianceServiceImpl) GenerateReport(ctx context.Context, assessmentID string, format string) ([]byte, error) {
	assessment, err := s.repo.GetAssessment(ctx, assessmentID)
	if err != nil {
		return nil, err
	}

	export := &compliance.AssessmentExport{
		Assessment:  assessment,
		GeneratedAt: time.Now().Format(time.RFC3339),
	}

	return json.Marshal(export)
}

// ExportAssessment exports an assessment
func (s *ComplianceServiceImpl) ExportAssessment(ctx context.Context, assessmentID string) (*compliance.AssessmentExport, error) {
	assessment, err := s.repo.GetAssessment(ctx, assessmentID)
	if err != nil {
		return nil, err
	}

	framework, _ := s.repo.GetFramework(ctx, assessment.FrameworkID)

	var findings []compliance.AssessmentFinding
	json.Unmarshal(assessment.Findings, &findings)

	summary := &compliance.ExportSummary{
		TotalControls: assessment.TotalControls,
		Passed:        assessment.PassedControls,
		Failed:        assessment.FailedControls,
		Score:         assessment.CompliancePercent,
	}

	for _, f := range findings {
		switch f.Severity {
		case compliance.SeverityCritical:
			summary.CriticalCount++
		case compliance.SeverityHigh:
			summary.HighCount++
		case compliance.SeverityMedium:
			summary.MediumCount++
		case compliance.SeverityLow:
			summary.LowCount++
		}
	}

	return &compliance.AssessmentExport{
		Assessment:  assessment,
		Framework:   framework,
		Findings:    findings,
		Summary:     summary,
		GeneratedAt: time.Now().Format(time.RFC3339),
	}, nil
}

// InitializeFrameworks initializes built-in compliance frameworks
func (s *ComplianceServiceImpl) InitializeFrameworks(ctx context.Context) error {
	// The frameworks are seeded via migration, but we can add controls here
	// For now, we check if CIS AWS controls exist and add them if not

	framework, err := s.repo.GetFrameworkByName(ctx, "CIS AWS Foundations Benchmark")
	if err != nil {
		return nil // Framework not found, skip
	}

	count, _ := s.repo.CountControls(ctx, framework.ID)
	if count > 0 {
		return nil // Controls already exist
	}

	// Add CIS AWS controls
	cisControls := compliance.CISAWSControls()
	for _, control := range cisControls {
		control.FrameworkID = framework.ID
		s.repo.CreateControl(ctx, control)
	}

	s.logger.WithFields(map[string]interface{}{
		"framework": framework.Name,
		"controls":  len(cisControls),
	}).Info("Initialized compliance framework controls")

	return nil
}
