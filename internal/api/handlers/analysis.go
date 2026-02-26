package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/pratik-mahalle/infraudit/internal/integrations"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
)

// AnalysisHandler handles AI-powered resource analysis requests
type AnalysisHandler struct {
	geminiClient *integrations.GeminiClient
	logger       *logger.Logger
}

// NewAnalysisHandler creates a new analysis handler
func NewAnalysisHandler(geminiClient *integrations.GeminiClient, log *logger.Logger) *AnalysisHandler {
	return &AnalysisHandler{
		geminiClient: geminiClient,
		logger:       log,
	}
}

// AnalyzeResource handles POST /api/v1/resources/analyze
func (h *AnalysisHandler) AnalyzeResource(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var resourceData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&resourceData); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if h.geminiClient == nil {
		// Return rule-based fallback analysis when no AI is configured
		result := generateRuleBasedAnalysis(resourceData)
		respondJSON(w, http.StatusOK, result)
		return
	}

	// Call Gemini for AI analysis
	response, err := h.geminiClient.AnalyzeResourceForRecommendations(r.Context(), resourceData)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to analyze resource with Gemini")
		// Fall back to rule-based
		result := generateRuleBasedAnalysis(resourceData)
		respondJSON(w, http.StatusOK, result)
		return
	}

	// Parse the AI response (it should be a JSON array)
	var recommendations []map[string]interface{}
	// Strip markdown code fences if present
	cleaned := strings.TrimSpace(response)
	if strings.HasPrefix(cleaned, "```") {
		lines := strings.Split(cleaned, "\n")
		if len(lines) > 2 {
			lines = lines[1 : len(lines)-1]
			cleaned = strings.Join(lines, "\n")
		}
	}

	if err := json.Unmarshal([]byte(cleaned), &recommendations); err != nil {
		h.logger.ErrorWithErr(err, "Failed to parse Gemini response, using raw text")
		// Return the raw AI text as a single recommendation
		recommendations = []map[string]interface{}{
			{
				"type":        "general",
				"priority":    "medium",
				"title":       "AI Analysis",
				"description": response,
				"category":    "general",
				"savings":     0,
				"effort":      "medium",
				"impact":      "medium",
			},
		}
	}

	// Build structured response
	result := buildAnalysisResult(resourceData, recommendations)
	respondJSON(w, http.StatusOK, result)
}

// buildAnalysisResult structures the AI recommendations into cost + security analysis
func buildAnalysisResult(resource map[string]interface{}, recommendations []map[string]interface{}) map[string]interface{} {
	costRecs := []string{}
	securityRecs := []string{}
	vulns := []string{}
	complianceImpact := []string{}
	var estimatedSavings float64
	costSeverity := "low"
	securitySeverity := "low"
	costDescription := "No significant cost optimizations identified for this resource."
	securityDescription := "No security drifts detected for this resource."
	costDetected := false
	securityDetected := false

	for _, rec := range recommendations {
		recType, _ := rec["type"].(string)
		priority, _ := rec["priority"].(string)
		title, _ := rec["title"].(string)
		description, _ := rec["description"].(string)
		text := title
		if description != "" {
			text = description
		}

		switch recType {
		case "cost_optimization":
			costDetected = true
			costRecs = append(costRecs, text)
			if savings, ok := rec["savings"].(float64); ok {
				estimatedSavings += savings
			}
			if prioritySeverity(priority) > prioritySeverity(costSeverity) {
				costSeverity = priority
				costDescription = fmt.Sprintf("Cost optimization opportunities detected: %s", text)
			}
		case "security_improvement":
			securityDetected = true
			securityRecs = append(securityRecs, text)
			if prioritySeverity(priority) > prioritySeverity(securitySeverity) {
				securitySeverity = priority
				securityDescription = fmt.Sprintf("Security issue detected: %s", text)
			}
			// Extract vulnerabilities from description
			vulns = append(vulns, title)
		case "compliance":
			securityDetected = true
			securityRecs = append(securityRecs, text)
			if cat, ok := rec["category"].(string); ok {
				complianceImpact = append(complianceImpact, strings.ToUpper(cat))
			}
		default:
			// Classify based on content
			costRecs = append(costRecs, text)
			costDetected = true
		}
	}

	return map[string]interface{}{
		"cost_analysis": map[string]interface{}{
			"detected":          costDetected,
			"description":       costDescription,
			"severity":          costSeverity,
			"recommendations":   costRecs,
			"estimated_savings": estimatedSavings,
		},
		"security_analysis": map[string]interface{}{
			"detected":          securityDetected,
			"description":       securityDescription,
			"severity":          securitySeverity,
			"vulnerabilities":   vulns,
			"recommendations":   securityRecs,
			"compliance_impact": complianceImpact,
		},
	}
}

