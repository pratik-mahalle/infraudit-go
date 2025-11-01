package scanners

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pratik-mahalle/infraudit/internal/domain/vulnerability"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
)

// NVDScanner integrates with the National Vulnerability Database API
type NVDScanner struct {
	logger     *logger.Logger
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewNVDScanner creates a new NVD scanner instance
func NewNVDScanner(log *logger.Logger, apiKey string) *NVDScanner {
	return &NVDScanner{
		logger:  log,
		baseURL: "https://services.nvd.nist.gov/rest/json/cves/2.0",
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NVD API Response structures
type NVDResponse struct {
	ResultsPerPage  int          `json:"resultsPerPage"`
	StartIndex      int          `json:"startIndex"`
	TotalResults    int          `json:"totalResults"`
	Format          string       `json:"format"`
	Version         string       `json:"version"`
	Timestamp       string       `json:"timestamp"`
	Vulnerabilities []NVDVulnItem `json:"vulnerabilities"`
}

type NVDVulnItem struct {
	CVE NVDCVEItem `json:"cve"`
}

type NVDCVEItem struct {
	ID               string               `json:"id"`
	SourceIdentifier string               `json:"sourceIdentifier"`
	Published        string               `json:"published"`
	LastModified     string               `json:"lastModified"`
	VulnStatus       string               `json:"vulnStatus"`
	Descriptions     []NVDDescription     `json:"descriptions"`
	Metrics          NVDMetrics           `json:"metrics"`
	Weaknesses       []NVDWeakness        `json:"weaknesses"`
	Configurations   []NVDConfiguration   `json:"configurations"`
	References       []NVDReference       `json:"references"`
}

type NVDDescription struct {
	Lang  string `json:"lang"`
	Value string `json:"value"`
}

type NVDMetrics struct {
	CVSSMetricV31 []NVDCVSSMetric `json:"cvssMetricV31"`
	CVSSMetricV30 []NVDCVSSMetric `json:"cvssMetricV30"`
	CVSSMetricV2  []NVDCVSSMetric `json:"cvssMetricV2"`
}

type NVDCVSSMetric struct {
	Source              string      `json:"source"`
	Type                string      `json:"type"`
	CVSSData            NVDCVSSData `json:"cvssData"`
	BaseSeverity        string      `json:"baseSeverity"`
	ExploitabilityScore float64     `json:"exploitabilityScore"`
	ImpactScore         float64     `json:"impactScore"`
}

type NVDCVSSData struct {
	Version               string  `json:"version"`
	VectorString          string  `json:"vectorString"`
	AttackVector          string  `json:"attackVector"`
	AttackComplexity      string  `json:"attackComplexity"`
	PrivilegesRequired    string  `json:"privilegesRequired"`
	UserInteraction       string  `json:"userInteraction"`
	Scope                 string  `json:"scope"`
	ConfidentialityImpact string  `json:"confidentialityImpact"`
	IntegrityImpact       string  `json:"integrityImpact"`
	AvailabilityImpact    string  `json:"availabilityImpact"`
	BaseScore             float64 `json:"baseScore"`
}

type NVDWeakness struct {
	Source      string              `json:"source"`
	Type        string              `json:"type"`
	Description []NVDDescription    `json:"description"`
}

type NVDConfiguration struct {
	Nodes []NVDNode `json:"nodes"`
}

type NVDNode struct {
	Operator string      `json:"operator"`
	Negate   bool        `json:"negate"`
	CPEMatch []NVDCPEMatch `json:"cpeMatch"`
}

type NVDCPEMatch struct {
	Vulnerable            bool   `json:"vulnerable"`
	Criteria              string `json:"criteria"`
	VersionStartIncluding string `json:"versionStartIncluding"`
	VersionEndExcluding   string `json:"versionEndExcluding"`
	MatchCriteriaID       string `json:"matchCriteriaId"`
}

type NVDReference struct {
	URL    string   `json:"url"`
	Source string   `json:"source"`
	Tags   []string `json:"tags"`
}

// GetCVEByID fetches a specific CVE by ID from NVD
func (ns *NVDScanner) GetCVEByID(ctx context.Context, cveID string) (*NVDCVEItem, error) {
	ns.logger.WithFields(map[string]interface{}{
		"cve_id": cveID,
	}).Info("Fetching CVE from NVD")

	// Build URL
	endpoint := fmt.Sprintf("%s?cveId=%s", ns.baseURL, url.QueryEscape(cveID))

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add API key if available
	if ns.apiKey != "" {
		req.Header.Set("apiKey", ns.apiKey)
	}

	// Execute request
	resp, err := ns.httpClient.Do(req)
	if err != nil {
		ns.logger.WithError(err).Error("Failed to fetch CVE from NVD")
		return nil, fmt.Errorf("failed to fetch CVE: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		ns.logger.WithFields(map[string]interface{}{
			"status_code": resp.StatusCode,
			"body":        string(body),
		}).Error("NVD API returned error")
		return nil, fmt.Errorf("NVD API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var nvdResp NVDResponse
	if err := json.NewDecoder(resp.Body).Decode(&nvdResp); err != nil {
		return nil, fmt.Errorf("failed to decode NVD response: %w", err)
	}

	if len(nvdResp.Vulnerabilities) == 0 {
		return nil, fmt.Errorf("CVE %s not found in NVD", cveID)
	}

	ns.logger.WithFields(map[string]interface{}{
		"cve_id": cveID,
	}).Info("Successfully fetched CVE from NVD")

	return &nvdResp.Vulnerabilities[0].CVE, nil
}

// SearchCVEs searches for CVEs based on keywords
func (ns *NVDScanner) SearchCVEs(ctx context.Context, keyword string, limit int) ([]NVDCVEItem, error) {
	ns.logger.WithFields(map[string]interface{}{
		"keyword": keyword,
		"limit":   limit,
	}).Info("Searching CVEs in NVD")

	// Build URL
	endpoint := fmt.Sprintf("%s?keywordSearch=%s&resultsPerPage=%d",
		ns.baseURL, url.QueryEscape(keyword), limit)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add API key if available
	if ns.apiKey != "" {
		req.Header.Set("apiKey", ns.apiKey)
	}

	// Execute request
	resp, err := ns.httpClient.Do(req)
	if err != nil {
		ns.logger.WithError(err).Error("Failed to search CVEs in NVD")
		return nil, fmt.Errorf("failed to search CVEs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("NVD API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var nvdResp NVDResponse
	if err := json.NewDecoder(resp.Body).Decode(&nvdResp); err != nil {
		return nil, fmt.Errorf("failed to decode NVD response: %w", err)
	}

	cves := make([]NVDCVEItem, len(nvdResp.Vulnerabilities))
	for i, v := range nvdResp.Vulnerabilities {
		cves[i] = v.CVE
	}

	ns.logger.WithFields(map[string]interface{}{
		"keyword": keyword,
		"found":   len(cves),
	}).Info("CVE search completed")

	return cves, nil
}

// ConvertToVulnerability converts NVD CVE item to our vulnerability model
func (ns *NVDScanner) ConvertToVulnerability(
	userID int64,
	scanID *int64,
	resourceID string,
	provider string,
	nvdCVE *NVDCVEItem,
) *vulnerability.Vulnerability {
	// Get English description
	description := ""
	for _, desc := range nvdCVE.Descriptions {
		if desc.Lang == "en" {
			description = desc.Value
			break
		}
	}

	// Get CVSS data (prefer v3.1, fallback to v3.0, then v2)
	var cvssScore *float64
	var cvssVector string
	var severity string

	if len(nvdCVE.Metrics.CVSSMetricV31) > 0 {
		metric := nvdCVE.Metrics.CVSSMetricV31[0]
		cvssScore = &metric.CVSSData.BaseScore
		cvssVector = metric.CVSSData.VectorString
		severity = strings.ToLower(metric.BaseSeverity)
	} else if len(nvdCVE.Metrics.CVSSMetricV30) > 0 {
		metric := nvdCVE.Metrics.CVSSMetricV30[0]
		cvssScore = &metric.CVSSData.BaseScore
		cvssVector = metric.CVSSData.VectorString
		severity = strings.ToLower(metric.BaseSeverity)
	} else if len(nvdCVE.Metrics.CVSSMetricV2) > 0 {
		metric := nvdCVE.Metrics.CVSSMetricV2[0]
		cvssScore = &metric.CVSSData.BaseScore
		cvssVector = metric.CVSSData.VectorString
		// V2 doesn't have baseSeverity, calculate from score
		severity = ns.scoresToSeverity(*cvssScore)
	}

	// Default severity if not available
	if severity == "" {
		if cvssScore != nil {
			severity = ns.scoresToSeverity(*cvssScore)
		} else {
			severity = vulnerability.SeverityInfo
		}
	}

	// Parse timestamps
	publishedDate, _ := time.Parse(time.RFC3339, nvdCVE.Published)
	lastModifiedDate, _ := time.Parse(time.RFC3339, nvdCVE.LastModified)

	// Convert references to JSON
	referencesJSON := ""
	if len(nvdCVE.References) > 0 {
		refs := make([]string, len(nvdCVE.References))
		for i, ref := range nvdCVE.References {
			refs[i] = ref.URL
		}
		if data, err := json.Marshal(refs); err == nil {
			referencesJSON = string(data)
		}
	}

	vuln := &vulnerability.Vulnerability{
		UserID:           userID,
		ScanID:           scanID,
		ResourceID:       resourceID,
		Provider:         provider,
		CVEID:            nvdCVE.ID,
		VulnerabilityID:  nvdCVE.ID,
		Title:            fmt.Sprintf("Vulnerability: %s", nvdCVE.ID),
		Description:      description,
		Severity:         severity,
		CVSSScore:        cvssScore,
		CVSSVector:       cvssVector,
		ScannerType:      vulnerability.ScanTypeNVD,
		DetectionMethod:  "nvd-api",
		Status:           vulnerability.StatusOpen,
		ReferenceURLs:    referencesJSON,
		PublishedDate:    &publishedDate,
		LastModifiedDate: &lastModifiedDate,
		DetectedAt:       time.Now(),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Extract affected packages from CPE configurations
	if len(nvdCVE.Configurations) > 0 {
		for _, config := range nvdCVE.Configurations {
			for _, node := range config.Nodes {
				for _, cpe := range node.CPEMatch {
					if cpe.Vulnerable {
						// Parse CPE string (e.g., cpe:2.3:a:vendor:product:version:...)
						parts := strings.Split(cpe.Criteria, ":")
						if len(parts) >= 5 {
							vuln.PackageName = parts[4] // product name
							if len(parts) >= 6 && parts[5] != "*" {
								vuln.PackageVersion = parts[5]
							}
							if cpe.VersionEndExcluding != "" {
								vuln.FixedVersion = cpe.VersionEndExcluding
							}
							break
						}
					}
				}
			}
		}
	}

	return vuln
}

// scoresToSeverity converts CVSS score to severity level
func (ns *NVDScanner) scoresToSeverity(score float64) string {
	if score >= 9.0 {
		return vulnerability.SeverityCritical
	} else if score >= 7.0 {
		return vulnerability.SeverityHigh
	} else if score >= 4.0 {
		return vulnerability.SeverityMedium
	} else if score > 0.0 {
		return vulnerability.SeverityLow
	}
	return vulnerability.SeverityInfo
}
