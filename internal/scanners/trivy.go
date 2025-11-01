package scanners

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"infraaudit/backend/internal/domain/vulnerability"
	"infraaudit/backend/internal/pkg/logger"
)

// TrivyScanner wraps Trivy vulnerability scanner
type TrivyScanner struct {
	logger       *logger.Logger
	trivyPath    string
	cacheDir     string
	scanTimeout  time.Duration
}

// NewTrivyScanner creates a new Trivy scanner instance
func NewTrivyScanner(log *logger.Logger, trivyPath string, cacheDir string) *TrivyScanner {
	if trivyPath == "" {
		trivyPath = "trivy" // Use trivy from PATH
	}
	if cacheDir == "" {
		cacheDir = "/tmp/trivy-cache"
	}

	return &TrivyScanner{
		logger:      log,
		trivyPath:   trivyPath,
		cacheDir:    cacheDir,
		scanTimeout: 10 * time.Minute,
	}
}

// TrivyResult represents Trivy JSON output
type TrivyResult struct {
	SchemaVersion int                `json:"SchemaVersion"`
	ArtifactName  string             `json:"ArtifactName"`
	ArtifactType  string             `json:"ArtifactType"`
	Metadata      TrivyMetadata      `json:"Metadata"`
	Results       []TrivyVulnResult  `json:"Results"`
}

type TrivyMetadata struct {
	OS          TrivyOS            `json:"OS"`
	ImageID     string             `json:"ImageID"`
	RepoTags    []string           `json:"RepoTags"`
	RepoDigests []string           `json:"RepoDigests"`
}

type TrivyOS struct {
	Family string `json:"Family"`
	Name   string `json:"Name"`
}

type TrivyVulnResult struct {
	Target          string                `json:"Target"`
	Class           string                `json:"Class"`
	Type            string                `json:"Type"`
	Vulnerabilities []TrivyVulnerability  `json:"Vulnerabilities"`
}

type TrivyVulnerability struct {
	VulnerabilityID  string             `json:"VulnerabilityID"`
	PkgID            string             `json:"PkgID"`
	PkgName          string             `json:"PkgName"`
	InstalledVersion string             `json:"InstalledVersion"`
	FixedVersion     string             `json:"FixedVersion"`
	Severity         string             `json:"Severity"`
	Title            string             `json:"Title"`
	Description      string             `json:"Description"`
	References       []string           `json:"References"`
	PublishedDate    *time.Time         `json:"PublishedDate"`
	LastModifiedDate *time.Time         `json:"LastModifiedDate"`
	CVSS             map[string]CVSS    `json:"CVSS"`
	PrimaryURL       string             `json:"PrimaryURL"`
}

type CVSS struct {
	V2Vector string  `json:"V2Vector"`
	V3Vector string  `json:"V3Vector"`
	V2Score  float64 `json:"V2Score"`
	V3Score  float64 `json:"V3Score"`
}

// ScanTarget represents a target for Trivy scanning
type ScanTarget struct {
	Type       string // image, filesystem, rootfs, repository
	Target     string // image:tag, /path/to/dir, github.com/owner/repo
	Provider   string
	ResourceID string
}

// ScanImage scans a container image for vulnerabilities
func (ts *TrivyScanner) ScanImage(ctx context.Context, imageName string) (*TrivyResult, error) {
	ts.logger.WithFields(map[string]interface{}{
		"image": imageName,
	}).Info("Scanning container image with Trivy")

	args := []string{
		"image",
		"--format", "json",
		"--cache-dir", ts.cacheDir,
		"--timeout", ts.scanTimeout.String(),
		"--severity", "CRITICAL,HIGH,MEDIUM,LOW",
		imageName,
	}

	return ts.executeScan(ctx, args)
}

// ScanFilesystem scans a filesystem for vulnerabilities
func (ts *TrivyScanner) ScanFilesystem(ctx context.Context, path string) (*TrivyResult, error) {
	ts.logger.WithFields(map[string]interface{}{
		"path": path,
	}).Info("Scanning filesystem with Trivy")

	args := []string{
		"fs",
		"--format", "json",
		"--cache-dir", ts.cacheDir,
		"--timeout", ts.scanTimeout.String(),
		"--severity", "CRITICAL,HIGH,MEDIUM,LOW",
		path,
	}

	return ts.executeScan(ctx, args)
}

// ScanRepository scans a git repository for vulnerabilities
func (ts *TrivyScanner) ScanRepository(ctx context.Context, repoURL string) (*TrivyResult, error) {
	ts.logger.WithFields(map[string]interface{}{
		"repository": repoURL,
	}).Info("Scanning git repository with Trivy")

	args := []string{
		"repo",
		"--format", "json",
		"--cache-dir", ts.cacheDir,
		"--timeout", ts.scanTimeout.String(),
		"--severity", "CRITICAL,HIGH,MEDIUM,LOW",
		repoURL,
	}

	return ts.executeScan(ctx, args)
}

