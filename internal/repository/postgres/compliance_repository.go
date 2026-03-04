package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pratik-mahalle/infraudit/internal/domain/compliance"
)

// ComplianceRepository implements compliance.Repository
type ComplianceRepository struct {
	db *sql.DB
}

// NewComplianceRepository creates a new compliance repository
func NewComplianceRepository(db *sql.DB) *ComplianceRepository {
	return &ComplianceRepository{db: db}
}

// CreateFramework creates a new compliance framework
func (r *ComplianceRepository) CreateFramework(ctx context.Context, f *compliance.Framework) error {
	if f.ID == "" {
		f.ID = uuid.New().String()
	}

	query := `
		INSERT INTO compliance_frameworks (id, name, version, description, provider, is_enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		f.ID, f.Name, f.Version, f.Description, f.Provider, f.IsEnabled,
		time.Now(), time.Now(),
	)
	return err
}

// GetFramework retrieves a framework by ID
func (r *ComplianceRepository) GetFramework(ctx context.Context, id string) (*compliance.Framework, error) {
	query := `
		SELECT id, name, version, description, provider, is_enabled, created_at, updated_at
		FROM compliance_frameworks
		WHERE id = $1
	`
	f := &compliance.Framework{}
	var provider sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&f.ID, &f.Name, &f.Version, &f.Description, &provider, &f.IsEnabled,
		&f.CreatedAt, &f.UpdatedAt,
	)
	if provider.Valid {
		f.Provider = provider.String
	}
	if err != nil {
		return nil, err
	}
	return f, nil
}

// GetFrameworkByName retrieves a framework by name
func (r *ComplianceRepository) GetFrameworkByName(ctx context.Context, name string) (*compliance.Framework, error) {
	query := `
		SELECT id, name, version, description, provider, is_enabled, created_at, updated_at
		FROM compliance_frameworks
		WHERE name = $1
	`
	f := &compliance.Framework{}
	var provider sql.NullString
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&f.ID, &f.Name, &f.Version, &f.Description, &provider, &f.IsEnabled,
		&f.CreatedAt, &f.UpdatedAt,
	)
	if provider.Valid {
		f.Provider = provider.String
	}
	if err != nil {
		return nil, err
	}
	return f, nil
}

// ListFrameworks lists all frameworks
func (r *ComplianceRepository) ListFrameworks(ctx context.Context) ([]*compliance.Framework, error) {
	query := `
		SELECT id, name, version, description, provider, is_enabled, created_at, updated_at
		FROM compliance_frameworks
		ORDER BY name
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var frameworks []*compliance.Framework
	for rows.Next() {
		f := &compliance.Framework{}
		var provider sql.NullString
		err := rows.Scan(
			&f.ID, &f.Name, &f.Version, &f.Description, &provider, &f.IsEnabled,
			&f.CreatedAt, &f.UpdatedAt,
		)
		if provider.Valid {
			f.Provider = provider.String
		}
		if err != nil {
			return nil, err
		}
		frameworks = append(frameworks, f)
	}

	return frameworks, rows.Err()
}

