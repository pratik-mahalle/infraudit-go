package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pratik-mahalle/infraudit/internal/domain/cost"
	"github.com/pratik-mahalle/infraudit/internal/domain/provider"
	"github.com/pratik-mahalle/infraudit/internal/integrations"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
	"github.com/pratik-mahalle/infraudit/internal/providers"
)

// CostServiceImpl implements cost.Service
type CostServiceImpl struct {
	repo         cost.Repository
	providerRepo provider.Repository
	geminiClient *integrations.GeminiClient
	logger       *logger.Logger
}

// NewCostService creates a new cost service
func NewCostService(repo cost.Repository, providerRepo provider.Repository, geminiClient *integrations.GeminiClient, log *logger.Logger) cost.Service {
	return &CostServiceImpl{
		repo:         repo,
		providerRepo: providerRepo,
		geminiClient: geminiClient,
		logger:       log,
	}
}

// SyncCosts syncs costs for a specific provider
func (s *CostServiceImpl) SyncCosts(ctx context.Context, userID int64, provider string) error {
	s.logger.WithFields(map[string]interface{}{
		"user_id":  userID,
		"provider": provider,
	}).Info("Syncing costs from provider")

	switch provider {
	case cost.ProviderAWS:
		return s.syncAWSCosts(ctx, userID)
	case cost.ProviderGCP:
		return s.syncGCPCosts(ctx, userID)
	case cost.ProviderAzure:
		return s.syncAzureCosts(ctx, userID)
	default:
		return nil
	}
}

// SyncAllProviders syncs costs from all configured providers
func (s *CostServiceImpl) SyncAllProviders(ctx context.Context, userID int64) error {
	providers := []string{cost.ProviderAWS, cost.ProviderGCP, cost.ProviderAzure}
	for _, provider := range providers {
		if err := s.SyncCosts(ctx, userID, provider); err != nil {
			s.logger.WithFields(map[string]interface{}{
				"user_id":  userID,
				"provider": provider,
			}).ErrorWithErr(err, "Failed to sync costs")
			// Continue with other providers
		}
	}
	return nil
}

// GetCostOverview returns a high-level cost summary
func (s *CostServiceImpl) GetCostOverview(ctx context.Context, userID int64) (*cost.CostOverview, error) {
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := now

	// Get monthly costs by provider
	byProvider := make(map[string]float64)
	providers := []string{cost.ProviderAWS, cost.ProviderGCP, cost.ProviderAzure}

	var totalCost float64
	for _, provider := range providers {
		summary, err := s.repo.GetCostSummary(ctx, userID, cost.Filter{Provider: provider}, monthStart, monthEnd)
		if err == nil && summary != nil {
			byProvider[provider] = summary.TotalCost
			totalCost += summary.TotalCost
		}
	}

	// Get daily cost average
	daysCnt := now.Day()
	dailyCost := 0.0
	if daysCnt > 0 {
		dailyCost = totalCost / float64(daysCnt)
	}

	// Get top services
	summary, _ := s.repo.GetCostSummary(ctx, userID, cost.Filter{}, monthStart, monthEnd)
	var topServices []cost.ServiceCost
	if summary != nil {
		for service, svcCost := range summary.ByService {
			pct := 0.0
			if totalCost > 0 {
				pct = (svcCost / totalCost) * 100
			}
			topServices = append(topServices, cost.ServiceCost{
				ServiceName: service,
				Cost:        svcCost,
				Percentage:  pct,
			})
		}
	}

	// Get potential savings
	savings, _ := s.repo.GetTotalPotentialSavings(ctx, userID)

	// Get anomaly count
	anomalies, _, _ := s.repo.ListAnomalies(ctx, userID, cost.AnomalyStatusOpen, 100, 0)
	anomalyCount := len(anomalies)

	// Get cost trend
	trend, _ := s.GetCostTrends(ctx, userID, "", "monthly")

	return &cost.CostOverview{
		TotalCost:        totalCost,
		MonthlyCost:      totalCost,
		DailyCost:        dailyCost,
		Currency:         "USD",
		ByProvider:       byProvider,
		TopServices:      topServices,
		Trend:            trend,
		AnomalyCount:     anomalyCount,
		PotentialSavings: savings,
	}, nil
}