func prioritySeverity(p string) int {
	switch p {
	case "critical":
		return 4
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}

// generateRuleBasedAnalysis creates a basic analysis without AI
func generateRuleBasedAnalysis(resource map[string]interface{}) map[string]interface{} {
	resourceType, _ := resource["type"].(string)
	metadata, _ := resource["metadata"].(map[string]interface{})
	costRecs := []string{}
	securityRecs := []string{}
	vulns := []string{}
	var estimatedSavings float64
	costSeverity := "low"
	securitySeverity := "low"
	costDetected := false
	securityDetected := false

	switch strings.ToUpper(resourceType) {
	case "EC2":
		if metadata != nil {
			if util, ok := metadata["utilizationData"].(map[string]interface{}); ok {
				if cpu, ok := util["cpu"].(map[string]interface{}); ok {
					if avg, ok := cpu["average"].(float64); ok && avg < 30 {
						costDetected = true
						costSeverity = "high"
						costRecs = append(costRecs, "Instance is underutilized. Consider downsizing to a smaller instance type.")
						if cost, ok := resource["cost"].(float64); ok {
							estimatedSavings = cost * 0.4
						}
					}
				}
			}
		}
		costRecs = append(costRecs, "Evaluate Reserved Instances or Savings Plans for long-running workloads.")
		securityRecs = append(securityRecs, "Ensure security groups follow least-privilege access principles.")

	case "S3":
		if metadata != nil {
			if public, ok := metadata["publicAccess"].(bool); ok && public {
				securityDetected = true
				securitySeverity = "critical"
				vulns = append(vulns, "Public access is enabled for this bucket")
				securityRecs = append(securityRecs, "Disable public access immediately to prevent data exposure.")
			}
		}
		costRecs = append(costRecs, "Implement lifecycle policies to transition infrequently accessed data to cheaper storage classes.")
		costDetected = true
		costSeverity = "medium"

	case "RDS":
		if metadata != nil {
			if multiAZ, ok := metadata["multiAZ"].(bool); ok && !multiAZ {
				securityDetected = true
				securitySeverity = "high"
				vulns = append(vulns, "No Multi-AZ deployment — single point of failure")
				securityRecs = append(securityRecs, "Enable Multi-AZ for high availability and disaster recovery.")
			}
		}
		costRecs = append(costRecs, "Review instance sizing based on actual CPU and connection usage patterns.")
		costDetected = true
		costSeverity = "medium"

	default:
		costRecs = append(costRecs, "Monitor resource utilization and review for optimization opportunities.")
		securityRecs = append(securityRecs, "Ensure resource follows security best practices and compliance standards.")
	}

	costDesc := "No significant cost anomalies detected."
	if costDetected {
		costDesc = fmt.Sprintf("Cost optimization opportunities found for this %s resource.", resourceType)
	}
	secDesc := "No security drifts detected."
	if securityDetected {
		secDesc = fmt.Sprintf("Security issues detected for this %s resource.", resourceType)
	}

	return map[string]interface{}{
		"cost_analysis": map[string]interface{}{
			"detected":          costDetected,
			"description":       costDesc,
			"severity":          costSeverity,
			"recommendations":   costRecs,
			"estimated_savings": estimatedSavings,
		},
		"security_analysis": map[string]interface{}{
			"detected":          securityDetected,
			"description":       secDesc,
			"severity":          securitySeverity,
			"vulnerabilities":   vulns,
			"recommendations":   securityRecs,
			"compliance_impact": []string{},
		},
	}
}
