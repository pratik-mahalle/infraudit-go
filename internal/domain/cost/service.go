package cost

import (
	"context"
	"time"
)

// Service defines the cost analytics service interface
type Service interface {
	// Cost Sync
	SyncCosts(ctx context.Context, userID int64, provider string) error
	SyncAllProviders(ctx context.Context, userID int64) error

	// Cost Queries
	GetCostOverview(ctx context.Context, userID int64) (*CostOverview, error)
	GetCostsByProvider(ctx context.Context, userID int64, provider string, filter Filter, period string) (*CostSummary, error)
	GetResourceCosts(ctx context.Context, userID int64, resourceID string, days int) ([]*Cost, error)
	GetCostTrends(ctx context.Context, userID int64, provider string, period string) (*CostTrend, error)
	GetCostForecast(ctx context.Context, userID int64, provider string, days int) (*CostForecast, error)

	// Cost Anomalies
	DetectAnomalies(ctx context.Context, userID int64, provider string) ([]*CostAnomaly, error)
	GetAnomalies(ctx context.Context, userID int64, status string, limit, offset int) ([]*CostAnomaly, int64, error)
	UpdateAnomalyStatus(ctx context.Context, id string, status string, notes string) error

	// Cost Optimizations
	GenerateOptimizations(ctx context.Context, userID int64, provider string) ([]*CostOptimization, error)
	GetOptimizations(ctx context.Context, userID int64, status string, limit, offset int) ([]*CostOptimization, int64, error)
	UpdateOptimizationStatus(ctx context.Context, id string, status string) error
	GetPotentialSavings(ctx context.Context, userID int64) (float64, error)

	// Provider-specific
	GetAWSCosts(ctx context.Context, userID int64, startDate, endDate time.Time) ([]*Cost, error)
	GetGCPCosts(ctx context.Context, userID int64, startDate, endDate time.Time) ([]*Cost, error)
	GetAzureCosts(ctx context.Context, userID int64, startDate, endDate time.Time) ([]*Cost, error)
}

// CostOverview represents a high-level cost summary
type CostOverview struct {
	TotalCost        float64            `json:"total_cost"`
	MonthlyCost      float64            `json:"monthly_cost"`
	DailyCost        float64            `json:"daily_cost"`
	Currency         string             `json:"currency"`
	ByProvider       map[string]float64 `json:"by_provider"`
	TopServices      []ServiceCost      `json:"top_services"`
	Trend            *CostTrend         `json:"trend"`
	AnomalyCount     int                `json:"anomaly_count"`
	PotentialSavings float64            `json:"potential_savings"`
}

// ServiceCost represents cost for a specific service
type ServiceCost struct {
	Provider    string  `json:"provider"`
	ServiceName string  `json:"service_name"`
	Cost        float64 `json:"cost"`
	Percentage  float64 `json:"percentage"`
}