// GetCostsByProvider returns cost summary for a specific provider
func (s *CostServiceImpl) GetCostsByProvider(ctx context.Context, userID int64, provider string, filter cost.Filter, period string) (*cost.CostSummary, error) {
	startDate, endDate := s.getPeriodDates(period)
	filter.Provider = provider
	return s.repo.GetCostSummary(ctx, userID, filter, startDate, endDate)
}

// GetResourceCosts returns costs for a specific resource
func (s *CostServiceImpl) GetResourceCosts(ctx context.Context, userID int64, resourceID string, days int) ([]*cost.Cost, error) {
	startDate := time.Now().AddDate(0, 0, -days)
	endDate := time.Now()
	return s.repo.GetCostsByDateRange(ctx, userID, cost.Filter{ResourceID: resourceID}, startDate, endDate)
}

// GetCostTrends returns cost trends
func (s *CostServiceImpl) GetCostTrends(ctx context.Context, userID int64, provider string, period string) (*cost.CostTrend, error) {
	var days int
	switch period {
	case "weekly":
		days = 14
	case "monthly":
		days = 60
	default:
		days = 30
	}

	dataPoints, err := s.repo.GetDailyCosts(ctx, userID, cost.Filter{Provider: provider}, days)
	if err != nil {
		return nil, err
	}

	trend := &cost.CostTrend{
		Period:     period,
		DataPoints: dataPoints,
	}

	// Calculate current and previous period costs
	midpoint := len(dataPoints) / 2
	if len(dataPoints) > 1 {
		var currentSum, previousSum float64
		for i, dp := range dataPoints {
			if i >= midpoint {
				currentSum += dp.Cost
			} else {
				previousSum += dp.Cost
			}
		}
		trend.CurrentCost = currentSum
		trend.PreviousCost = previousSum

		if previousSum > 0 {
			trend.ChangePercent = ((currentSum - previousSum) / previousSum) * 100
		}

		if currentSum > previousSum {
			trend.Trend = "up"
		} else if currentSum < previousSum {
			trend.Trend = "down"
		} else {
			trend.Trend = "stable"
		}
	}

	return trend, nil
}

// GetCostForecast returns cost forecast
func (s *CostServiceImpl) GetCostForecast(ctx context.Context, userID int64, provider string, days int) (*cost.CostForecast, error) {
	// Get historical data
	dataPoints, err := s.repo.GetDailyCosts(ctx, userID, cost.Filter{Provider: provider}, 30)
	if err != nil {
		return nil, err
	}

	if len(dataPoints) == 0 {
		return &cost.CostForecast{
			Provider:        provider,
			Period:          "forecast",
			ForecastedCost:  0,
			ConfidenceLevel: 0,
			Currency:        "USD",
			EndDate:         time.Now().AddDate(0, 0, days),
		}, nil
	}

	// Simple linear extrapolation
	var sum float64
	for _, dp := range dataPoints {
		sum += dp.Cost
	}
	avgDailyCost := sum / float64(len(dataPoints))
	forecasted := avgDailyCost * float64(days)

	return &cost.CostForecast{
		Provider:        provider,
		Period:          "forecast",
		ForecastedCost:  forecasted,
		ConfidenceLevel: 0.7,
		LowerBound:      forecasted * 0.8,
		UpperBound:      forecasted * 1.2,
		Currency:        "USD",
		EndDate:         time.Now().AddDate(0, 0, days),
	}, nil
}

