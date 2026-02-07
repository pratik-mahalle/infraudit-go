package dto

import (
	"encoding/json"
	"time"
)

// ======= Cost DTOs =======

// CostOverviewResponse represents the cost overview response
type CostOverviewResponse struct {
	TotalCost        float64            `json:"totalCost"`
	MonthlyCost      float64            `json:"monthlyCost"`
	DailyCost        float64            `json:"dailyCost"`
	Currency         string             `json:"currency"`
	ByProvider       map[string]float64 `json:"byProvider"`
	TopServices      []ServiceCostDTO   `json:"topServices"`
	Trend            *CostTrendDTO      `json:"trend,omitempty"`
	AnomalyCount     int                `json:"anomalyCount"`
	PotentialSavings float64            `json:"potentialSavings"`
}

// ServiceCostDTO represents cost for a specific service
type ServiceCostDTO struct {
	Provider    string  `json:"provider,omitempty"`
	ServiceName string  `json:"serviceName"`
	Cost        float64 `json:"cost"`
	Percentage  float64 `json:"percentage"`
}

// CostSummaryResponse represents a cost summary
type CostSummaryResponse struct {
	Provider  string             `json:"provider"`
	TotalCost float64            `json:"totalCost"`
	Currency  string             `json:"currency"`
	Period    string             `json:"period"`
	StartDate string             `json:"startDate"`
	EndDate   string             `json:"endDate"`
	ByService map[string]float64 `json:"byService"`
	ByRegion  map[string]float64 `json:"byRegion"`
}

// CostTrendDTO represents cost trend data
type CostTrendDTO struct {
	Period        string             `json:"period"`
	CurrentCost   float64            `json:"currentCost"`
	PreviousCost  float64            `json:"previousCost"`
	Change        float64            `json:"change"`
	ChangePercent float64            `json:"changePercent"`
	Trend         string             `json:"trend"`
	DataPoints    []CostDataPointDTO `json:"dataPoints"`
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
	ForecastedCost  float64 `json:"forecastedCost"`
	ConfidenceLevel float64 `json:"confidenceLevel"`
	LowerBound      float64 `json:"lowerBound"`
	UpperBound      float64 `json:"upperBound"`
	Currency        string  `json:"currency"`
	EndDate         string  `json:"endDate"`
}

// CostAnomalyResponse represents a cost anomaly
type CostAnomalyResponse struct {
	ID           string    `json:"id"`
	Provider     string    `json:"provider"`
	ServiceName  string    `json:"serviceName"`
	ResourceID   *string   `json:"resourceId,omitempty"`
	AnomalyType  string    `json:"anomalyType"`
	ExpectedCost float64   `json:"expectedCost"`
	ActualCost   float64   `json:"actualCost"`
	Deviation    float64   `json:"deviation"`
	Severity     string    `json:"severity"`
	Status       string    `json:"status"`
	Notes        string    `json:"notes,omitempty"`
	DetectedAt   time.Time `json:"detectedAt"`
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
	ResourceID       *string         `json:"resourceId,omitempty"`
	ResourceType     string          `json:"resourceType"`
	OptimizationType string          `json:"optimizationType"`
	Title            string          `json:"title"`
	Description      string          `json:"description"`
	CurrentCost      float64         `json:"currentCost"`
	EstimatedSavings float64         `json:"estimatedSavings"`
	SavingsPercent   float64         `json:"savingsPercent"`
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
	IsEnabled   bool   `json:"isEnabled"`
}

// ListFrameworksResponse represents a list of frameworks
type ListFrameworksResponse struct {
	Frameworks []ComplianceFrameworkResponse `json:"frameworks"`
}

// ComplianceControlResponse represents a compliance control
type ComplianceControlResponse struct {
	ID           string `json:"id"`
	ControlID    string `json:"controlId"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Category     string `json:"category"`
	Severity     string `json:"severity"`
	Remediation  string `json:"remediation,omitempty"`
	ReferenceURL string `json:"referenceUrl,omitempty"`
}

// ListControlsResponse represents a list of controls
type ListControlsResponse struct {
	Controls []ComplianceControlResponse `json:"controls"`
	Total    int                         `json:"total"`
}

// ComplianceAssessmentResponse represents a compliance assessment
type ComplianceAssessmentResponse struct {
	ID                    string    `json:"id"`
	FrameworkID           string    `json:"frameworkId"`
	FrameworkName         string    `json:"frameworkName"`
	AssessmentDate        time.Time `json:"assessmentDate"`
	TotalControls         int       `json:"totalControls"`
	PassedControls        int       `json:"passedControls"`
	FailedControls        int       `json:"failedControls"`
	NotApplicableControls int       `json:"notApplicableControls"`
	CompliancePercent     float64   `json:"compliancePercent"`
	Status                string    `json:"status"`
}

// ListAssessmentsResponse represents a list of assessments
type ListAssessmentsResponse struct {
	Assessments []ComplianceAssessmentResponse `json:"assessments"`
	Total       int64                          `json:"total"`
}

// RunAssessmentRequest represents a request to run an assessment
type RunAssessmentRequest struct {
	FrameworkID string `json:"frameworkId" validate:"required"`
}

// ComplianceOverviewResponse represents a compliance overview
type ComplianceOverviewResponse struct {
	TotalControls      int                      `json:"totalControls"`
	PassedControls     int                      `json:"passedControls"`
	FailedControls     int                      `json:"failedControls"`
	CompliancePercent  float64                  `json:"compliancePercent"`
	ByFramework        []FrameworkComplianceDTO `json:"byFramework"`
	TopFailingControls []AssessmentFindingDTO   `json:"topFailingControls,omitempty"`
	BySeverity         map[string]int           `json:"bySeverity"`
}

// FrameworkComplianceDTO represents compliance status for a framework
type FrameworkComplianceDTO struct {
	FrameworkID       string  `json:"frameworkId"`
	FrameworkName     string  `json:"frameworkName"`
	TotalControls     int     `json:"totalControls"`
	PassedControls    int     `json:"passedControls"`
	FailedControls    int     `json:"failedControls"`
	CompliancePercent float64 `json:"compliancePercent"`
	LastAssessment    string  `json:"lastAssessment,omitempty"`
}

// AssessmentFindingDTO represents a finding in an assessment
type AssessmentFindingDTO struct {
	ControlID         string   `json:"controlId"`
	ControlTitle      string   `json:"controlTitle"`
	Category          string   `json:"category"`
	Severity          string   `json:"severity"`
	Status            string   `json:"status"`
	AffectedCount     int      `json:"affectedCount"`
	AffectedResources []string `json:"affectedResources,omitempty"`
	Remediation       string   `json:"remediation"`
}

// ComplianceTrendResponse represents compliance trend
type ComplianceTrendResponse struct {
	FrameworkID   string              `json:"frameworkId"`
	FrameworkName string              `json:"frameworkName"`
	CurrentScore  float64             `json:"currentScore"`
	PreviousScore float64             `json:"previousScore"`
	ChangePercent float64             `json:"changePercent"`
	Trend         string              `json:"trend"`
	DataPoints    []TrendDataPointDTO `json:"dataPoints"`
}

// TrendDataPointDTO represents a trend data point
type TrendDataPointDTO struct {
	Date            string  `json:"date"`
	ComplianceScore float64 `json:"complianceScore"`
	PassedControls  int     `json:"passedControls"`
	TotalControls   int     `json:"totalControls"`
}
