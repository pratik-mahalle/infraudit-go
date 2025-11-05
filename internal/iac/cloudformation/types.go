package cloudformation

// CloudFormationTemplate represents a parsed CloudFormation template
type CloudFormationTemplate struct {
	AWSTemplateFormatVersion string                            `json:"AWSTemplateFormatVersion" yaml:"AWSTemplateFormatVersion"`
	Description              string                            `json:"Description,omitempty" yaml:"Description,omitempty"`
	Metadata                 map[string]interface{}            `json:"Metadata,omitempty" yaml:"Metadata,omitempty"`
	Parameters               map[string]CFParameter            `json:"Parameters,omitempty" yaml:"Parameters,omitempty"`
	Mappings                 map[string]interface{}            `json:"Mappings,omitempty" yaml:"Mappings,omitempty"`
	Conditions               map[string]interface{}            `json:"Conditions,omitempty" yaml:"Conditions,omitempty"`
	Transform                interface{}                       `json:"Transform,omitempty" yaml:"Transform,omitempty"`
	Resources                map[string]CFResource             `json:"Resources" yaml:"Resources"`
	Outputs                  map[string]CFOutput               `json:"Outputs,omitempty" yaml:"Outputs,omitempty"`
}

// CFParameter represents a CloudFormation parameter
type CFParameter struct {
	Type                  string      `json:"Type" yaml:"Type"`
	Default               interface{} `json:"Default,omitempty" yaml:"Default,omitempty"`
	AllowedValues         []string    `json:"AllowedValues,omitempty" yaml:"AllowedValues,omitempty"`
	AllowedPattern        string      `json:"AllowedPattern,omitempty" yaml:"AllowedPattern,omitempty"`
	ConstraintDescription string      `json:"ConstraintDescription,omitempty" yaml:"ConstraintDescription,omitempty"`
	Description           string      `json:"Description,omitempty" yaml:"Description,omitempty"`
	MaxLength             *int        `json:"MaxLength,omitempty" yaml:"MaxLength,omitempty"`
	MinLength             *int        `json:"MinLength,omitempty" yaml:"MinLength,omitempty"`
	MaxValue              *int        `json:"MaxValue,omitempty" yaml:"MaxValue,omitempty"`
	MinValue              *int        `json:"MinValue,omitempty" yaml:"MinValue,omitempty"`
	NoEcho                bool        `json:"NoEcho,omitempty" yaml:"NoEcho,omitempty"`
}

// CFResource represents a CloudFormation resource
type CFResource struct {
	Type           string                 `json:"Type" yaml:"Type"`
	Properties     map[string]interface{} `json:"Properties,omitempty" yaml:"Properties,omitempty"`
	DependsOn      interface{}            `json:"DependsOn,omitempty" yaml:"DependsOn,omitempty"` // Can be string or []string
	Condition      string                 `json:"Condition,omitempty" yaml:"Condition,omitempty"`
	DeletionPolicy string                 `json:"DeletionPolicy,omitempty" yaml:"DeletionPolicy,omitempty"`
	UpdatePolicy   map[string]interface{} `json:"UpdatePolicy,omitempty" yaml:"UpdatePolicy,omitempty"`
	Metadata       map[string]interface{} `json:"Metadata,omitempty" yaml:"Metadata,omitempty"`
}

// CFOutput represents a CloudFormation output
type CFOutput struct {
	Description string                 `json:"Description,omitempty" yaml:"Description,omitempty"`
	Value       interface{}            `json:"Value" yaml:"Value"`
	Export      map[string]interface{} `json:"Export,omitempty" yaml:"Export,omitempty"`
	Condition   string                 `json:"Condition,omitempty" yaml:"Condition,omitempty"`
}

// ParsedCloudFormation represents a structured CloudFormation template
type ParsedCloudFormation struct {
	Resources  []CloudFormationResource `json:"resources"`
	Parameters []CloudFormationParam    `json:"parameters,omitempty"`
	Outputs    []CloudFormationOut      `json:"outputs,omitempty"`
	Metadata   map[string]interface{}   `json:"metadata,omitempty"`
}

// CloudFormationResource represents a parsed CF resource
type CloudFormationResource struct {
	LogicalID     string                 `json:"logical_id"`
	Type          string                 `json:"type"`
	Provider      string                 `json:"provider"` // Always "aws" for CloudFormation
	Properties    map[string]interface{} `json:"properties"`
	DependsOn     []string               `json:"depends_on,omitempty"`
	Condition     string                 `json:"condition,omitempty"`
	DeletionPolicy string                `json:"deletion_policy,omitempty"`
}

// CloudFormationParam represents a parsed parameter
type CloudFormationParam struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Default     interface{} `json:"default,omitempty"`
	Description string      `json:"description,omitempty"`
}

// CloudFormationOut represents a parsed output
type CloudFormationOut struct {
	Name        string      `json:"name"`
	Value       interface{} `json:"value"`
	Description string      `json:"description,omitempty"`
	Export      string      `json:"export,omitempty"`
}

// ParseResult holds the result of parsing a CloudFormation template
type ParseResult struct {
	Parsed *ParsedCloudFormation
	Errors []error
}

// HasErrors returns true if there are any errors
func (r *ParseResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// ErrorMessages returns all error messages
func (r *ParseResult) ErrorMessages() []string {
	messages := make([]string, 0, len(r.Errors))
	for _, err := range r.Errors {
		messages = append(messages, err.Error())
	}
	return messages
}