// DetectAnomalies detects cost anomalies
func (s *CostServiceImpl) DetectAnomalies(ctx context.Context, userID int64, provider string) ([]*cost.CostAnomaly, error) {
	// Get historical daily costs
	dataPoints, err := s.repo.GetDailyCosts(ctx, userID, cost.Filter{Provider: provider}, 30)
	if err != nil {
		return nil, err
	}

	if len(dataPoints) < 7 {
		return nil, nil // Not enough data
	}

	// Calculate mean and standard deviation
	var sum float64
	for _, dp := range dataPoints {
		sum += dp.Cost
	}
	mean := sum / float64(len(dataPoints))

	var varianceSum float64
	for _, dp := range dataPoints {
		diff := dp.Cost - mean
		varianceSum += diff * diff
	}
	stdDev := 0.0
	if len(dataPoints) > 1 {
		stdDev = varianceSum / float64(len(dataPoints)-1)
	}

	// Detect anomalies (costs > 2 standard deviations from mean)
	var anomalies []*cost.CostAnomaly
	threshold := mean + (2 * stdDev)

	for _, dp := range dataPoints {
		if dp.Cost > threshold {
			deviation := 0.0
			if mean > 0 {
				deviation = ((dp.Cost - mean) / mean) * 100
			}

			anomaly := &cost.CostAnomaly{
				UserID:       userID,
				Provider:     provider,
				AnomalyType:  cost.AnomalyTypeSpike,
				ExpectedCost: mean,
				ActualCost:   dp.Cost,
				Deviation:    deviation,
				Severity:     s.getSeverityFromDeviation(deviation),
				Status:       cost.AnomalyStatusOpen,
				DetectedAt:   dp.Date,
			}

			if err := s.repo.CreateAnomaly(ctx, anomaly); err == nil {
				anomalies = append(anomalies, anomaly)
			}
		}
	}

	return anomalies, nil
}

// GetAnomalies returns cost anomalies
func (s *CostServiceImpl) GetAnomalies(ctx context.Context, userID int64, status string, limit, offset int) ([]*cost.CostAnomaly, int64, error) {
	return s.repo.ListAnomalies(ctx, userID, status, limit, offset)
}

// UpdateAnomalyStatus updates an anomaly status
func (s *CostServiceImpl) UpdateAnomalyStatus(ctx context.Context, id string, status string, notes string) error {
	anomaly, err := s.repo.GetAnomaly(ctx, id)
	if err != nil {
		return err
	}
	anomaly.Status = status
	anomaly.Notes = notes
	return s.repo.UpdateAnomaly(ctx, anomaly)
}

// GenerateOptimizations generates cost optimization recommendations using Gemini AI
func (s *CostServiceImpl) GenerateOptimizations(ctx context.Context, userID int64, providerFilter string) ([]*cost.CostOptimization, error) {
	if s.geminiClient == nil {
		s.logger.Info("Gemini client not configured, skipping optimization generation")
		return nil, nil
	}

	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	filter := cost.Filter{}
	if providerFilter != "" {
		filter.Provider = providerFilter
	}

	summary, err := s.repo.GetCostSummary(ctx, userID, filter, monthStart, now)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost summary: %w", err)
	}
	if summary == nil || summary.TotalCost == 0 {
		return nil, nil
	}

	summaryJSON, _ := json.MarshalIndent(summary, "", "  ")

	prompt := fmt.Sprintf(`Analyze the following cloud cost summary and provide cost optimization recommendations.

Cost Summary:
%s

Provide recommendations in JSON format. Each recommendation should have:
{
  "provider": "aws|gcp|azure",
  "resource_type": "type of resource",
  "optimization_type": "rightsize|unused|reserved_instance|savings_plan|storage_class|idle_resource|region_optimization",
  "title": "Brief title",
  "description": "Detailed description with specific action items",
  "current_cost": estimated_current_monthly_cost,
  "estimated_savings": estimated_monthly_savings,
  "savings_percent": percentage_savings,
  "implementation": "easy|moderate|complex"
}

Return ONLY a JSON array of recommendations, no additional text.`, string(summaryJSON))

	response, err := s.geminiClient.GenerateContent(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate optimizations: %w", err)
	}

	// Strip markdown code fences if present
	response = strings.TrimSpace(response)
	if strings.HasPrefix(response, "```") {
		lines := strings.SplitN(response, "\n", 2)
		if len(lines) > 1 {
			response = lines[1]
		}
		if idx := strings.LastIndex(response, "```"); idx != -1 {
			response = response[:idx]
		}
		response = strings.TrimSpace(response)
	}

	var rawOptimizations []struct {
		Provider         string  `json:"provider"`
		ResourceType     string  `json:"resource_type"`
		OptimizationType string  `json:"optimization_type"`
		Title            string  `json:"title"`
		Description      string  `json:"description"`
		CurrentCost      float64 `json:"current_cost"`
		EstimatedSavings float64 `json:"estimated_savings"`
		SavingsPercent   float64 `json:"savings_percent"`
		Implementation   string  `json:"implementation"`
	}

	if err := json.Unmarshal([]byte(response), &rawOptimizations); err != nil {
		s.logger.WithFields(map[string]interface{}{
			"error":    err.Error(),
			"response": response[:min(200, len(response))],
		}).Warn("Failed to parse AI optimization response")
		return nil, nil
	}

	var optimizations []*cost.CostOptimization
	for _, raw := range rawOptimizations {
		opt := &cost.CostOptimization{
			UserID:           userID,
			Provider:         raw.Provider,
			ResourceType:     raw.ResourceType,
			OptimizationType: raw.OptimizationType,
			Title:            raw.Title,
			Description:      raw.Description,
			CurrentCost:      raw.CurrentCost,
			EstimatedSavings: raw.EstimatedSavings,
			SavingsPercent:   raw.SavingsPercent,
			Implementation:   raw.Implementation,
			Status:           cost.OptStatusPending,
		}

		if err := s.repo.CreateOptimization(ctx, opt); err != nil {
			s.logger.WithFields(map[string]interface{}{
				"title": opt.Title,
			}).ErrorWithErr(err, "Failed to save optimization")
			continue
		}
		optimizations = append(optimizations, opt)
	}

	return optimizations, nil
}

