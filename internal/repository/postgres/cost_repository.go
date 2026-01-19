package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/pratik-mahalle/infraudit/internal/domain/cost"
)

// CostRepository implements cost.Repository
type CostRepository struct {
	db *sql.DB
}

// NewCostRepository creates a new cost repository
func NewCostRepository(db *sql.DB) *CostRepository {
	return &CostRepository{db: db}
}

// CreateCost creates a new cost record
func (r *CostRepository) CreateCost(ctx context.Context, c *cost.Cost) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}

	query := `
		INSERT INTO resource_costs (id, user_id, resource_id, provider, region, service_name, resource_type, cost_date, daily_cost, monthly_cost, currency, cost_details, tags, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		c.ID, c.UserID, c.ResourceID, c.Provider, c.Region, c.ServiceName, c.ResourceType,
		c.CostDate, c.DailyCost, c.MonthlyCost, c.Currency, c.CostDetails, c.Tags,
		time.Now(), time.Now(),
	)
	return err
}

// GetCostsByDateRange retrieves costs within a date range
func (r *CostRepository) GetCostsByDateRange(ctx context.Context, userID int64, filter cost.Filter, startDate, endDate time.Time) ([]*cost.Cost, error) {
	query := `
		SELECT id, user_id, resource_id, provider, region, service_name, resource_type, cost_date, daily_cost, monthly_cost, currency, cost_details, tags, created_at, updated_at
		FROM resource_costs
		WHERE user_id = ? AND cost_date BETWEEN ? AND ?
	`
	args := []interface{}{userID, startDate, endDate}

	if filter.Provider != "" {
		query += " AND provider = ?"
		args = append(args, filter.Provider)
	}
	if filter.ServiceName != "" {
		query += " AND service_name = ?"
		args = append(args, filter.ServiceName)
	}
	if filter.ResourceID != "" {
		query += " AND resource_id = ?"
		args = append(args, filter.ResourceID)
	}
	if filter.Region != "" {
		query += " AND region = ?"
		args = append(args, filter.Region)
	}

	query += " ORDER BY cost_date DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var costs []*cost.Cost
	for rows.Next() {
		c := &cost.Cost{}
		err := rows.Scan(
			&c.ID, &c.UserID, &c.ResourceID, &c.Provider, &c.Region, &c.ServiceName, &c.ResourceType,
			&c.CostDate, &c.DailyCost, &c.MonthlyCost, &c.Currency, &c.CostDetails, &c.Tags,
			&c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		costs = append(costs, c)
	}

	return costs, rows.Err()
}

// GetCostSummary retrieves aggregated cost data
func (r *CostRepository) GetCostSummary(ctx context.Context, userID int64, filter cost.Filter, startDate, endDate time.Time) (*cost.CostSummary, error) {
	summary := &cost.CostSummary{
		Currency:   "USD",
		Period:     "custom",
		StartDate:  startDate,
		EndDate:    endDate,
		ByService:  make(map[string]float64),
		ByRegion:   make(map[string]float64),
		ByResource: make(map[string]float64),
	}

	// Get total cost
	query := `
		SELECT COALESCE(SUM(daily_cost), 0)
		FROM resource_costs
		WHERE user_id = ? AND cost_date BETWEEN ? AND ?
	`
	args := []interface{}{userID, startDate, endDate}

	if filter.Provider != "" {
		query += " AND provider = ?"
		args = append(args, filter.Provider)
		summary.Provider = filter.Provider
	}

	err := r.db.QueryRowContext(ctx, query, args...).Scan(&summary.TotalCost)
	if err != nil {
		return nil, err
	}

	// Get cost by service
	serviceQuery := `
		SELECT service_name, COALESCE(SUM(daily_cost), 0) as total
		FROM resource_costs
		WHERE user_id = ? AND cost_date BETWEEN ? AND ?
		GROUP BY service_name
		ORDER BY total DESC
	`
	rows, err := r.db.QueryContext(ctx, serviceQuery, userID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var service string
		var total float64
		if err := rows.Scan(&service, &total); err != nil {
			return nil, err
		}
		summary.ByService[service] = total
	}

	// Get cost by region
	regionQuery := `
		SELECT region, COALESCE(SUM(daily_cost), 0) as total
		FROM resource_costs
		WHERE user_id = ? AND cost_date BETWEEN ? AND ? AND region IS NOT NULL
		GROUP BY region
		ORDER BY total DESC
	`
	rows, err = r.db.QueryContext(ctx, regionQuery, userID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var region string
		var total float64
		if err := rows.Scan(&region, &total); err != nil {
			return nil, err
		}
		summary.ByRegion[region] = total
	}

	return summary, nil
}

// GetDailyCosts retrieves daily cost data points
func (r *CostRepository) GetDailyCosts(ctx context.Context, userID int64, filter cost.Filter, days int) ([]cost.CostDataPoint, error) {
	startDate := time.Now().AddDate(0, 0, -days)

	query := `
		SELECT cost_date, COALESCE(SUM(daily_cost), 0) as total
		FROM resource_costs
		WHERE user_id = ? AND cost_date >= ?
	`
	args := []interface{}{userID, startDate}

	if filter.Provider != "" {
		query += " AND provider = ?"
		args = append(args, filter.Provider)
	}

	query += " GROUP BY cost_date ORDER BY cost_date ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dataPoints []cost.CostDataPoint
	for rows.Next() {
		var dp cost.CostDataPoint
		if err := rows.Scan(&dp.Date, &dp.Cost); err != nil {
			return nil, err
		}
		dataPoints = append(dataPoints, dp)
	}

	return dataPoints, rows.Err()
}

// DeleteCostsByDate deletes costs older than a given date
func (r *CostRepository) DeleteCostsByDate(ctx context.Context, userID int64, beforeDate time.Time) error {
	query := `DELETE FROM resource_costs WHERE user_id = ? AND cost_date < ?`
	_, err := r.db.ExecContext(ctx, query, userID, beforeDate)
	return err
}

// CreateAnomaly creates a new cost anomaly record
func (r *CostRepository) CreateAnomaly(ctx context.Context, a *cost.CostAnomaly) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}

	query := `
		INSERT INTO cost_anomalies (id, user_id, provider, service_name, resource_id, anomaly_type, expected_cost, actual_cost, deviation, severity, status, notes, detected_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		a.ID, a.UserID, a.Provider, a.ServiceName, a.ResourceID, a.AnomalyType,
		a.ExpectedCost, a.ActualCost, a.Deviation, a.Severity, a.Status, a.Notes,
		a.DetectedAt, time.Now(),
	)
	return err
}

