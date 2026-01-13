package cloudformation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Parser handles parsing of CloudFormation templates
type Parser struct{}

// NewParser creates a new CloudFormation parser
func NewParser() *Parser {
	return &Parser{}
}

// ParseFile parses a CloudFormation template file (JSON or YAML)
func (p *Parser) ParseFile(filename string) (*ParseResult, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Determine format based on file extension
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".json":
		return p.ParseJSON(content)
	case ".yaml", ".yml":
		return p.ParseYAML(content)
	default:
		// Try JSON first, then YAML
		result, err := p.ParseJSON(content)
		if err != nil {
			return p.ParseYAML(content)
		}
		return result, nil
	}
}

// ParseJSON parses a CloudFormation JSON template
func (p *Parser) ParseJSON(content []byte) (*ParseResult, error) {
	var template CloudFormationTemplate

	if err := json.Unmarshal(content, &template); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return p.processTemplate(&template)
}

// ParseYAML parses a CloudFormation YAML template
func (p *Parser) ParseYAML(content []byte) (*ParseResult, error) {
	var template CloudFormationTemplate

	if err := yaml.Unmarshal(content, &template); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return p.processTemplate(&template)
}

// processTemplate converts the raw template to our structured format
func (p *Parser) processTemplate(template *CloudFormationTemplate) (*ParseResult, error) {
	result := &ParseResult{
		Parsed: &ParsedCloudFormation{
			Resources:  make([]CloudFormationResource, 0),
			Parameters: make([]CloudFormationParam, 0),
			Outputs:    make([]CloudFormationOut, 0),
			Metadata:   template.Metadata,
		},
		Errors: make([]error, 0),
	}

	// Parse resources
	for logicalID, resource := range template.Resources {
		cfResource, err := p.parseResource(logicalID, resource)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("resource %s: %w", logicalID, err))
			continue
		}
		result.Parsed.Resources = append(result.Parsed.Resources, *cfResource)
	}

	// Parse parameters
	for name, param := range template.Parameters {
		cfParam := CloudFormationParam{
			Name:        name,
			Type:        param.Type,
			Default:     param.Default,
			Description: param.Description,
		}
		result.Parsed.Parameters = append(result.Parsed.Parameters, cfParam)
	}

	// Parse outputs
	for name, output := range template.Outputs {
		cfOutput := CloudFormationOut{
			Name:        name,
			Value:       output.Value,
			Description: output.Description,
		}

		// Extract export name if present
		if output.Export != nil {
			if exportName, ok := output.Export["Name"]; ok {
				if exportStr, ok := exportName.(string); ok {
					cfOutput.Export = exportStr
				}
			}
		}

		result.Parsed.Outputs = append(result.Parsed.Outputs, cfOutput)
	}

	return result, nil
}

// parseResource parses a single CloudFormation resource
func (p *Parser) parseResource(logicalID string, resource CFResource) (*CloudFormationResource, error) {
	if resource.Type == "" {
		return nil, fmt.Errorf("resource type is required")
	}

	cfResource := &CloudFormationResource{
		LogicalID:      logicalID,
		Type:           resource.Type,
		Provider:       "aws", // CloudFormation is AWS-only
		Properties:     resource.Properties,
		Condition:      resource.Condition,
		DeletionPolicy: resource.DeletionPolicy,
		DependsOn:      make([]string, 0),
	}

	// Parse DependsOn (can be string or array)
	if resource.DependsOn != nil {
		switch deps := resource.DependsOn.(type) {
		case string:
			cfResource.DependsOn = []string{deps}
		case []interface{}:
			for _, dep := range deps {
				if depStr, ok := dep.(string); ok {
					cfResource.DependsOn = append(cfResource.DependsOn, depStr)
				}
			}
		case []string:
			cfResource.DependsOn = deps
		}
	}

	// Resolve intrinsic functions in properties
	resolved := p.resolveIntrinsicFunctions(resource.Properties)
	if props, ok := resolved.(map[string]interface{}); ok {
		cfResource.Properties = props
	} else {
		cfResource.Properties = make(map[string]interface{})
	}

	return cfResource, nil
}

