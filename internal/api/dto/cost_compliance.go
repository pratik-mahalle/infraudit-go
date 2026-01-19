package dto

import (
	"encoding/json"
	"time"
)

// ======= Cost DTOs =======

// CostOverviewResponse represents the cost overview response
type CostOverviewResponse struct {
	TotalCost        float64            `json:"total_cost"`
	MonthlyCost      float64            `json:"monthly_cost"`
	DailyCost        float64            `json:"daily_cost"`
	Currency         string             `json:"currency"`
	ByProvider       map[string]float64 `json:"by_provider"`
	TopServices      []ServiceCostDTO   `json:"top_services"`
	Trend            *CostTrendDTO      `json:"trend,omitempty"`
	AnomalyCount     int                `json:"anomaly_count"`
	PotentialSavings float64            `json:"potential_savings"`
}

// ServiceCostDTO represents cost for a specific service
type ServiceCostDTO struct {
	Provider    string  `json:"provider,omitempty"`
	ServiceName string  `json:"service_name"`
	Cost        float64 `json:"cost"`
	Percentage  float64 `json:"percentage"`
}

// CostSummaryResponse represents a cost summary
type CostSummaryResponse struct {
	Provider  string             `json:"provider"`
	TotalCost float64            `json:"total_cost"`
	Currency  string             `json:"currency"`
	Period    string             `json:"period"`
	StartDate string             `json:"start_date"`
	EndDate   string             `json:"end_date"`
	ByService map[string]float64 `json:"by_service"`
	ByRegion  map[string]float64 `json:"by_region"`
}

// CostTrendDTO represents cost trend data
type CostTrendDTO struct {
	Period        string             `json:"period"`
	CurrentCost   float64            `json:"current_cost"`
	PreviousCost  float64            `json:"previous_cost"`
	Change        float64            `json:"change"`
	ChangePercent float64            `json:"change_percent"`
	Trend         string             `json:"trend"`
	DataPoints    []CostDataPointDTO `json:"data_points"`
}

// CostDataPointDTO represents a single cost data point
type CostDataPointDTO struct {
	Date string  `json:"date"`
	Cost float64 `json:"cost"`
}

// CostForecastResponse represents cost forecast
type CostForecastResponse struct {
	Provider        string  `json:"provider"`
	Period          string  `json:"period"`
	ForecastedCost  float64 `json:"forecasted_cost"`
	ConfidenceLevel float64 `json:"confidence_level"`
	LowerBound      float64 `json:"lower_bound"`
	UpperBound      float64 `json:"upper_bound"`
	Currency        string  `json:"currency"`
	EndDate         string  `json:"end_date"`
}

// CostAnomalyResponse represents a cost anomaly
type CostAnomalyResponse struct {
	ID           string    `json:"id"`
	Provider     string    `json:"provider"`
	ServiceName  string    `json:"service_name"`
	ResourceID   *string   `json:"resource_id,omitempty"`
	AnomalyType  string    `json:"anomaly_type"`
	ExpectedCost float64   `json:"expected_cost"`
	ActualCost   float64   `json:"actual_cost"`
	Deviation    float64   `json:"deviation"`
	Severity     string    `json:"severity"`
	Status       string    `json:"status"`
	Notes        string    `json:"notes,omitempty"`
	DetectedAt   time.Time `json:"detected_at"`
}

// ListAnomaliesResponse represents a list of anomalies
type ListAnomaliesResponse struct {
	Anomalies []CostAnomalyResponse `json:"anomalies"`
	Total     int64                 `json:"total"`
}

// UpdateCostAnomalyRequest represents a request to update a cost anomaly
type UpdateCostAnomalyRequest struct {
	Status string `json:"status"`
	Notes  string `json:"notes,omitempty"`
}

// CostOptimizationResponse represents a cost optimization
type CostOptimizationResponse struct {
	ID               string          `json:"id"`
	Provider         string          `json:"provider"`
	ResourceID       *string         `json:"resource_id,omitempty"`
	ResourceType     string          `json:"resource_type"`
	OptimizationType string          `json:"optimization_type"`
	Title            string          `json:"title"`
	Description      string          `json:"description"`
	CurrentCost      float64         `json:"current_cost"`
	EstimatedSavings float64         `json:"estimated_savings"`
	SavingsPercent   float64         `json:"savings_percent"`
	Implementation   string          `json:"implementation"`
	Status           string          `json:"status"`
	Details          json.RawMessage `json:"details,omitempty"`
}

