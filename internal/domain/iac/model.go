package iac

import (
	"encoding/json"
	"time"
)

// IaCType represents the type of Infrastructure as Code
type IaCType string

const (
	IaCTypeTerraform      IaCType = "terraform"
	IaCTypeCloudFormation IaCType = "cloudformation"
	IaCTypeKubernetes     IaCType = "kubernetes"
	IaCTypeHelm           IaCType = "helm"
)

// DriftCategory represents the category of IaC drift
type DriftCategory string

const (
	DriftCategoryMissing   DriftCategory = "missing"   // In IaC but not deployed
	DriftCategoryShadow    DriftCategory = "shadow"    // Deployed but not in IaC
	DriftCategoryModified  DriftCategory = "modified"  // Configuration mismatch
	DriftCategoryCompliant DriftCategory = "compliant" // No drift detected
)

// DriftStatus represents the status of a drift finding
type DriftStatus string

const (
	DriftStatusDetected      DriftStatus = "detected"
	DriftStatusAcknowledged  DriftStatus = "acknowledged"
	DriftStatusResolved      DriftStatus = "resolved"
	DriftStatusIgnored       DriftStatus = "ignored"
)

// Severity represents the severity level
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
)

// IaCDefinition represents an uploaded IaC file
type IaCDefinition struct {
	ID              string                 `json:"id"`
	UserID          string                 `json:"user_id"`
	Name            string                 `json:"name"`
	IaCType         IaCType                `json:"iac_type"`
	FilePath        string                 `json:"file_path,omitempty"`
	Content         string                 `json:"content"`
	ParsedResources map[string]interface{} `json:"parsed_resources,omitempty"`
	LastParsed      *time.Time             `json:"last_parsed,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// IaCResource represents a resource parsed from an IaC file
type IaCResource struct {
	ID                string                 `json:"id"`
	IaCDefinitionID   string                 `json:"iac_definition_id"`
	UserID            string                 `json:"user_id"`
	ResourceType      string                 `json:"resource_type"`
	ResourceName      string                 `json:"resource_name"`
	ResourceAddress   string                 `json:"resource_address"` // Full address (e.g., module.vpc.aws_instance.web)
	Provider          string                 `json:"provider"`
	Configuration     map[string]interface{} `json:"configuration"`
	CreatedAt         time.Time              `json:"created_at"`
}

// IaCDriftResult represents a drift comparison result
type IaCDriftResult struct {
	ID               string                 `json:"id"`
	UserID           string                 `json:"user_id"`
	IaCDefinitionID  string                 `json:"iac_definition_id"`
	IaCResourceID    *string                `json:"iac_resource_id,omitempty"`
	ActualResourceID *string                `json:"actual_resource_id,omitempty"`
	DriftCategory    DriftCategory          `json:"drift_category"`
	Severity         *Severity              `json:"severity,omitempty"`
	Details          map[string]interface{} `json:"details,omitempty"`
	DetectedAt       time.Time              `json:"detected_at"`
	Status           DriftStatus            `json:"status"`
	ResolvedAt       *time.Time             `json:"resolved_at,omitempty"`
}

// ParsedResource represents a generic parsed resource structure
type ParsedResource struct {
	Type          string                 `json:"type"`
	Name          string                 `json:"name"`
	Address       string                 `json:"address"`
	Provider      string                 `json:"provider"`
	Configuration map[string]interface{} `json:"configuration"`
	Dependencies  []string               `json:"dependencies,omitempty"`
	Count         *int                   `json:"count,omitempty"`
	ForEach       interface{}            `json:"for_each,omitempty"`
}

// DriftDetails contains detailed information about a drift
type DriftDetails struct {
	IaCValue      interface{}            `json:"iac_value,omitempty"`
	ActualValue   interface{}            `json:"actual_value,omitempty"`
	Changes       []FieldChange          `json:"changes,omitempty"`
	Message       string                 `json:"message"`
	Recommendation string                `json:"recommendation,omitempty"`
}

// FieldChange represents a change in a specific field
type FieldChange struct {
	Field     string      `json:"field"`
	IaCValue  interface{} `json:"iac_value"`
	ActualValue interface{} `json:"actual_value"`
	ChangeType string      `json:"change_type"` // added, removed, modified
}

// MarshalJSON custom marshal for IaCDefinition
func (d *IaCDefinition) MarshalJSON() ([]byte, error) {
	type Alias IaCDefinition
	return json.Marshal(&struct {
		*Alias
		ParsedResources json.RawMessage `json:"parsed_resources,omitempty"`
	}{
		Alias:           (*Alias)(d),
		ParsedResources: marshalToRawJSON(d.ParsedResources),
	})
}

// MarshalJSON custom marshal for IaCResource
func (r *IaCResource) MarshalJSON() ([]byte, error) {
	type Alias IaCResource
	return json.Marshal(&struct {
		*Alias
		Configuration json.RawMessage `json:"configuration"`
	}{
		Alias:         (*Alias)(r),
		Configuration: marshalToRawJSON(r.Configuration),
	})
}

// MarshalJSON custom marshal for IaCDriftResult
func (d *IaCDriftResult) MarshalJSON() ([]byte, error) {
	type Alias IaCDriftResult
	return json.Marshal(&struct {
		*Alias
		Details json.RawMessage `json:"details,omitempty"`
	}{
		Alias:   (*Alias)(d),
		Details: marshalToRawJSON(d.Details),
	})
}

// Helper function to marshal map to JSON
func marshalToRawJSON(data interface{}) json.RawMessage {
	if data == nil {
		return nil
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil
	}
	return bytes
}

// Validate validates the IaCDefinition
func (d *IaCDefinition) Validate() error {
	if d.UserID == "" {
		return ErrMissingUserID
	}
	if d.Name == "" {
		return ErrMissingName
	}
	if d.IaCType == "" {
		return ErrMissingIaCType
	}
	if d.Content == "" {
		return ErrMissingContent
	}

	// Validate IaC type
	validTypes := []IaCType{IaCTypeTerraform, IaCTypeCloudFormation, IaCTypeKubernetes, IaCTypeHelm}
	valid := false
	for _, t := range validTypes {
		if d.IaCType == t {
			valid = true
			break
		}
	}
	if !valid {
		return ErrInvalidIaCType
	}

	return nil
}

// Validate validates the IaCResource
func (r *IaCResource) Validate() error {
	if r.UserID == "" {
		return ErrMissingUserID
	}
	if r.IaCDefinitionID == "" {
		return ErrMissingDefinitionID
	}
	if r.ResourceType == "" {
		return ErrMissingResourceType
	}
	if r.ResourceName == "" {
		return ErrMissingResourceName
	}
	if r.Provider == "" {
		return ErrMissingProvider
	}
	return nil
}

// Validate validates the IaCDriftResult
func (d *IaCDriftResult) Validate() error {
	if d.UserID == "" {
		return ErrMissingUserID
	}
	if d.IaCDefinitionID == "" {
		return ErrMissingDefinitionID
	}
	if d.DriftCategory == "" {
		return ErrMissingDriftCategory
	}

	// Validate drift category
	validCategories := []DriftCategory{DriftCategoryMissing, DriftCategoryShadow, DriftCategoryModified, DriftCategoryCompliant}
	valid := false
	for _, c := range validCategories {
		if d.DriftCategory == c {
			valid = true
			break
		}
	}
	if !valid {
		return ErrInvalidDriftCategory
	}

	return nil
}
