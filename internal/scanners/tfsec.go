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

// TfsecScanner wraps the tfsec Terraform security scanner
type TfsecScanner struct {
	logger      *logger.Logger
	tfsecPath   string
	scanTimeout time.Duration
}

// NewTfsecScanner creates a new tfsec scanner instance
func NewTfsecScanner(log *logger.Logger, tfsecPath string) *TfsecScanner {
	if tfsecPath == "" {
		tfsecPath = "tfsec"
	}
	return &TfsecScanner{
		logger:      log,
		tfsecPath:   tfsecPath,
		scanTimeout: 5 * time.Minute,
	}
}

// TfsecResult represents tfsec JSON output
type TfsecResult struct {
	Results []TfsecFinding `json:"results"`
}

// TfsecFinding represents a single tfsec finding
type TfsecFinding struct {
	RuleID          string        `json:"rule_id"`
	LongID          string        `json:"long_id"`
	RuleDescription string        `json:"rule_description"`
	RuleProvider    string        `json:"rule_provider"`
	RuleService     string        `json:"rule_service"`
	Link            string        `json:"link"`
	Location        TfsecLocation `json:"location"`
	Description     string        `json:"description"`
	Impact          string        `json:"impact"`
	Resolution      string        `json:"resolution"`
	Severity        string        `json:"severity"`
	Passed          bool          `json:"passed"`
	Resource        string        `json:"resource"`
}

// TfsecLocation contains file location information
type TfsecLocation struct {
	Filename  string `json:"filename"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

// ScanDirectory scans a directory of Terraform files
func (s *TfsecScanner) ScanDirectory(ctx context.Context, path string) (*TfsecResult, error) {
	args := []string{
		path,
		"--format", "json",
		"--no-colour",
	}

	return s.executeScan(ctx, args)
}

// ScanFile scans a single Terraform file
func (s *TfsecScanner) ScanFile(ctx context.Context, filePath string) (*TfsecResult, error) {
	args := []string{
		filePath,
		"--format", "json",
		"--no-colour",
	}

	return s.executeScan(ctx, args)
}

// executeScan runs tfsec with the given arguments
func (s *TfsecScanner) executeScan(ctx context.Context, args []string) (*TfsecResult, error) {
	ctx, cancel := context.WithTimeout(ctx, s.scanTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, s.tfsecPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	// tfsec returns non-zero exit code when findings are detected
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			s.logger.WithFields(map[string]interface{}{
				"error":  err.Error(),
				"stderr": stderr.String(),
			}).ErrorWithErr(err, "tfsec execution failed")
			return nil, err
		}
	}

	// Parse JSON output
	var result TfsecResult
	if stdout.Len() > 0 {
		if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
			s.logger.ErrorWithErr(err, "Failed to parse tfsec output")
			return nil, err
		}
	}

	return &result, nil
}

// ConvertToVulnerabilities converts tfsec results to vulnerability model
func (s *TfsecScanner) ConvertToVulnerabilities(
	userID int64,
	scanID int64,
	resourceID string,
	result *TfsecResult,
) []*vulnerability.Vulnerability {
	var vulnerabilities []*vulnerability.Vulnerability
	now := time.Now()

	for _, finding := range result.Results {
		if finding.Passed {
			continue // Skip passed checks
		}

		vuln := &vulnerability.Vulnerability{
			UserID:          userID,
			ScanID:          &scanID,
			ResourceID:      resourceID,
			ResourceType:    finding.RuleService,
			Provider:        finding.RuleProvider,
			VulnerabilityID: finding.RuleID,
			Title:           finding.RuleDescription,
			Description:     finding.Description,
			Severity:        mapTfsecSeverity(finding.Severity),
			ScannerType:     "tfsec",
			DetectionMethod: "static_analysis",
			Status:          vulnerability.StatusOpen,
			Remediation:     finding.Resolution,
			ReferenceURLs:   finding.Link,
			DetectedAt:      now,
		}
		vulnerabilities = append(vulnerabilities, vuln)
	}

	return vulnerabilities
}

// CheckInstallation checks if tfsec is installed
func (s *TfsecScanner) CheckInstallation(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, s.tfsecPath, "--version")
	return cmd.Run()
}

// GetVersion returns the tfsec version
func (s *TfsecScanner) GetVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, s.tfsecPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(bytes.TrimSpace(output)), nil
}

// mapTfsecSeverity maps tfsec severity to our model
func mapTfsecSeverity(severity string) string {
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