// executeScan runs Trivy with the given arguments
func (ts *TrivyScanner) executeScan(ctx context.Context, args []string) (*TrivyResult, error) {
	// Create context with timeout
	scanCtx, cancel := context.WithTimeout(ctx, ts.scanTimeout)
	defer cancel()

	// Execute Trivy command
	cmd := exec.CommandContext(scanCtx, ts.trivyPath, args...)

	ts.logger.WithFields(map[string]interface{}{
		"command": fmt.Sprintf("%s %s", ts.trivyPath, strings.Join(args, " ")),
	}).Debug("Executing Trivy scan")

	output, err := cmd.CombinedOutput()
	if err != nil {
		ts.logger.WithError(err).WithFields(map[string]interface{}{
			"output": string(output),
		}).Error("Trivy scan failed")
		return nil, fmt.Errorf("trivy scan failed: %w, output: %s", err, string(output))
	}

	// Parse JSON output
	var result TrivyResult
	if err := json.Unmarshal(output, &result); err != nil {
		ts.logger.WithError(err).Error("Failed to parse Trivy output")
		return nil, fmt.Errorf("failed to parse trivy output: %w", err)
	}

	ts.logger.WithFields(map[string]interface{}{
		"total_results": len(result.Results),
	}).Info("Trivy scan completed")

	return &result, nil
}

// ConvertToVulnerabilities converts Trivy results to our vulnerability model
func (ts *TrivyScanner) ConvertToVulnerabilities(
	userID int64,
	scanID int64,
	target ScanTarget,
	trivyResult *TrivyResult,
) []*vulnerability.Vulnerability {
	var vulnerabilities []*vulnerability.Vulnerability

	for _, result := range trivyResult.Results {
		for _, vuln := range result.Vulnerabilities {
			// Determine package type
			packageType := vulnerability.PackageTypeLibrary
			if result.Class == "os-pkgs" {
				packageType = vulnerability.PackageTypeOS
			}

			// Get CVSS score and vector
			var cvssScore *float64
			var cvssVector string
			for vendor, cvss := range vuln.CVSS {
				if cvss.V3Score > 0 {
					cvssScore = &cvss.V3Score
					cvssVector = cvss.V3Vector
					break
				} else if vendor == "nvd" && cvss.V2Score > 0 {
					cvssScore = &cvss.V2Score
					cvssVector = cvss.V2Vector
				}
			}

			// Convert severity to lowercase
			severity := strings.ToLower(vuln.Severity)
			if severity == "" {
				severity = vulnerability.SeverityInfo
			}

			// Convert references to JSON
			referencesJSON := ""
			if len(vuln.References) > 0 {
				if data, err := json.Marshal(vuln.References); err == nil {
					referencesJSON = string(data)
				}
			}

			v := &vulnerability.Vulnerability{
				UserID:           userID,
				ScanID:           &scanID,
				ResourceID:       target.ResourceID,
				Provider:         target.Provider,
				ResourceType:     result.Target,
				CVEID:            vuln.VulnerabilityID,
				VulnerabilityID:  vuln.VulnerabilityID,
				Title:            vuln.Title,
				Description:      vuln.Description,
				Severity:         severity,
				CVSSScore:        cvssScore,
				CVSSVector:       cvssVector,
				PackageName:      vuln.PkgName,
				PackageVersion:   vuln.InstalledVersion,
				FixedVersion:     vuln.FixedVersion,
				PackageType:      packageType,
				ScannerType:      vulnerability.ScanTypeTrivy,
				DetectionMethod:  fmt.Sprintf("trivy-%s", target.Type),
				Status:           vulnerability.StatusOpen,
				ReferenceURLs:    referencesJSON,
				PublishedDate:    vuln.PublishedDate,
				LastModifiedDate: vuln.LastModifiedDate,
				DetectedAt:       time.Now(),
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			}

			// Generate remediation text
			if vuln.FixedVersion != "" {
				v.Remediation = fmt.Sprintf("Update %s from version %s to %s or later",
					vuln.PkgName, vuln.InstalledVersion, vuln.FixedVersion)
			} else {
				v.Remediation = fmt.Sprintf("No fix available yet for %s %s. Monitor %s for updates",
					vuln.PkgName, vuln.InstalledVersion, vuln.PrimaryURL)
			}

			vulnerabilities = append(vulnerabilities, v)
		}
	}

	ts.logger.WithFields(map[string]interface{}{
		"total_vulnerabilities": len(vulnerabilities),
	}).Info("Converted Trivy results to vulnerabilities")

	return vulnerabilities
}

// GetVersion returns the Trivy version
func (ts *TrivyScanner) GetVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, ts.trivyPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get trivy version: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// CheckInstallation checks if Trivy is installed and accessible
func (ts *TrivyScanner) CheckInstallation(ctx context.Context) error {
	_, err := exec.LookPath(ts.trivyPath)
	if err != nil {
		return fmt.Errorf("trivy not found in PATH: %w", err)
	}

	version, err := ts.GetVersion(ctx)
	if err != nil {
		return fmt.Errorf("trivy not working properly: %w", err)
	}

	ts.logger.WithFields(map[string]interface{}{
		"version": version,
	}).Info("Trivy scanner is installed and working")

	return nil
}