// UpdateFramework updates a framework
func (r *ComplianceRepository) UpdateFramework(ctx context.Context, f *compliance.Framework) error {
	query := `UPDATE compliance_frameworks SET is_enabled = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, f.IsEnabled, time.Now(), f.ID)
	return err
}

// CreateControl creates a new compliance control
func (r *ComplianceRepository) CreateControl(ctx context.Context, c *compliance.Control) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}

	query := `
		INSERT INTO compliance_controls (id, framework_id, control_id, title, description, category, severity, remediation, reference_url, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.ExecContext(ctx, query,
		c.ID, c.FrameworkID, c.ControlID, c.Title, c.Description, c.Category,
		c.Severity, c.Remediation, c.ReferenceURL, time.Now(),
	)
	return err
}

// GetControl retrieves a control by ID
func (r *ComplianceRepository) GetControl(ctx context.Context, id string) (*compliance.Control, error) {
	query := `
		SELECT id, framework_id, control_id, title, description, category, severity, remediation, reference_url, created_at
		FROM compliance_controls
		WHERE id = $1
	`
	c := &compliance.Control{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID, &c.FrameworkID, &c.ControlID, &c.Title, &c.Description, &c.Category,
		&c.Severity, &c.Remediation, &c.ReferenceURL, &c.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// GetControlByFrameworkAndID retrieves a control by framework ID and control ID
func (r *ComplianceRepository) GetControlByFrameworkAndID(ctx context.Context, frameworkID, controlID string) (*compliance.Control, error) {
	query := `
		SELECT id, framework_id, control_id, title, description, category, severity, remediation, reference_url, created_at
		FROM compliance_controls
		WHERE framework_id = $1 AND control_id = $2
	`
	c := &compliance.Control{}
	err := r.db.QueryRowContext(ctx, query, frameworkID, controlID).Scan(
		&c.ID, &c.FrameworkID, &c.ControlID, &c.Title, &c.Description, &c.Category,
		&c.Severity, &c.Remediation, &c.ReferenceURL, &c.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// ListControls lists controls for a framework
func (r *ComplianceRepository) ListControls(ctx context.Context, frameworkID string, category string) ([]*compliance.Control, error) {
	paramN := 1
	query := fmt.Sprintf(`
		SELECT id, framework_id, control_id, title, description, category, severity, remediation, reference_url, created_at
		FROM compliance_controls
		WHERE framework_id = $%d
	`, paramN)
	args := []interface{}{frameworkID}
	paramN++

	if category != "" {
		query += fmt.Sprintf(" AND category = $%d", paramN)
		args = append(args, category)
		paramN++
	}

	query += " ORDER BY control_id"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var controls []*compliance.Control
	for rows.Next() {
		c := &compliance.Control{}
		err := rows.Scan(
			&c.ID, &c.FrameworkID, &c.ControlID, &c.Title, &c.Description, &c.Category,
			&c.Severity, &c.Remediation, &c.ReferenceURL, &c.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		controls = append(controls, c)
	}

	return controls, rows.Err()
}

// CountControls counts controls for a framework
func (r *ComplianceRepository) CountControls(ctx context.Context, frameworkID string) (int, error) {
	query := `SELECT COUNT(*) FROM compliance_controls WHERE framework_id = $1`
	var count int
	err := r.db.QueryRowContext(ctx, query, frameworkID).Scan(&count)
	return count, err
}

// CreateMapping creates a new control mapping
func (r *ComplianceRepository) CreateMapping(ctx context.Context, m *compliance.ControlMapping) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}

	query := `
		INSERT INTO compliance_mappings (id, control_id, security_rule_type, resource_type, provider, mapping_confidence, check_query, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		m.ID, m.ControlID, m.SecurityRuleType, m.ResourceType, m.Provider,
		m.MappingConfidence, m.CheckQuery, time.Now(),
	)
	return err
}

// GetMappingsForControl retrieves mappings for a control
func (r *ComplianceRepository) GetMappingsForControl(ctx context.Context, controlID string) ([]*compliance.ControlMapping, error) {
	query := `
		SELECT id, control_id, security_rule_type, resource_type, provider, mapping_confidence, check_query
		FROM compliance_mappings
		WHERE control_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, controlID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mappings []*compliance.ControlMapping
	for rows.Next() {
		m := &compliance.ControlMapping{}
		err := rows.Scan(
			&m.ID, &m.ControlID, &m.SecurityRuleType, &m.ResourceType, &m.Provider,
			&m.MappingConfidence, &m.CheckQuery,
		)
		if err != nil {
			return nil, err
		}
		mappings = append(mappings, m)
	}

	return mappings, rows.Err()
}

// GetMappingsForSecurityRule retrieves mappings for a security rule type
func (r *ComplianceRepository) GetMappingsForSecurityRule(ctx context.Context, ruleType string, resourceType string) ([]*compliance.ControlMapping, error) {
	paramN := 1
	query := fmt.Sprintf(`
		SELECT id, control_id, security_rule_type, resource_type, provider, mapping_confidence, check_query
		FROM compliance_mappings
		WHERE security_rule_type = $%d
	`, paramN)
	args := []interface{}{ruleType}
	paramN++

	if resourceType != "" {
		query += fmt.Sprintf(" AND (resource_type = $%d OR resource_type IS NULL)", paramN)
		args = append(args, resourceType)
		paramN++
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mappings []*compliance.ControlMapping
	for rows.Next() {
		m := &compliance.ControlMapping{}
		err := rows.Scan(
			&m.ID, &m.ControlID, &m.SecurityRuleType, &m.ResourceType, &m.Provider,
			&m.MappingConfidence, &m.CheckQuery,
		)
		if err != nil {
			return nil, err
		}
		mappings = append(mappings, m)
	}

	return mappings, rows.Err()
}

// CreateAssessment creates a new compliance assessment
func (r *ComplianceRepository) CreateAssessment(ctx context.Context, a *compliance.Assessment) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}

	findingsJSON, _ := json.Marshal(a.Findings)

	query := `
		INSERT INTO compliance_assessments (id, user_id, framework_id, framework_name, assessment_date, total_controls, passed_controls, failed_controls, not_applicable_controls, compliance_percent, findings, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.db.ExecContext(ctx, query,
		a.ID, a.UserID, a.FrameworkID, a.FrameworkName, a.AssessmentDate,
		a.TotalControls, a.PassedControls, a.FailedControls, a.NotApplicableControls,
		a.CompliancePercent, findingsJSON, a.Status, time.Now(),
	)
	return err
}

// GetAssessment retrieves an assessment by ID
func (r *ComplianceRepository) GetAssessment(ctx context.Context, id string) (*compliance.Assessment, error) {
	query := `
		SELECT id, user_id, framework_id, framework_name, assessment_date, total_controls, passed_controls, failed_controls, not_applicable_controls, compliance_percent, findings, status, created_at
		FROM compliance_assessments
		WHERE id = $1
	`
	a := &compliance.Assessment{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&a.ID, &a.UserID, &a.FrameworkID, &a.FrameworkName, &a.AssessmentDate,
		&a.TotalControls, &a.PassedControls, &a.FailedControls, &a.NotApplicableControls,
		&a.CompliancePercent, &a.Findings, &a.Status, &a.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return a, nil
}

// UpdateAssessment updates an assessment
func (r *ComplianceRepository) UpdateAssessment(ctx context.Context, a *compliance.Assessment) error {
	findingsJSON, _ := json.Marshal(a.Findings)

	query := `
		UPDATE compliance_assessments
		SET total_controls = $1, passed_controls = $2, failed_controls = $3, not_applicable_controls = $4,
		    compliance_percent = $5, findings = $6, status = $7
		WHERE id = $8
	`
	_, err := r.db.ExecContext(ctx, query,
		a.TotalControls, a.PassedControls, a.FailedControls, a.NotApplicableControls,
		a.CompliancePercent, findingsJSON, a.Status, a.ID,
	)
	return err
}

// ListAssessments lists assessments
func (r *ComplianceRepository) ListAssessments(ctx context.Context, userID int64, frameworkID string, limit, offset int) ([]*compliance.Assessment, int64, error) {
	paramN := 1
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM compliance_assessments WHERE user_id = $%d`, paramN)
	args := []interface{}{userID}
	paramN++

	if frameworkID != "" {
		countQuery += fmt.Sprintf(" AND framework_id = $%d", paramN)
		args = append(args, frameworkID)
		paramN++
	}

	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Reset for main query
	paramN = 1
	query := fmt.Sprintf(`
		SELECT id, user_id, framework_id, framework_name, assessment_date, total_controls, passed_controls, failed_controls, not_applicable_controls, compliance_percent, findings, status, created_at
		FROM compliance_assessments
		WHERE user_id = $%d
	`, paramN)
	queryArgs := []interface{}{userID}
	paramN++

	if frameworkID != "" {
		query += fmt.Sprintf(" AND framework_id = $%d", paramN)
		queryArgs = append(queryArgs, frameworkID)
		paramN++
	}

	query += fmt.Sprintf(" ORDER BY assessment_date DESC LIMIT $%d OFFSET $%d", paramN, paramN+1)
	queryArgs = append(queryArgs, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var assessments []*compliance.Assessment
	for rows.Next() {
		a := &compliance.Assessment{}
		err := rows.Scan(
			&a.ID, &a.UserID, &a.FrameworkID, &a.FrameworkName, &a.AssessmentDate,
			&a.TotalControls, &a.PassedControls, &a.FailedControls, &a.NotApplicableControls,
			&a.CompliancePercent, &a.Findings, &a.Status, &a.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		assessments = append(assessments, a)
	}

	return assessments, total, rows.Err()
}

// GetLatestAssessment retrieves the latest assessment for a framework
func (r *ComplianceRepository) GetLatestAssessment(ctx context.Context, userID int64, frameworkID string) (*compliance.Assessment, error) {
	query := `
		SELECT id, user_id, framework_id, framework_name, assessment_date, total_controls, passed_controls, failed_controls, not_applicable_controls, compliance_percent, findings, status, created_at
		FROM compliance_assessments
		WHERE user_id = $1 AND framework_id = $2 AND status = 'completed'
		ORDER BY assessment_date DESC
		LIMIT 1
	`
	a := &compliance.Assessment{}
	err := r.db.QueryRowContext(ctx, query, userID, frameworkID).Scan(
		&a.ID, &a.UserID, &a.FrameworkID, &a.FrameworkName, &a.AssessmentDate,
		&a.TotalControls, &a.PassedControls, &a.FailedControls, &a.NotApplicableControls,
		&a.CompliancePercent, &a.Findings, &a.Status, &a.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return a, nil
}
