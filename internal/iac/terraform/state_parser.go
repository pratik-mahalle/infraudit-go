package terraform

import (
	"encoding/json"
	"fmt"
	"os"
)

// TerraformState represents a Terraform state file structure
type TerraformState struct {
	Version          int                      `json:"version"`
	TerraformVersion string                   `json:"terraform_version"`
	Serial           int                      `json:"serial"`
	Lineage          string                   `json:"lineage"`
	Outputs          map[string]StateOutput   `json:"outputs,omitempty"`
	Resources        []StateResource          `json:"resources"`
}

// StateOutput represents an output value in the state
type StateOutput struct {
	Value     interface{} `json:"value"`
	Type      interface{} `json:"type,omitempty"`
	Sensitive bool        `json:"sensitive,omitempty"`
}

// StateResource represents a resource in the Terraform state
type StateResource struct {
	Module         string                 `json:"module,omitempty"`
	Mode           string                 `json:"mode"` // "managed" or "data"
	Type           string                 `json:"type"`
	Name           string                 `json:"name"`
	Provider       string                 `json:"provider"`
	Instances      []StateInstance        `json:"instances"`
	EachMode       string                 `json:"each,omitempty"` // "list" or "map"
}

// StateInstance represents an instance of a resource
type StateInstance struct {
	SchemaVersion    int                    `json:"schema_version"`
	Attributes       map[string]interface{} `json:"attributes"`
	AttributesFlat   map[string]string      `json:"attributes_flat,omitempty"`
	Private          string                 `json:"private,omitempty"`
	Dependencies     []string               `json:"dependencies,omitempty"`
	IndexKey         interface{}            `json:"index_key,omitempty"` // For count/for_each
	Status           string                 `json:"status,omitempty"`
	Deposed          string                 `json:"deposed,omitempty"`
}

// StateParser handles parsing of Terraform state files
type StateParser struct{}

// NewStateParser creates a new Terraform state parser
func NewStateParser() *StateParser {
	return &StateParser{}
}

// ParseStateFile parses a Terraform state file
func (sp *StateParser) ParseStateFile(filename string) (*TerraformState, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	return sp.ParseState(content)
}

// ParseState parses Terraform state JSON content
func (sp *StateParser) ParseState(content []byte) (*TerraformState, error) {
	var state TerraformState

	if err := json.Unmarshal(content, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state file: %w", err)
	}

	// Validate state version
	if state.Version < 3 {
		return nil, fmt.Errorf("unsupported state version: %d (minimum supported: 3)", state.Version)
	}

	return &state, nil
}

// ExtractResources extracts resources from the state and converts them to TerraformResource format
func (sp *StateParser) ExtractResources(state *TerraformState) ([]TerraformResource, error) {
	resources := make([]TerraformResource, 0)

	for _, stateRes := range state.Resources {
		// Skip data sources (mode == "data") or include them based on requirements
		if stateRes.Mode != "managed" {
			continue
		}

		// Extract provider name from provider string (e.g., "provider[\"registry.terraform.io/hashicorp/aws\"]" -> "aws")
		provider := sp.extractProviderName(stateRes.Provider)

		// Handle resources with multiple instances (count/for_each)
		for idx, instance := range stateRes.Instances {
			resourceName := stateRes.Name

			// Handle count index
			if instance.IndexKey != nil {
				switch v := instance.IndexKey.(type) {
				case float64:
					resourceName = fmt.Sprintf("%s[%d]", stateRes.Name, int(v))
				case string:
					resourceName = fmt.Sprintf("%s[\"%s\"]", stateRes.Name, v)
				}
			}

			// Build resource address
			address := fmt.Sprintf("%s.%s", stateRes.Type, resourceName)
			if stateRes.Module != "" {
				address = fmt.Sprintf("%s.%s", stateRes.Module, address)
			}

			resource := TerraformResource{
				Type:          stateRes.Type,
				Name:          resourceName,
				Address:       address,
				Provider:      provider,
				Configuration: instance.Attributes,
				Dependencies:  instance.Dependencies,
			}

			// Add count metadata if this is from a count
			if instance.IndexKey != nil {
				if intKey, ok := instance.IndexKey.(float64); ok {
					resource.Count = int(intKey)
				}
			}

			// Filter out computed-only attributes that shouldn't be compared
			resource.Configuration = sp.filterComputedAttributes(stateRes.Type, instance.Attributes)

			resources = append(resources, resource)

			// If there are multiple instances, we've already handled the indexing above
			_ = idx
		}
	}

	return resources, nil
}