// GetOptimizations returns cost optimizations
func (s *CostServiceImpl) GetOptimizations(ctx context.Context, userID int64, status string, limit, offset int) ([]*cost.CostOptimization, int64, error) {
	return s.repo.ListOptimizations(ctx, userID, status, limit, offset)
}

// UpdateOptimizationStatus updates an optimization status
func (s *CostServiceImpl) UpdateOptimizationStatus(ctx context.Context, id string, status string) error {
	opt, err := s.repo.GetOptimization(ctx, id)
	if err != nil {
		return err
	}
	opt.Status = status
	return s.repo.UpdateOptimization(ctx, opt)
}

// GetPotentialSavings returns total potential savings
func (s *CostServiceImpl) GetPotentialSavings(ctx context.Context, userID int64) (float64, error) {
	return s.repo.GetTotalPotentialSavings(ctx, userID)
}

// GetAWSCosts returns AWS-specific costs
func (s *CostServiceImpl) GetAWSCosts(ctx context.Context, userID int64, startDate, endDate time.Time) ([]*cost.Cost, error) {
	return s.repo.GetCostsByDateRange(ctx, userID, cost.Filter{Provider: cost.ProviderAWS}, startDate, endDate)
}

// GetGCPCosts returns GCP-specific costs
func (s *CostServiceImpl) GetGCPCosts(ctx context.Context, userID int64, startDate, endDate time.Time) ([]*cost.Cost, error) {
	return s.repo.GetCostsByDateRange(ctx, userID, cost.Filter{Provider: cost.ProviderGCP}, startDate, endDate)
}

// GetAzureCosts returns Azure-specific costs
func (s *CostServiceImpl) GetAzureCosts(ctx context.Context, userID int64, startDate, endDate time.Time) ([]*cost.Cost, error) {
	return s.repo.GetCostsByDateRange(ctx, userID, cost.Filter{Provider: cost.ProviderAzure}, startDate, endDate)
}

// Helper methods

func (s *CostServiceImpl) syncAWSCosts(ctx context.Context, userID int64) error {
	providerAccount, err := s.providerRepo.GetByProvider(ctx, userID, "aws")
	if err != nil || providerAccount == nil || !providerAccount.IsConnected {
		s.logger.Info("AWS provider not connected, skipping cost sync")
		return nil
	}

	creds := providers.AWSCredentials{
		AccessKeyID:    providerAccount.Credentials.AWSAccessKeyID,
		SecretAccessKey: providerAccount.Credentials.AWSSecretAccessKey,
		Region:         providerAccount.Credentials.AWSRegion,
	}

	costs, err := providers.FetchAWSCosts(ctx, creds)
	if err != nil {
		return fmt.Errorf("failed to fetch AWS costs: %w", err)
	}

	for i := range costs {
		costs[i].UserID = userID
		if err := s.repo.CreateCost(ctx, &costs[i]); err != nil {
			s.logger.WithFields(map[string]interface{}{
				"service": costs[i].ServiceName,
				"date":    costs[i].CostDate,
			}).ErrorWithErr(err, "Failed to persist AWS cost record")
		}
	}

	s.logger.WithFields(map[string]interface{}{
		"user_id": userID,
		"count":   len(costs),
	}).Info("AWS cost sync completed")
	return nil
}