// ListOptimizationsResponse represents a list of optimizations
type ListOptimizationsResponse struct {
	Optimizations []CostOptimizationResponse `json:"optimizations"`
	Total         int64                      `json:"total"`
}

// ======= Compliance DTOs =======

// ComplianceFrameworkResponse represents a compliance framework
type ComplianceFrameworkResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Provider    string `json:"provider,omitempty"`
	IsEnabled   bool   `json:"is_enabled"`
}

// ListFrameworksResponse represents a list of frameworks
type ListFrameworksResponse struct {
	Frameworks []ComplianceFrameworkResponse `json:"frameworks"`
}

// ComplianceControlResponse represents a compliance control
type ComplianceControlResponse struct {
	ID           string `json:"id"`
	ControlID    string `json:"control_id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Category     string `json:"category"`
	Severity     string `json:"severity"`
	Remediation  string `json:"remediation,omitempty"`
	ReferenceURL string `json:"reference_url,omitempty"`
}

// ListControlsResponse represents a list of controls
type ListControlsResponse struct {
	Controls []ComplianceControlResponse `json:"controls"`
	Total    int                         `json:"total"`
}

// ComplianceAssessmentResponse represents a compliance assessment
type ComplianceAssessmentResponse struct {
	ID                    string    `json:"id"`
	FrameworkID           string    `json:"framework_id"`
	FrameworkName         string    `json:"framework_name"`
	AssessmentDate        time.Time `json:"assessment_date"`
	TotalControls         int       `json:"total_controls"`
	PassedControls        int       `json:"passed_controls"`
	FailedControls        int       `json:"failed_controls"`
	NotApplicableControls int       `json:"not_applicable_controls"`
	CompliancePercent     float64   `json:"compliance_percent"`
	Status                string    `json:"status"`
}

// ListAssessmentsResponse represents a list of assessments
type ListAssessmentsResponse struct {
	Assessments []ComplianceAssessmentResponse `json:"assessments"`
	Total       int64                          `json:"total"`
}

// RunAssessmentRequest represents a request to run an assessment
type RunAssessmentRequest struct {
	FrameworkID string `json:"framework_id" validate:"required"`
}

// ComplianceOverviewResponse represents a compliance overview
type ComplianceOverviewResponse struct {
	TotalControls      int                      `json:"total_controls"`
	PassedControls     int                      `json:"passed_controls"`
	FailedControls     int                      `json:"failed_controls"`
	CompliancePercent  float64                  `json:"compliance_percent"`
	ByFramework        []FrameworkComplianceDTO `json:"by_framework"`
	TopFailingControls []AssessmentFindingDTO   `json:"top_failing_controls,omitempty"`
	BySeverity         map[string]int           `json:"by_severity"`
}

// FrameworkComplianceDTO represents compliance status for a framework
type FrameworkComplianceDTO struct {
	FrameworkID       string  `json:"framework_id"`
	FrameworkName     string  `json:"framework_name"`
	TotalControls     int     `json:"total_controls"`
	PassedControls    int     `json:"passed_controls"`
	FailedControls    int     `json:"failed_controls"`
	CompliancePercent float64 `json:"compliance_percent"`
	LastAssessment    string  `json:"last_assessment,omitempty"`
}

// AssessmentFindingDTO represents a finding in an assessment
type AssessmentFindingDTO struct {
	ControlID         string   `json:"control_id"`
	ControlTitle      string   `json:"control_title"`
	Category          string   `json:"category"`
	Severity          string   `json:"severity"`
	Status            string   `json:"status"`
	AffectedCount     int      `json:"affected_count"`
	AffectedResources []string `json:"affected_resources,omitempty"`
	Remediation       string   `json:"remediation"`
}

// ComplianceTrendResponse represents compliance trend
type ComplianceTrendResponse struct {
	FrameworkID   string              `json:"framework_id"`
	FrameworkName string              `json:"framework_name"`
	CurrentScore  float64             `json:"current_score"`
	PreviousScore float64             `json:"previous_score"`
	ChangePercent float64             `json:"change_percent"`
	Trend         string              `json:"trend"`
	DataPoints    []TrendDataPointDTO `json:"data_points"`
}

// TrendDataPointDTO represents a trend data point
type TrendDataPointDTO struct {
	Date            string  `json:"date"`
	ComplianceScore float64 `json:"compliance_score"`
	PassedControls  int     `json:"passed_controls"`
	TotalControls   int     `json:"total_controls"`
}