// extractProviderName extracts the provider name from the provider string
// e.g., "provider[\"registry.terraform.io/hashicorp/aws\"]" -> "aws"
func (sp *StateParser) extractProviderName(providerStr string) string {
	// Try to extract from the standard format
	if len(providerStr) == 0 {
		return "unknown"
	}

	// Parse provider string format: provider["registry.terraform.io/hashicorp/aws"]
	// We want to extract "aws"
	start := -1
	end := -1

	for i := len(providerStr) - 1; i >= 0; i-- {
		if providerStr[i] == '/' {
			start = i + 1
			break
		}
	}

	for i := len(providerStr) - 1; i >= 0; i-- {
		if providerStr[i] == '"' || providerStr[i] == ']' {
			end = i
			break
		}
	}

	if start > 0 && end > start {
		provider := providerStr[start:end]
		return normalizeProviderName(provider)
	}

	return "unknown"
}

// normalizeProviderName normalizes provider names
func normalizeProviderName(provider string) string {
	switch provider {
	case "google", "google-beta":
		return "gcp"
	case "azurerm", "azuread", "azapi":
		return "azure"
	default:
		return provider
	}
}

// filterComputedAttributes removes computed-only attributes that shouldn't be compared
// These are attributes that are generated by the provider and not configurable
func (sp *StateParser) filterComputedAttributes(resourceType string, attrs map[string]interface{}) map[string]interface{} {
	filtered := make(map[string]interface{})

	// Define computed-only attributes per resource type
	computedOnlyAttrs := map[string][]string{
		"aws_instance": {
			"arn",
			"id",
			"public_ip",
			"public_dns",
			"private_ip",
			"private_dns",
			"outpost_arn",
			"primary_network_interface_id",
			"instance_state",
		},
		"aws_s3_bucket": {
			"id",
			"arn",
			"bucket_domain_name",
			"bucket_regional_domain_name",
			"hosted_zone_id",
			"region",
		},
		"google_compute_instance": {
			"id",
			"instance_id",
			"self_link",
			"cpu_platform",
			"current_status",
			"label_fingerprint",
		},
		"google_storage_bucket": {
			"id",
			"self_link",
			"url",
		},
		"azurerm_virtual_machine": {
			"id",
		},
		"azurerm_linux_virtual_machine": {
			"id",
			"virtual_machine_id",
			"private_ip_address",
			"public_ip_address",
		},
		"azurerm_windows_virtual_machine": {
			"id",
			"virtual_machine_id",
			"private_ip_address",
			"public_ip_address",
		},
	}

	// Get the list of computed-only attributes for this resource type
	computedAttrs, exists := computedOnlyAttrs[resourceType]
	if !exists {
		// If no specific list, just filter common computed attributes
		computedAttrs = []string{"id", "arn", "self_link"}
	}

	// Create a map for quick lookup
	computedMap := make(map[string]bool)
	for _, attr := range computedAttrs {
		computedMap[attr] = true
	}

	// Filter out computed-only attributes
	for key, value := range attrs {
		if !computedMap[key] {
			filtered[key] = value
		}
	}

	return filtered
}