// GetAnomaly retrieves an anomaly by ID
func (r *CostRepository) GetAnomaly(ctx context.Context, id string) (*cost.CostAnomaly, error) {
	query := `
		SELECT id, user_id, provider, service_name, resource_id, anomaly_type, expected_cost, actual_cost, deviation, severity, status, notes, detected_at, created_at
		FROM cost_anomalies
		WHERE id = ?
	`
	a := &cost.CostAnomaly{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&a.ID, &a.UserID, &a.Provider, &a.ServiceName, &a.ResourceID, &a.AnomalyType,
		&a.ExpectedCost, &a.ActualCost, &a.Deviation, &a.Severity, &a.Status, &a.Notes,
		&a.DetectedAt, &a.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return a, nil
}

// UpdateAnomaly updates an anomaly
func (r *CostRepository) UpdateAnomaly(ctx context.Context, a *cost.CostAnomaly) error {
	query := `UPDATE cost_anomalies SET status = ?, notes = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, a.Status, a.Notes, a.ID)
	return err
}

// ListAnomalies lists cost anomalies
func (r *CostRepository) ListAnomalies(ctx context.Context, userID int64, status string, limit, offset int) ([]*cost.CostAnomaly, int64, error) {
	countQuery := `SELECT COUNT(*) FROM cost_anomalies WHERE user_id = ?`
	args := []interface{}{userID}

	if status != "" {
		countQuery += " AND status = ?"
		args = append(args, status)
	}

	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, user_id, provider, service_name, resource_id, anomaly_type, expected_cost, actual_cost, deviation, severity, status, notes, detected_at, created_at
		FROM cost_anomalies
		WHERE user_id = ?
	`
	queryArgs := []interface{}{userID}

	if status != "" {
		query += " AND status = ?"
		queryArgs = append(queryArgs, status)
	}

	query += " ORDER BY detected_at DESC LIMIT ? OFFSET ?"
	queryArgs = append(queryArgs, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var anomalies []*cost.CostAnomaly
	for rows.Next() {
		a := &cost.CostAnomaly{}
		err := rows.Scan(
			&a.ID, &a.UserID, &a.Provider, &a.ServiceName, &a.ResourceID, &a.AnomalyType,
			&a.ExpectedCost, &a.ActualCost, &a.Deviation, &a.Severity, &a.Status, &a.Notes,
			&a.DetectedAt, &a.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		anomalies = append(anomalies, a)
	}

	return anomalies, total, rows.Err()
}

// CreateOptimization creates a new optimization record
func (r *CostRepository) CreateOptimization(ctx context.Context, o *cost.CostOptimization) error {
	if o.ID == "" {
		o.ID = uuid.New().String()
	}

	detailsJSON, _ := json.Marshal(o.Details)

	query := `
		INSERT INTO cost_optimizations (id, user_id, provider, resource_id, resource_type, optimization_type, title, description, current_cost, estimated_savings, savings_percent, implementation, status, details, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		o.ID, o.UserID, o.Provider, o.ResourceID, o.ResourceType, o.OptimizationType,
		o.Title, o.Description, o.CurrentCost, o.EstimatedSavings, o.SavingsPercent,
		o.Implementation, o.Status, detailsJSON, time.Now(), time.Now(),
	)
	return err
}

// GetOptimization retrieves an optimization by ID
func (r *CostRepository) GetOptimization(ctx context.Context, id string) (*cost.CostOptimization, error) {
	query := `
		SELECT id, user_id, provider, resource_id, resource_type, optimization_type, title, description, current_cost, estimated_savings, savings_percent, implementation, status, details, created_at, updated_at
		FROM cost_optimizations
		WHERE id = ?
	`
	o := &cost.CostOptimization{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&o.ID, &o.UserID, &o.Provider, &o.ResourceID, &o.ResourceType, &o.OptimizationType,
		&o.Title, &o.Description, &o.CurrentCost, &o.EstimatedSavings, &o.SavingsPercent,
		&o.Implementation, &o.Status, &o.Details, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return o, nil
}

// UpdateOptimization updates an optimization
func (r *CostRepository) UpdateOptimization(ctx context.Context, o *cost.CostOptimization) error {
	query := `UPDATE cost_optimizations SET status = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, o.Status, time.Now(), o.ID)
	return err
}

// ListOptimizations lists optimizations
func (r *CostRepository) ListOptimizations(ctx context.Context, userID int64, status string, limit, offset int) ([]*cost.CostOptimization, int64, error) {
	countQuery := `SELECT COUNT(*) FROM cost_optimizations WHERE user_id = ?`
	args := []interface{}{userID}

	if status != "" {
		countQuery += " AND status = ?"
		args = append(args, status)
	}

	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, user_id, provider, resource_id, resource_type, optimization_type, title, description, current_cost, estimated_savings, savings_percent, implementation, status, details, created_at, updated_at
		FROM cost_optimizations
		WHERE user_id = ?
	`
	queryArgs := []interface{}{userID}

	if status != "" {
		query += " AND status = ?"
		queryArgs = append(queryArgs, status)
	}

	query += " ORDER BY estimated_savings DESC LIMIT ? OFFSET ?"
	queryArgs = append(queryArgs, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var optimizations []*cost.CostOptimization
	for rows.Next() {
		o := &cost.CostOptimization{}
		err := rows.Scan(
			&o.ID, &o.UserID, &o.Provider, &o.ResourceID, &o.ResourceType, &o.OptimizationType,
			&o.Title, &o.Description, &o.CurrentCost, &o.EstimatedSavings, &o.SavingsPercent,
			&o.Implementation, &o.Status, &o.Details, &o.CreatedAt, &o.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		optimizations = append(optimizations, o)
	}

	return optimizations, total, rows.Err()
}

// GetTotalPotentialSavings returns total potential savings
func (r *CostRepository) GetTotalPotentialSavings(ctx context.Context, userID int64) (float64, error) {
	query := `SELECT COALESCE(SUM(estimated_savings), 0) FROM cost_optimizations WHERE user_id = ? AND status = 'pending'`
	var total float64
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&total)
	return total, err
}
