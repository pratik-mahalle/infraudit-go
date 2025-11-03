package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pratik-mahalle/infraudit/internal/domain/drift"
	"github.com/pratik-mahalle/infraudit/internal/domain/recommendation"
	"github.com/pratik-mahalle/infraudit/internal/domain/resource"
	"github.com/pratik-mahalle/infraudit/internal/domain/vulnerability"
	"github.com/pratik-mahalle/infraudit/internal/integrations"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
)

// RecommendationEngine generates intelligent recommendations using AI
type RecommendationEngine struct {
	geminiClient *integrations.GeminiClient
	resourceRepo resource.Repository
	vulnRepo     vulnerability.Repository
	driftRepo    drift.Repository
	recRepo      recommendation.Repository
	logger       *logger.Logger
}

// NewRecommendationEngine creates a new recommendation engine
func NewRecommendationEngine(
	geminiClient *integrations.GeminiClient,
	resourceRepo resource.Repository,
	vulnRepo vulnerability.Repository,
	driftRepo drift.Repository,
	recRepo recommendation.Repository,
	logger *logger.Logger,
) *RecommendationEngine {
	return &RecommendationEngine{
		geminiClient: geminiClient,
		resourceRepo: resourceRepo,
		vulnRepo:     vulnRepo,
		driftRepo:    driftRepo,
		recRepo:      recRepo,
		logger:       logger,
	}
}

// GenerateRecommendations generates all types of recommendations for a user
func (e *RecommendationEngine) GenerateRecommendations(ctx context.Context, userID int64) error {
	e.logger.WithFields(map[string]interface{}{
		"user_id": userID,
	}).Info("Starting recommendation generation")

	// Generate cost optimization recommendations
	if err := e.generateCostOptimizationRecommendations(ctx, userID); err != nil {
		e.logger.ErrorWithErr(err, "Failed to generate cost optimization recommendations")
		// Continue with other recommendations even if this fails
	}

	// Generate security recommendations from vulnerabilities
	if err := e.generateSecurityRecommendations(ctx, userID); err != nil {
		e.logger.ErrorWithErr(err, "Failed to generate security recommendations")
		// Continue with other recommendations
	}

	// Generate compliance and drift recommendations
	if err := e.generateComplianceRecommendations(ctx, userID); err != nil {
		e.logger.ErrorWithErr(err, "Failed to generate compliance recommendations")
		// Continue
	}

	e.logger.WithFields(map[string]interface{}{
		"user_id": userID,
	}).Info("Completed recommendation generation")

	return nil
}

// generateCostOptimizationRecommendations analyzes resources for cost optimization
func (e *RecommendationEngine) generateCostOptimizationRecommendations(ctx context.Context, userID int64) error {
	e.logger.Info("Generating cost optimization recommendations")

	// Process all resources with pagination
	pageSize := 1000
	offset := 0
	batchSize := 10
	allResources := make([]*resource.Resource, 0)

	for {
		// Get page of resources
		resources, total, err := e.resourceRepo.List(ctx, userID, resource.Filter{}, pageSize, offset)
		if err != nil {
			return fmt.Errorf("failed to list resources: %w", err)
		}

		if len(resources) == 0 {
			break
		}

		allResources = append(allResources, resources...)

		// Check if we've fetched all resources
		if int64(offset+len(resources)) >= total {
			break
		}

		offset += len(resources)
	}

	if len(allResources) == 0 {
		e.logger.Info("No resources found, skipping cost optimization recommendations")
		return nil
	}

	e.logger.WithFields(map[string]interface{}{
		"total_resources": len(allResources),
	}).Info("Processing resources for cost optimization")

	// Analyze resources in batches
	for i := 0; i < len(allResources); i += batchSize {
		end := i + batchSize
		if end > len(allResources) {
			end = len(allResources)
		}

		batch := allResources[i:end]
		if err := e.analyzeCostOptimizationBatch(ctx, userID, batch); err != nil {
			e.logger.ErrorWithErr(err, "Failed to analyze cost optimization batch")
			continue
		}
	}

	return nil
}

