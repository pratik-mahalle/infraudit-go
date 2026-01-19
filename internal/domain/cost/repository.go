package cost

import (
	"context"
	"time"
)

// Repository defines the cost repository interface
type Repository interface {
	// Costs
	CreateCost(ctx context.Context, cost *Cost) error
	GetCostsByDateRange(ctx context.Context, userID int64, filter Filter, startDate, endDate time.Time) ([]*Cost, error)
	GetCostSummary(ctx context.Context, userID int64, filter Filter, startDate, endDate time.Time) (*CostSummary, error)
	GetDailyCosts(ctx context.Context, userID int64, filter Filter, days int) ([]CostDataPoint, error)
	DeleteCostsByDate(ctx context.Context, userID int64, beforeDate time.Time) error

	// Anomalies
	CreateAnomaly(ctx context.Context, anomaly *CostAnomaly) error
	GetAnomaly(ctx context.Context, id string) (*CostAnomaly, error)
	UpdateAnomaly(ctx context.Context, anomaly *CostAnomaly) error
	ListAnomalies(ctx context.Context, userID int64, status string, limit, offset int) ([]*CostAnomaly, int64, error)

	// Optimizations
	CreateOptimization(ctx context.Context, opt *CostOptimization) error
	GetOptimization(ctx context.Context, id string) (*CostOptimization, error)
	UpdateOptimization(ctx context.Context, opt *CostOptimization) error
	ListOptimizations(ctx context.Context, userID int64, status string, limit, offset int) ([]*CostOptimization, int64, error)
	GetTotalPotentialSavings(ctx context.Context, userID int64) (float64, error)
}
