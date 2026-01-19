package scanners

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"time"

	"github.com/pratik-mahalle/infraudit/internal/domain/vulnerability"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
)

// CheckovScanner wraps the Checkov security scanner for IaC files
type CheckovScanner struct {
	logger      *logger.Logger
	checkovPath string
	scanTimeout time.Duration
}

// NewCheckovScanner creates a new Checkov scanner instance
func NewCheckovScanner(log *logger.Logger, checkovPath string) *CheckovScanner {
	if checkovPath == "" {
		checkovPath = "checkov"
	}
	return &CheckovScanner{
		logger:      log,
		checkovPath: checkovPath,
		scanTimeout: 5 * time.Minute,
	}
}

// CheckovResult represents Checkov JSON output
type CheckovResult struct {
	CheckType   string         `json:"check_type"`
	Results     CheckovResults `json:"results"`
	SummaryLine string         `json:"summary_line"`
}

// CheckovResults contains passed and failed checks
type CheckovResults struct {
	PassedChecks  []CheckovCheck `json:"passed_checks"`
	FailedChecks  []CheckovCheck `json:"failed_checks"`
	SkippedChecks []CheckovCheck `json:"skipped_checks"`
}

// CheckovCheck represents a single Checkov check result
type CheckovCheck struct {
	CheckID         string             `json:"check_id"`
	BCCheckID       string             `json:"bc_check_id,omitempty"`
	CheckName       string             `json:"check_name"`
	CheckResult     CheckovCheckResult `json:"check_result"`
	Severity        string             `json:"severity"`
	FilePathLevel   string             `json:"file_path"`
	FilePath        string             `json:"file_path_level"`
	FileLine        int                `json:"file_line_range,omitempty"`
	ResourceType    string             `json:"resource_type"`
	ResourceAddress string             `json:"resource_address"`
	Guideline       string             `json:"guideline,omitempty"`
}

// CheckovCheckResult contains the check result status
type CheckovCheckResult struct {
	Result string `json:"result"`
}

// ScanDirectory scans a directory of IaC files
func (s *CheckovScanner) ScanDirectory(ctx context.Context, path string, framework string) ([]*CheckovResult, error) {
	args := []string{
		"-d", path,
		"--output", "json",
		"--quiet",
	}

	if framework != "" {
		args = append(args, "--framework", framework)
	}

	return s.executeScan(ctx, args)
}

// ScanFile scans a single IaC file
func (s *CheckovScanner) ScanFile(ctx context.Context, filePath string, framework string) ([]*CheckovResult, error) {
	args := []string{
		"-f", filePath,
		"--output", "json",
		"--quiet",
	}

	if framework != "" {
		args = append(args, "--framework", framework)
	}

	return s.executeScan(ctx, args)
}

// ScanContent scans IaC content from memory
func (s *CheckovScanner) ScanContent(ctx context.Context, content string, framework string, filename string) ([]*CheckovResult, error) {
	// Write content to a temporary file and scan it
	args := []string{
		"--file-content", content,
		"--output", "json",
		"--quiet",
	}

	if framework != "" {
		args = append(args, "--framework", framework)
	}

	return s.executeScan(ctx, args)
}

// executeScan runs Checkov with the given arguments
func (s *CheckovScanner) executeScan(ctx context.Context, args []string) ([]*CheckovResult, error) {
	ctx, cancel := context.WithTimeout(ctx, s.scanTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, s.checkovPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	// Checkov returns non-zero exit code when findings are detected
	// We only care about actual execution errors
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			s.logger.WithFields(map[string]interface{}{
				"error":  err.Error(),
				"stderr": stderr.String(),
			}).ErrorWithErr(err, "Checkov execution failed")
			return nil, err
		}
	}

	// Parse JSON output
	var results []*CheckovResult
	if stdout.Len() > 0 {
		// Checkov might output array or single object
		if stdout.Bytes()[0] == '[' {
			if err := json.Unmarshal(stdout.Bytes(), &results); err != nil {
				// Try single object
				var single CheckovResult
				if err := json.Unmarshal(stdout.Bytes(), &single); err == nil {
					results = []*CheckovResult{&single}
				}
			}
		} else {
			var single CheckovResult
			if err := json.Unmarshal(stdout.Bytes(), &single); err == nil {
				results = []*CheckovResult{&single}
			}
		}
	}

	return results, nil
}

// ConvertToVulnerabilities converts Checkov results to vulnerability model
func (s *CheckovScanner) ConvertToVulnerabilities(
	userID int64,
	scanID int64,
	resourceID string,
	results []*CheckovResult,
) []*vulnerability.Vulnerability {
	var vulnerabilities []*vulnerability.Vulnerability
	now := time.Now()

	for _, result := range results {
		for _, check := range result.Results.FailedChecks {
			vuln := &vulnerability.Vulnerability{
				UserID:          userID,
				ScanID:          &scanID,
				ResourceID:      resourceID,
				ResourceType:    check.ResourceType,
				VulnerabilityID: check.CheckID,
				Title:           check.CheckName,
				Description:     check.CheckName,
				Severity:        mapCheckovSeverity(check.Severity),
				ScannerType:     "checkov",
				DetectionMethod: "static_analysis",
				Status:          vulnerability.StatusOpen,
				Remediation:     check.Guideline,
				DetectedAt:      now,
			}
			vulnerabilities = append(vulnerabilities, vuln)
		}
	}

	return vulnerabilities
}

// CheckInstallation checks if Checkov is installed
func (s *CheckovScanner) CheckInstallation(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, s.checkovPath, "--version")
	return cmd.Run()
}

// GetVersion returns the Checkov version
func (s *CheckovScanner) GetVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, s.checkovPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(bytes.TrimSpace(output)), nil
}

// mapCheckovSeverity maps Checkov severity to our model
func mapCheckovSeverity(severity string) string {
	switch severity {
	case "CRITICAL":
		return vulnerability.SeverityCritical
	case "HIGH":
		return vulnerability.SeverityHigh
	case "MEDIUM":
		return vulnerability.SeverityMedium
	case "LOW":
		return vulnerability.SeverityLow
	default:
		return vulnerability.SeverityInfo
	}
}