// resolveIntrinsicFunctions attempts to resolve CloudFormation intrinsic functions
// This is a simplified version - full resolution would require evaluating parameters and conditions
func (p *Parser) resolveIntrinsicFunctions(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		// Check for intrinsic functions
		if len(v) == 1 {
			for key, value := range v {
				switch key {
				case "Ref":
					// Reference to parameter or resource
					if refStr, ok := value.(string); ok {
						return map[string]interface{}{
							"_ref": refStr,
						}
					}

				case "Fn::GetAtt":
					// Get attribute from another resource
					return map[string]interface{}{
						"_getatt": value,
					}

				case "Fn::Sub":
					// String substitution
					return map[string]interface{}{
						"_sub": value,
					}

				case "Fn::Join":
					// Join strings
					if joinData, ok := value.([]interface{}); ok && len(joinData) == 2 {
						delimiter := ""
						if d, ok := joinData[0].(string); ok {
							delimiter = d
						}
						values := []interface{}{}
						if v, ok := joinData[1].([]interface{}); ok {
							values = v
						}
						return map[string]interface{}{
							"_join": map[string]interface{}{
								"delimiter": delimiter,
								"values":    values,
							},
						}
					}

				case "Fn::Select":
					// Select from array
					return map[string]interface{}{
						"_select": value,
					}

				case "Fn::GetAZs":
					// Get availability zones
					return map[string]interface{}{
						"_getazs": value,
					}

				case "Fn::FindInMap":
					// Find in mapping
					return map[string]interface{}{
						"_findinmap": value,
					}

				case "Fn::Base64":
					// Base64 encoding
					return map[string]interface{}{
						"_base64": value,
					}

				case "Fn::If":
					// Conditional
					return map[string]interface{}{
						"_if": value,
					}
				}
			}
		}

		// Recursively process nested maps
		result := make(map[string]interface{})
		for k, val := range v {
			result[k] = p.resolveIntrinsicFunctions(val)
		}
		return result

	case []interface{}:
		// Recursively process arrays
		result := make([]interface{}, len(v))
		for i, val := range v {
			result[i] = p.resolveIntrinsicFunctions(val)
		}
		return result

	default:
		// Return primitive values as-is
		return v
	}
}

// GetResourceByLogicalID finds a resource by its logical ID
func (p *Parser) GetResourceByLogicalID(parsed *ParsedCloudFormation, logicalID string) *CloudFormationResource {
	for i := range parsed.Resources {
		if parsed.Resources[i].LogicalID == logicalID {
			return &parsed.Resources[i]
		}
	}
	return nil
}

// GetResourcesByType returns all resources of a specific type
func (p *Parser) GetResourcesByType(parsed *ParsedCloudFormation, resourceType string) []CloudFormationResource {
	resources := make([]CloudFormationResource, 0)
	for _, res := range parsed.Resources {
		if res.Type == resourceType {
			resources = append(resources, res)
		}
	}
	return resources
}

// ValidateTemplate performs basic validation on the template
func (p *Parser) ValidateTemplate(template *CloudFormationTemplate) []error {
	errors := make([]error, 0)

	// Check for required fields
	if len(template.Resources) == 0 {
		errors = append(errors, fmt.Errorf("template must contain at least one resource"))
	}

	// Validate resource types
	for logicalID, resource := range template.Resources {
		if resource.Type == "" {
			errors = append(errors, fmt.Errorf("resource %s: Type is required", logicalID))
		}

		// Check if resource type is valid (basic check)
		if !strings.HasPrefix(resource.Type, "AWS::") &&
		   !strings.HasPrefix(resource.Type, "Custom::") &&
		   !strings.HasPrefix(resource.Type, "Alexa::") {
			errors = append(errors, fmt.Errorf("resource %s: invalid resource type %s", logicalID, resource.Type))
		}
	}

	// Validate dependencies
	resourceMap := make(map[string]bool)
	for logicalID := range template.Resources {
		resourceMap[logicalID] = true
	}

	for logicalID, resource := range template.Resources {
		if resource.DependsOn != nil {
			var deps []string
			switch d := resource.DependsOn.(type) {
			case string:
				deps = []string{d}
			case []interface{}:
				for _, dep := range d {
					if depStr, ok := dep.(string); ok {
						deps = append(deps, depStr)
					}
				}
			case []string:
				deps = d
			}

			for _, dep := range deps {
				if !resourceMap[dep] {
					errors = append(errors, fmt.Errorf("resource %s: DependsOn references non-existent resource %s", logicalID, dep))
				}
			}
		}
	}

	return errors
}