// GetResourceByAddress finds a resource in the state by its address
func (sp *StateParser) GetResourceByAddress(state *TerraformState, address string) (*TerraformResource, error) {
	resources, err := sp.ExtractResources(state)
	if err != nil {
		return nil, err
	}

	for _, res := range resources {
		if res.Address == address {
			return &res, nil
		}
	}

	return nil, fmt.Errorf("resource not found: %s", address)
}

// CompareWithHCL compares resources from state file with HCL configuration
// This helps identify differences between desired state (HCL) and actual state (state file)
func (sp *StateParser) CompareWithHCL(state *TerraformState, hclResources []TerraformResource) ([]DriftComparison, error) {
	stateResources, err := sp.ExtractResources(state)
	if err != nil {
		return nil, err
	}

	comparisons := make([]DriftComparison, 0)

	// Create maps for quick lookup
	stateMap := make(map[string]*TerraformResource)
	for i := range stateResources {
		stateMap[stateResources[i].Address] = &stateResources[i]
	}

	hclMap := make(map[string]*TerraformResource)
	for i := range hclResources {
		hclMap[hclResources[i].Address] = &hclResources[i]
	}

	// Compare HCL resources with state
	for address, hclRes := range hclMap {
		comparison := DriftComparison{
			Address:      address,
			ResourceType: hclRes.Type,
			ResourceName: hclRes.Name,
			HCLConfig:    hclRes.Configuration,
		}

		if stateRes, exists := stateMap[address]; exists {
			comparison.StateConfig = stateRes.Configuration
			comparison.ExistsInState = true
			comparison.ExistsInHCL = true

			// Compare configurations
			mapper := NewResourceMapper()
			iacDiffs := mapper.CompareResources(hclRes.Configuration, stateRes.Configuration)

			// Convert iac.FieldChange to terraform.FieldChange
			comparison.Differences = make([]FieldChange, len(iacDiffs))
			for i, diff := range iacDiffs {
				comparison.Differences[i] = FieldChange{
					Field:      diff.Field,
					HCLValue:   diff.IaCValue,
					StateValue: diff.ActualValue,
					ChangeType: diff.ChangeType,
				}
			}
			comparison.HasDrift = len(comparison.Differences) > 0
		} else {
			comparison.ExistsInState = false
			comparison.ExistsInHCL = true
			comparison.HasDrift = true
			comparison.DriftType = "missing_in_state"
		}

		comparisons = append(comparisons, comparison)
	}

	// Check for resources in state but not in HCL (shadow resources)
	for address, stateRes := range stateMap {
		if _, exists := hclMap[address]; !exists {
			comparisons = append(comparisons, DriftComparison{
				Address:        address,
				ResourceType:   stateRes.Type,
				ResourceName:   stateRes.Name,
				StateConfig:    stateRes.Configuration,
				ExistsInState:  true,
				ExistsInHCL:    false,
				HasDrift:       true,
				DriftType:      "not_in_hcl",
			})
		}
	}

	return comparisons, nil
}

// DriftComparison represents a comparison between HCL and state
type DriftComparison struct {
	Address        string                   `json:"address"`
	ResourceType   string                   `json:"resource_type"`
	ResourceName   string                   `json:"resource_name"`
	HCLConfig      map[string]interface{}   `json:"hcl_config,omitempty"`
	StateConfig    map[string]interface{}   `json:"state_config,omitempty"`
	ExistsInHCL    bool                     `json:"exists_in_hcl"`
	ExistsInState  bool                     `json:"exists_in_state"`
	HasDrift       bool                     `json:"has_drift"`
	DriftType      string                   `json:"drift_type,omitempty"` // missing_in_state, not_in_hcl, config_diff
	Differences    []FieldChange            `json:"differences,omitempty"`
}

// FieldChange represents a configuration difference
type FieldChange struct {
	Field       string      `json:"field"`
	HCLValue    interface{} `json:"hcl_value"`
	StateValue  interface{} `json:"state_value"`
	ChangeType  string      `json:"change_type"` // added, removed, modified
}
