package dto

import "time"

// IaCDefinitionDTO represents an IaC definition in API responses
type IaCDefinitionDTO struct {
	ID              string                 `json:"id"`
	UserID          string                 `json:"user_id"`
	Name            string                 `json:"name"`
	IaCType         string                 `json:"iac_type"`
	FilePath        string                 `json:"file_path,omitempty"`
	ParsedResources map[string]interface{} `json:"parsed_resources,omitempty"`
	LastParsed      *time.Time             `json:"last_parsed,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// IaCUploadRequest represents a request to upload an IaC file
type IaCUploadRequest struct {
	Name    string `json:"name" validate:"required,min=1,max=255"`
	IaCType string `json:"iac_type" validate:"required,oneof=terraform cloudformation kubernetes helm"`
	Content string `json:"content" validate:"required,min=1"`
}

// IaCResourceDTO represents an IaC resource in API responses
type IaCResourceDTO struct {
	ID              string                 `json:"id"`
	IaCDefinitionID string                 `json:"iac_definition_id"`
	ResourceType    string                 `json:"resource_type"`
	ResourceName    string                 `json:"resource_name"`
	ResourceAddress string                 `json:"resource_address"`
	Provider        string                 `json:"provider"`
	Configuration   map[string]interface{} `json:"configuration,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
}

// IaCDriftResultDTO represents a drift result in API responses
type IaCDriftResultDTO struct {
	ID               string                 `json:"id"`
	IaCDefinitionID  string                 `json:"iac_definition_id"`
	IaCResourceID    *string                `json:"iac_resource_id,omitempty"`
	ActualResourceID *string                `json:"actual_resource_id,omitempty"`
	DriftCategory    string                 `json:"drift_category"`
	Severity         *string                `json:"severity,omitempty"`
	Details          map[string]interface{} `json:"details,omitempty"`
	DetectedAt       time.Time              `json:"detected_at"`
	Status           string                 `json:"status"`
	ResolvedAt       *time.Time             `json:"resolved_at,omitempty"`
}

// IaCDriftDetectRequest represents a request to detect drift
type IaCDriftDetectRequest struct {
	DefinitionID string `json:"definition_id" validate:"required"`
}

// IaCDriftSummaryDTO represents a drift summary in API responses
type IaCDriftSummaryDTO struct {
	Total      int            `json:"total"`
	ByCategory map[string]int `json:"by_category"`
	BySeverity map[string]int `json:"by_severity"`
}

// IaCDriftStatusUpdate represents a request to update drift status
type IaCDriftStatusUpdate struct {
	Status string `json:"status" validate:"required,oneof=detected acknowledged resolved ignored"`
}