func (s *CostServiceImpl) syncGCPCosts(ctx context.Context, userID int64) error {
	providerAccount, err := s.providerRepo.GetByProvider(ctx, userID, "gcp")
	if err != nil || providerAccount == nil || !providerAccount.IsConnected {
		s.logger.Info("GCP provider not connected, skipping cost sync")
		return nil
	}

	if providerAccount.Credentials.GCPBillingDataset == "" {
		s.logger.Info("GCP billing dataset not configured, skipping cost sync")
		return nil
	}

	creds := providers.GCPBillingCredentials{
		ProjectID:          providerAccount.Credentials.GCPProjectID,
		ServiceAccountJSON: providerAccount.Credentials.GCPServiceAccountJSON,
		BillingDataset:     providerAccount.Credentials.GCPBillingDataset,
	}

	costs, err := providers.FetchGCPCosts(ctx, creds)
	if err != nil {
		return fmt.Errorf("failed to fetch GCP costs: %w", err)
	}

	for i := range costs {
		costs[i].UserID = userID
		if err := s.repo.CreateCost(ctx, &costs[i]); err != nil {
			s.logger.WithFields(map[string]interface{}{
				"service": costs[i].ServiceName,
				"date":    costs[i].CostDate,
			}).ErrorWithErr(err, "Failed to persist GCP cost record")
		}
	}

	s.logger.WithFields(map[string]interface{}{
		"user_id": userID,
		"count":   len(costs),
	}).Info("GCP cost sync completed")
	return nil
}

func (s *CostServiceImpl) syncAzureCosts(ctx context.Context, userID int64) error {
	providerAccount, err := s.providerRepo.GetByProvider(ctx, userID, "azure")
	if err != nil || providerAccount == nil || !providerAccount.IsConnected {
		s.logger.Info("Azure provider not connected, skipping cost sync")
		return nil
	}

	creds := providers.AzureCredentials{
		TenantID:       providerAccount.Credentials.AzureTenantID,
		ClientID:       providerAccount.Credentials.AzureClientID,
		ClientSecret:   providerAccount.Credentials.AzureClientSecret,
		SubscriptionID: providerAccount.Credentials.AzureSubscriptionID,
		Location:       providerAccount.Credentials.AzureLocation,
	}

	costs, err := providers.FetchAzureCosts(ctx, creds)
	if err != nil {
		return fmt.Errorf("failed to fetch Azure costs: %w", err)
	}

	for i := range costs {
		costs[i].UserID = userID
		if err := s.repo.CreateCost(ctx, &costs[i]); err != nil {
			s.logger.WithFields(map[string]interface{}{
				"service": costs[i].ServiceName,
				"date":    costs[i].CostDate,
			}).ErrorWithErr(err, "Failed to persist Azure cost record")
		}
	}

	s.logger.WithFields(map[string]interface{}{
		"user_id": userID,
		"count":   len(costs),
	}).Info("Azure cost sync completed")
	return nil
}

func (s *CostServiceImpl) getPeriodDates(period string) (time.Time, time.Time) {
	now := time.Now()
	endDate := now

	var startDate time.Time
	switch period {
	case "daily":
		startDate = now.AddDate(0, 0, -1)
	case "weekly":
		startDate = now.AddDate(0, 0, -7)
	case "monthly":
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	case "yearly":
		startDate = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	default:
		startDate = now.AddDate(0, 0, -30)
	}

	return startDate, endDate
}

func (s *CostServiceImpl) getSeverityFromDeviation(deviation float64) string {
	if deviation > 100 {
		return "critical"
	} else if deviation > 50 {
		return "high"
	} else if deviation > 25 {
		return "medium"
	}
	return "low"
}