// analyzeCostOptimizationBatch analyzes a batch of resources for cost optimization
func (e *RecommendationEngine) analyzeCostOptimizationBatch(ctx context.Context, userID int64, resources []*resource.Resource) error {
	// Prepare resource data for AI analysis
	resourceData := make([]map[string]interface{}, 0, len(resources))
	for _, res := range resources {
		data := map[string]interface{}{
			"id":            res.ResourceID,
			"name":          res.Name,
			"type":          res.Type,
			"provider":      res.Provider,
			"region":        res.Region,
			"status":        res.Status,
			"configuration": res.Configuration,
		}
		resourceData = append(resourceData, data)
	}

	// Batch analyze with Gemini
	jsonData, err := json.MarshalIndent(map[string]interface{}{
		"resources": resourceData,
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal resources: %w", err)
	}

	prompt := fmt.Sprintf(`Analyze these cloud infrastructure resources for cost optimization opportunities:

Resources:
%s

Focus on:
1. **Rightsizing**: Instances that are oversized or underutilized
2. **Unused Resources**: Resources that are stopped, idle, or have no activity
3. **Storage Optimization**: Unattached volumes, old snapshots, inefficient storage tiers
4. **Reserved Instances**: Long-running resources that could benefit from reserved pricing

Provide recommendations in JSON format:
{
  "type": "cost_optimization",
  "priority": "critical|high|medium|low",
  "title": "Brief actionable title",
  "description": "Detailed issue and recommendation with specific metrics",
  "category": "rightsizing|unused_resources|storage_optimization|reserved_instances",
  "savings": estimated_monthly_savings_in_dollars,
  "effort": "low|medium|high",
  "impact": "high|medium|low",
  "resources": ["resource_id"]
}

Return ONLY a JSON array of recommendations. Be specific with savings estimates.`, string(jsonData))

	response, err := e.geminiClient.GenerateContent(ctx, prompt)
	if err != nil {
		return fmt.Errorf("failed to generate cost recommendations: %w", err)
	}

	// Parse and save recommendations
	return e.parseAndSaveRecommendations(ctx, userID, response)
}

// generateSecurityRecommendations analyzes vulnerabilities for security improvements
func (e *RecommendationEngine) generateSecurityRecommendations(ctx context.Context, userID int64) error {
	e.logger.Info("Generating security recommendations")

	// Process all vulnerabilities with pagination
	pageSize := 100
	offset := 0
	allVulns := make([]*vulnerability.Vulnerability, 0)

	for {
		// Get page of vulnerabilities
		vulns, total, err := e.vulnRepo.ListWithPagination(ctx, userID, vulnerability.Filter{
			Status: "open",
		}, pageSize, offset)
		if err != nil {
			return fmt.Errorf("failed to list vulnerabilities: %w", err)
		}

		if len(vulns) == 0 {
			break
		}

		allVulns = append(allVulns, vulns...)

		// Check if we've fetched all vulnerabilities
		if int64(offset+len(vulns)) >= total {
			break
		}

		offset += len(vulns)
	}

	if len(allVulns) == 0 {
		e.logger.Info("No open vulnerabilities found, skipping vulnerability-based security recommendations")
		// Still generate security recommendations from resource configurations
		return e.generateResourceSecurityRecommendations(ctx, userID)
	}

	e.logger.WithFields(map[string]interface{}{
		"total_vulnerabilities": len(allVulns),
	}).Info("Processing vulnerabilities for security recommendations")

	// Prepare vulnerability data
	vulnData := make([]map[string]interface{}, 0, len(allVulns))
	for _, v := range allVulns {
		data := map[string]interface{}{
			"cve_id":          v.CVEID,
			"title":           v.Title,
			"severity":        v.Severity,
			"cvss_score":      v.CVSSScore,
			"package_name":    v.PackageName,
			"package_version": v.PackageVersion,
			"fixed_version":   v.FixedVersion,
			"resource_id":     v.ResourceID,
			"resource_type":   v.ResourceType,
		}
		vulnData = append(vulnData, data)
	}

	response, err := e.geminiClient.AnalyzeVulnerabilitiesForRecommendations(ctx, vulnData)
	if err != nil {
		return fmt.Errorf("failed to analyze vulnerabilities: %w", err)
	}

	// Parse and save recommendations
	if err := e.parseAndSaveRecommendations(ctx, userID, response); err != nil {
		return err
	}

	// Also generate security recommendations from resource configurations
	return e.generateResourceSecurityRecommendations(ctx, userID)
}

// generateResourceSecurityRecommendations analyzes resource configurations for security
func (e *RecommendationEngine) generateResourceSecurityRecommendations(ctx context.Context, userID int64) error {
	e.logger.Info("Generating resource security recommendations")

	// Process all resources with pagination
	pageSize := 1000
	offset := 0
	allResources := make([]*resource.Resource, 0)

	for {
		// Get page of resources
		resources, total, err := e.resourceRepo.List(ctx, userID, resource.Filter{}, pageSize, offset)
		if err != nil {
			return fmt.Errorf("failed to list resources: %w", err)
		}

		if len(resources) == 0 {
			break
		}

		allResources = append(allResources, resources...)

		// Check if we've fetched all resources
		if int64(offset+len(resources)) >= total {
			break
		}

		offset += len(resources)
	}

	if len(allResources) == 0 {
		return nil
	}

	e.logger.WithFields(map[string]interface{}{
		"total_resources": len(allResources),
	}).Info("Processing resources for security recommendations")

	// Analyze each resource for security issues
	for _, res := range allResources {
		resourceData := map[string]interface{}{
			"id":            res.ResourceID,
			"name":          res.Name,
			"type":          res.Type,
			"provider":      res.Provider,
			"configuration": res.Configuration,
		}

		jsonData, err := json.MarshalIndent(resourceData, "", "  ")
		if err != nil {
			continue
		}

		prompt := fmt.Sprintf(`Analyze this cloud resource for security improvements:

Resource:
%s

Focus on:
1. **Encryption**: Data at rest and in transit encryption status
2. **Access Controls**: IAM policies, security groups, public access
3. **Network Security**: Overly permissive rules, public exposure
4. **Monitoring**: Logging and auditing configuration

Provide recommendations in JSON format:
{
  "type": "security_improvement",
  "priority": "critical|high|medium|low",
  "title": "Brief actionable title",
  "description": "Detailed security issue and remediation steps",
  "category": "encryption|access_control|network_security|monitoring",
  "savings": 0,
  "effort": "low|medium|high",
  "impact": "high|medium|low",
  "resources": ["%s"]
}

Return ONLY a JSON array of recommendations. Only include actual security issues found.`, string(jsonData), res.ResourceID)

		response, err := e.geminiClient.GenerateContent(ctx, prompt)
		if err != nil {
			e.logger.ErrorWithErr(err, "Failed to analyze resource security")
			continue
		}

		if err := e.parseAndSaveRecommendations(ctx, userID, response); err != nil {
			e.logger.ErrorWithErr(err, "Failed to save security recommendations")
			continue
		}
	}

	return nil
}

// generateComplianceRecommendations analyzes drifts for compliance issues
func (e *RecommendationEngine) generateComplianceRecommendations(ctx context.Context, userID int64) error {
	e.logger.Info("Generating compliance recommendations")

	// Process all drifts with pagination
	pageSize := 100
	offset := 0
	allDrifts := make([]*drift.Drift, 0)

	for {
		// Get page of drifts
		drifts, total, err := e.driftRepo.ListWithPagination(ctx, userID, drift.Filter{
			Status: "unresolved",
		}, pageSize, offset)
		if err != nil {
			return fmt.Errorf("failed to list drifts: %w", err)
		}

		if len(drifts) == 0 {
			break
		}

		allDrifts = append(allDrifts, drifts...)

		// Check if we've fetched all drifts
		if int64(offset+len(drifts)) >= total {
			break
		}

		offset += len(drifts)
	}

	if len(allDrifts) == 0 {
		e.logger.Info("No unresolved drifts found, skipping compliance recommendations")
		return nil
	}

	e.logger.WithFields(map[string]interface{}{
		"total_drifts": len(allDrifts),
	}).Info("Processing drifts for compliance recommendations")

	// Prepare drift data
	driftData := make([]map[string]interface{}, 0, len(allDrifts))
	for _, d := range allDrifts {
		data := map[string]interface{}{
			"resource_id": d.ResourceID,
			"drift_type":  d.DriftType,
			"severity":    d.Severity,
			"details":     d.Details,
			"detected_at": d.DetectedAt,
		}
		driftData = append(driftData, data)
	}

	response, err := e.geminiClient.AnalyzeDriftsForRecommendations(ctx, driftData)
	if err != nil {
		return fmt.Errorf("failed to analyze drifts: %w", err)
	}

	return e.parseAndSaveRecommendations(ctx, userID, response)
}

// parseAndSaveRecommendations parses AI response and saves recommendations
func (e *RecommendationEngine) parseAndSaveRecommendations(ctx context.Context, userID int64, response string) error {
	// Clean the response - remove markdown code blocks if present
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	// Parse JSON response
	var recommendations []struct {
		Type        string   `json:"type"`
		Priority    string   `json:"priority"`
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Category    string   `json:"category"`
		Savings     float64  `json:"savings"`
		Effort      string   `json:"effort"`
		Impact      string   `json:"impact"`
		Resources   []string `json:"resources"`
	}

	if err := json.Unmarshal([]byte(response), &recommendations); err != nil {
		e.logger.WithFields(map[string]interface{}{
			"response": response[:min(len(response), 500)],
		}).ErrorWithErr(err, "Failed to parse recommendations JSON")
		return fmt.Errorf("failed to parse recommendations: %w", err)
	}

	// Save each recommendation
	for _, rec := range recommendations {
		recommendation := &recommendation.Recommendation{
			UserID:      userID,
			Type:        rec.Type,
			Priority:    rec.Priority,
			Title:       rec.Title,
			Description: rec.Description,
			Category:    rec.Category,
			Savings:     rec.Savings,
			Effort:      rec.Effort,
			Impact:      rec.Impact,
			Resources:   rec.Resources,
		}

		id, err := e.recRepo.Create(ctx, recommendation)
		if err != nil {
			e.logger.ErrorWithErr(err, "Failed to save recommendation")
			continue
		}

		e.logger.WithFields(map[string]interface{}{
			"recommendation_id": id,
			"type":              rec.Type,
			"priority":          rec.Priority,
			"title":             rec.Title,
		}).Info("Saved recommendation")
	}

	return nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
