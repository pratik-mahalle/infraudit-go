package dto

import "time"

// IaCDefinitionDTO represents an IaC definition in API responses
type IaCDefinitionDTO struct {
	ID              string                 `json:"id"`
	UserID          string                 `json:"userId"`
	Name            string                 `json:"name"`
	IaCType         string                 `json:"iacType"`
	FilePath        string                 `json:"filePath,omitempty"`
	ParsedResources map[string]interface{} `json:"parsedResources,omitempty"`
	LastParsed      *time.Time             `json:"lastParsed,omitempty"`
	CreatedAt       time.Time              `json:"createdAt"`
	UpdatedAt       time.Time              `json:"updatedAt"`
}

// IaCUploadRequest represents a request to upload an IaC file
type IaCUploadRequest struct {
	Name    string `json:"name" validate:"required,min=1,max=255"`
	IaCType string `json:"iacType" validate:"required,oneof=terraform cloudformation kubernetes helm"`
	Content string `json:"content" validate:"required,min=1"`
}

// IaCResourceDTO represents an IaC resource in API responses
type IaCResourceDTO struct {
	ID              string                 `json:"id"`
	IaCDefinitionID string                 `json:"iacDefinitionId"`
	ResourceType    string                 `json:"resourceType"`
	ResourceName    string                 `json:"resourceName"`
	ResourceAddress string                 `json:"resourceAddress"`
	Provider        string                 `json:"provider"`
	Configuration   map[string]interface{} `json:"configuration,omitempty"`
	CreatedAt       time.Time              `json:"createdAt"`
}

// IaCDriftResultDTO represents a drift result in API responses
type IaCDriftResultDTO struct {
	ID               string                 `json:"id"`
	IaCDefinitionID  string                 `json:"iacDefinitionId"`
	IaCResourceID    *string                `json:"iacResourceId,omitempty"`
	ActualResourceID *string                `json:"actualResourceId,omitempty"`
	DriftCategory    string                 `json:"driftCategory"`
	Severity         *string                `json:"severity,omitempty"`
	Details          map[string]interface{} `json:"details,omitempty"`
	DetectedAt       time.Time              `json:"detectedAt"`
	Status           string                 `json:"status"`
	ResolvedAt       *time.Time             `json:"resolvedAt,omitempty"`
}

// IaCDriftDetectRequest represents a request to detect drift
type IaCDriftDetectRequest struct {
	DefinitionID string `json:"definitionId" validate:"required"`
}

// IaCDriftSummaryDTO represents a drift summary in API responses
type IaCDriftSummaryDTO struct {
	Total      int            `json:"total"`
	ByCategory map[string]int `json:"byCategory"`
	BySeverity map[string]int `json:"bySeverity"`
}

// IaCDriftStatusUpdate represents a request to update drift status
type IaCDriftStatusUpdate struct {
	Status string `json:"status" validate:"required,oneof=detected acknowledged resolved ignored"`
}
