package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/pratik-mahalle/infraudit/internal/domain/drift"
	"github.com/pratik-mahalle/infraudit/internal/domain/recommendation"
	"github.com/pratik-mahalle/infraudit/internal/domain/resource"
	"github.com/pratik-mahalle/infraudit/internal/domain/vulnerability"
	"github.com/pratik-mahalle/infraudit/internal/integrations"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
	"golang.org/x/time/rate"
)

// RecommendationEngine generates intelligent recommendations using AI
type RecommendationEngine struct {
	geminiClient *integrations.GeminiClient
	resourceRepo resource.Repository
	vulnRepo     vulnerability.Repository
	driftRepo    drift.Repository
	recRepo      recommendation.Repository
	logger       *logger.Logger
	rateLimiter  *rate.Limiter
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
	// Rate limit: 10 requests per second with burst of 20
	// This prevents overwhelming the AI API
	limiter := rate.NewLimiter(10, 20)
	
	return &RecommendationEngine{
		geminiClient: geminiClient,
		resourceRepo: resourceRepo,
		vulnRepo:     vulnRepo,
		driftRepo:    driftRepo,
		recRepo:      recRepo,
		logger:       logger,
		rateLimiter:  limiter,
	}
}

// processBatchesConcurrently processes batches concurrently with rate limiting
func (e *RecommendationEngine) processBatchesConcurrently(
	ctx context.Context,
	userID int64,
	batches [][]interface{},
	processFn func(context.Context, int64, []interface{}) error,
) error {
	const maxConcurrency = 3 // Process 3 batches concurrently
	
	var wg sync.WaitGroup
	errChan := make(chan error, len(batches))
	semaphore := make(chan struct{}, maxConcurrency)

	for _, batch := range batches {
		wg.Add(1)
		go func(b []interface{}) {
			defer wg.Done()
			
			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			// Rate limit API calls
			if err := e.rateLimiter.Wait(ctx); err != nil {
				errChan <- err
				return
			}
			
			if err := processFn(ctx, userID, b); err != nil {
				// Send error to channel for collection
				errChan <- err
				// Also log for monitoring
				e.logger.ErrorWithErr(err, "Failed to process batch")
			}
		}(batch)
	}

	wg.Wait()
	close(errChan)

	// Collect errors (if any)
	var firstErr error
	for err := range errChan {
		if firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
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

	// Process resources in streaming fashion without loading all into memory
	const pageSize = 100
	const batchSize = 10
	offset := 0
	totalProcessed := 0

	for {
		// Get page of resources
		resources, total, err := e.resourceRepo.List(ctx, userID, resource.Filter{}, pageSize, offset)
		if err != nil {
			return fmt.Errorf("failed to list resources: %w", err)
		}

		if len(resources) == 0 {
			break
		}

		totalProcessed += len(resources)

		// Analyze resources in batches within this page
		for i := 0; i < len(resources); i += batchSize {
			end := i + batchSize
			if end > len(resources) {
				end = len(resources)
			}

			batch := resources[i:end]
			if err := e.analyzeCostOptimizationBatch(ctx, userID, batch); err != nil {
				e.logger.ErrorWithErr(err, "Failed to analyze cost optimization batch")
				continue
			}
		}

		// Check if we've fetched all resources
		if int64(offset+len(resources)) >= total {
			break
		}

		offset += len(resources)
	}

	if totalProcessed == 0 {
		e.logger.Info("No resources found, skipping cost optimization recommendations")
		return nil
	}

	e.logger.WithFields(map[string]interface{}{
		"total_resources": totalProcessed,
	}).Info("Processing resources for cost optimization")

	return nil
}

// analyzeCostOptimizationBatch analyzes a batch of resources for cost optimization
func (e *RecommendationEngine) analyzeCostOptimizationBatch(ctx context.Context, userID int64, resources []*resource.Resource) error {
	// Rule-based fallback when Gemini is not available
	if e.geminiClient == nil {
		return e.analyzeCostOptimizationBatchRuleBased(ctx, userID, resources)
	}

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
	jsonData, err := json.Marshal(map[string]interface{}{
		"resources": resourceData,
	})
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

	return e.parseAndSaveRecommendations(ctx, userID, response)
}

// analyzeCostOptimizationBatchRuleBased generates cost recommendations using heuristics
func (e *RecommendationEngine) analyzeCostOptimizationBatchRuleBased(ctx context.Context, userID int64, resources []*resource.Resource) error {
	var errs []error
	for _, res := range resources {
		// Check for stopped/inactive resources
		if res.Status == "stopped" || res.Status == "inactive" || res.Status == "terminated" {
			rec := &recommendation.Recommendation{
				UserID:      userID,
				Type:        recommendation.TypeCostOptimization,
				Priority:    recommendation.PriorityMedium,
				Title:       fmt.Sprintf("Remove or downsize stopped resource: %s", res.Name),
				Description: fmt.Sprintf("Resource '%s' (%s on %s in %s) is currently %s. Consider removing it or switching to a smaller instance type to save costs.", res.Name, res.Type, res.Provider, res.Region, res.Status),
				Category:    "unused_resources",
				Savings:     10.0,
				Effort:      recommendation.EffortLow,
				Impact:      recommendation.ImpactMedium,
				Resources:   []string{res.ResourceID},
			}
			if _, err := e.recRepo.Create(ctx, rec); err != nil {
				e.logger.ErrorWithErr(err, "Failed to save cost recommendation")
				errs = append(errs, fmt.Errorf("failed to save cost recommendation for %s: %w", res.ResourceID, err))
			}
			continue
		}

		// Recommend reserved instances for running resources
		if res.Status == "running" || res.Status == "active" {
			rec := &recommendation.Recommendation{
				UserID:      userID,
				Type:        recommendation.TypeCostOptimization,
				Priority:    recommendation.PriorityLow,
				Title:       fmt.Sprintf("Evaluate reserved pricing for: %s", res.Name),
				Description: fmt.Sprintf("Resource '%s' (%s on %s) is actively running. If this is a long-running workload, consider switching to reserved instances or savings plans for up to 40%% cost reduction.", res.Name, res.Type, res.Provider),
				Category:    "reserved_instances",
				Savings:     15.0,
				Effort:      recommendation.EffortMedium,
				Impact:      recommendation.ImpactHigh,
				Resources:   []string{res.ResourceID},
			}
			if _, err := e.recRepo.Create(ctx, rec); err != nil {
				e.logger.ErrorWithErr(err, "Failed to save cost recommendation")
				errs = append(errs, fmt.Errorf("failed to save cost recommendation for %s: %w", res.ResourceID, err))
			}
		}
	}
	
	if len(errs) > 0 {
		return fmt.Errorf("encountered %d errors while creating cost recommendations: %v", len(errs), errs[0])
	}
	return nil
}

// generateSecurityRecommendations analyzes vulnerabilities for security improvements
func (e *RecommendationEngine) generateSecurityRecommendations(ctx context.Context, userID int64) error {
	e.logger.Info("Generating security recommendations")

	const (
		pageSize  = 100
		batchSize = 20
	)
	offset := 0
	totalProcessed := 0

	for {
		vulns, total, err := e.vulnRepo.ListWithPagination(ctx, userID, vulnerability.Filter{
			Status: "open",
		}, pageSize, offset)
		if err != nil {
			return fmt.Errorf("failed to list vulnerabilities: %w", err)
		}

		if len(vulns) == 0 {
			break
		}

		totalProcessed += len(vulns)

		// Rule-based fallback when Gemini is not available
		if e.geminiClient == nil {
			var errs []error
			for _, v := range vulns {
				priority := recommendation.PriorityMedium
				if v.Severity == "critical" {
					priority = recommendation.PriorityCritical
				} else if v.Severity == "high" {
					priority = recommendation.PriorityHigh
				}

				title := fmt.Sprintf("Fix %s vulnerability: %s", v.Severity, v.Title)
				desc := fmt.Sprintf("Vulnerability '%s' (severity: %s) detected on resource %s.", v.Title, v.Severity, v.ResourceID)
				if v.CVEID != "" {
					desc += fmt.Sprintf(" CVE: %s.", v.CVEID)
				}
				if v.FixedVersion != "" {
					desc += fmt.Sprintf(" Update %s from %s to %s to remediate.", v.PackageName, v.PackageVersion, v.FixedVersion)
				}

				rec := &recommendation.Recommendation{
					UserID:      userID,
					Type:        recommendation.TypeSecurityImprovement,
					Priority:    priority,
					Title:       title,
					Description: desc,
					Category:    "vulnerability_remediation",
					Savings:     0,
					Effort:      recommendation.EffortMedium,
					Impact:      recommendation.ImpactHigh,
					Resources:   []string{v.ResourceID},
				}
				if _, err := e.recRepo.Create(ctx, rec); err != nil {
					e.logger.ErrorWithErr(err, "Failed to save security recommendation")
					errs = append(errs, fmt.Errorf("failed to save security recommendation for %s: %w", v.ResourceID, err))
				}
			}
			if len(errs) > 0 {
				return fmt.Errorf("encountered %d errors while creating security recommendations: %v", len(errs), errs[0])
			}
		} else {
			// AI-powered analysis
			for i := 0; i < len(vulns); i += batchSize {
				end := i + batchSize
				if end > len(vulns) {
					end = len(vulns)
				}

				batch := vulns[i:end]

				vulnData := make([]map[string]interface{}, 0, len(batch))
				for _, v := range batch {
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
					e.logger.ErrorWithErr(err, "Failed to analyze vulnerability batch")
					continue
				}

				if err := e.parseAndSaveRecommendations(ctx, userID, response); err != nil {
					e.logger.ErrorWithErr(err, "Failed to save vulnerability recommendations")
					continue
				}
			}
		}

		if int64(offset+len(vulns)) >= total {
			break
		}

		offset += len(vulns)
	}

	if totalProcessed == 0 {
		e.logger.Info("No open vulnerabilities found, skipping vulnerability-based security recommendations")
	}

	// Also generate security recommendations from resource configurations (only with AI)
	if e.geminiClient != nil {
		return e.generateResourceSecurityRecommendations(ctx, userID)
	}
	return nil
}

// generateResourceSecurityRecommendations analyzes resource configurations for security
func (e *RecommendationEngine) generateResourceSecurityRecommendations(ctx context.Context, userID int64) error {
	e.logger.Info("Generating resource security recommendations")

	const (
		pageSize  = 100
		batchSize = 10
	)
	offset := 0
	totalProcessed := 0

	for {
		resources, total, err := e.resourceRepo.List(ctx, userID, resource.Filter{}, pageSize, offset)
		if err != nil {
			return fmt.Errorf("failed to list resources: %w", err)
		}

		if len(resources) == 0 {
			break
		}

		totalProcessed += len(resources)

		for i := 0; i < len(resources); i += batchSize {
			end := i + batchSize
			if end > len(resources) {
				end = len(resources)
			}

			batch := resources[i:end]

			resourceData := make([]map[string]interface{}, 0, len(batch))
			for _, res := range batch {
				resourceData = append(resourceData, map[string]interface{}{
					"id":            res.ResourceID,
					"name":          res.Name,
					"type":          res.Type,
					"provider":      res.Provider,
					"configuration": res.Configuration,
				})
			}

			// Use Marshal instead of MarshalIndent for better performance
			jsonData, err := json.Marshal(map[string]interface{}{
				"resources": resourceData,
			})
			if err != nil {
				e.logger.ErrorWithErr(err, "Failed to marshal resources for security analysis")
				continue
			}

			prompt := fmt.Sprintf(`Analyze these cloud resources for security improvements:

Resources:
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
  "resources": ["resource_id"]
}

Return ONLY a JSON array of recommendations. Only include actual security issues found.`, string(jsonData))

			response, err := e.geminiClient.GenerateContent(ctx, prompt)
			if err != nil {
				e.logger.ErrorWithErr(err, "Failed to analyze resource security batch")
				continue
			}

			if err := e.parseAndSaveRecommendations(ctx, userID, response); err != nil {
				e.logger.ErrorWithErr(err, "Failed to save security recommendations")
				continue
			}
		}

		if int64(offset+len(resources)) >= total {
			break
		}

		offset += len(resources)
	}

	if totalProcessed == 0 {
		e.logger.Info("No resources found, skipping resource security recommendations")
		return nil
	}

	e.logger.WithFields(map[string]interface{}{
		"total_resources": totalProcessed,
	}).Info("Processed resources for security recommendations")

	return nil
}

// generateComplianceRecommendations analyzes drifts for compliance issues
func (e *RecommendationEngine) generateComplianceRecommendations(ctx context.Context, userID int64) error {
	e.logger.Info("Generating compliance recommendations")

	const (
		pageSize  = 100
		batchSize = 20
	)
	offset := 0
	foundAny := false

	for {
		drifts, total, err := e.driftRepo.ListWithPagination(ctx, userID, drift.Filter{
			Status: "unresolved",
		}, pageSize, offset)
		if err != nil {
			return fmt.Errorf("failed to list drifts: %w", err)
		}

		if len(drifts) == 0 {
			break
		}

		foundAny = true

		// Rule-based fallback when Gemini is not available
		if e.geminiClient == nil {
			var errs []error
			for _, d := range drifts {
				priority := recommendation.PriorityMedium
				if d.Severity == "critical" {
					priority = recommendation.PriorityCritical
				} else if d.Severity == "high" {
					priority = recommendation.PriorityHigh
				}

				rec := &recommendation.Recommendation{
					UserID:      userID,
					Type:        recommendation.TypeCompliance,
					Priority:    priority,
					Title:       fmt.Sprintf("Resolve %s drift on resource %s", d.Severity, d.ResourceID),
					Description: fmt.Sprintf("Configuration drift (%s) detected on resource %s with %s severity. Review and remediate to maintain compliance.", d.DriftType, d.ResourceID, d.Severity),
					Category:    "drift_remediation",
					Savings:     0,
					Effort:      recommendation.EffortMedium,
					Impact:      recommendation.ImpactHigh,
					Resources:   []string{d.ResourceID},
				}
				if _, err := e.recRepo.Create(ctx, rec); err != nil {
					e.logger.ErrorWithErr(err, "Failed to save compliance recommendation")
					errs = append(errs, fmt.Errorf("failed to save compliance recommendation for %s: %w", d.ResourceID, err))
				}
			}
			if len(errs) > 0 {
				return fmt.Errorf("encountered %d errors while creating compliance recommendations: %v", len(errs), errs[0])
			}
		} else {
			// AI-powered analysis in batches
			for i := 0; i < len(drifts); i += batchSize {
				end := i + batchSize
				if end > len(drifts) {
					end = len(drifts)
				}

				batch := drifts[i:end]

				driftData := make([]map[string]interface{}, 0, len(batch))
				for _, d := range batch {
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
					e.logger.ErrorWithErr(err, "Failed to analyze drift batch")
					continue
				}

				if err := e.parseAndSaveRecommendations(ctx, userID, response); err != nil {
					e.logger.ErrorWithErr(err, "Failed to save compliance recommendations")
					continue
				}
			}
		}

		// Check if we've fetched all drifts
		if int64(offset+len(drifts)) >= total {
			break
		}

		offset += len(drifts)
	}

	if !foundAny {
		e.logger.Info("No unresolved drifts found, skipping compliance recommendations")
	}

	return nil
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
