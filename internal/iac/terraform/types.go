package terraform

import "github.com/hashicorp/hcl/v2"

// TerraformResource represents a parsed Terraform resource
type TerraformResource struct {
	Type          string                 `json:"type"`           // e.g., "aws_instance"
	Name          string                 `json:"name"`           // e.g., "web_server"
	Address       string                 `json:"address"`        // e.g., "aws_instance.web_server"
	Provider      string                 `json:"provider"`       // e.g., "aws"
	Configuration map[string]interface{} `json:"configuration"`  // Resource attributes
	Dependencies  []string               `json:"dependencies,omitempty"`
	Count         interface{}            `json:"count,omitempty"`    // Can be number or expression
	ForEach       interface{}            `json:"for_each,omitempty"` // For dynamic resources
	DependsOn     []string               `json:"depends_on,omitempty"`
	Lifecycle     *Lifecycle             `json:"lifecycle,omitempty"`
	Provisioners  []Provisioner          `json:"provisioners,omitempty"`
}

// Lifecycle represents the Terraform lifecycle meta-argument
type Lifecycle struct {
	CreateBeforeDestroy bool     `json:"create_before_destroy,omitempty"`
	PreventDestroy      bool     `json:"prevent_destroy,omitempty"`
	IgnoreChanges       []string `json:"ignore_changes,omitempty"`
}

// Provisioner represents a Terraform provisioner
type Provisioner struct {
	Type       string                 `json:"type"` // e.g., "remote-exec", "local-exec"
	Connection map[string]interface{} `json:"connection,omitempty"`
	Config     map[string]interface{} `json:"config"`
}

// TerraformModule represents a Terraform module
type TerraformModule struct {
	Name          string                 `json:"name"`
	Source        string                 `json:"source"`
	Version       string                 `json:"version,omitempty"`
	Configuration map[string]interface{} `json:"configuration"`
}

// TerraformVariable represents a Terraform variable
type TerraformVariable struct {
	Name        string      `json:"name"`
	Type        string      `json:"type,omitempty"`
	Description string      `json:"description,omitempty"`
	Default     interface{} `json:"default,omitempty"`
	Validation  []Validation `json:"validation,omitempty"`
	Sensitive   bool        `json:"sensitive,omitempty"`
}

// Validation represents a variable validation rule
type Validation struct {
	Condition    string `json:"condition"`
	ErrorMessage string `json:"error_message"`
}

// TerraformOutput represents a Terraform output
type TerraformOutput struct {
	Name        string      `json:"name"`
	Value       interface{} `json:"value"`
	Description string      `json:"description,omitempty"`
	Sensitive   bool        `json:"sensitive,omitempty"`
}

// TerraformProvider represents a provider configuration
type TerraformProvider struct {
	Name          string                 `json:"name"`    // e.g., "aws"
	Alias         string                 `json:"alias,omitempty"`
	Configuration map[string]interface{} `json:"configuration"`
}

// TerraformData represents a data source
type TerraformData struct {
	Type          string                 `json:"type"`    // e.g., "aws_ami"
	Name          string                 `json:"name"`    // e.g., "ubuntu"
	Address       string                 `json:"address"` // e.g., "data.aws_ami.ubuntu"
	Provider      string                 `json:"provider"`
	Configuration map[string]interface{} `json:"configuration"`
}

// ParsedTerraform represents a fully parsed Terraform configuration
type ParsedTerraform struct {
	Resources []TerraformResource `json:"resources"`
	Modules   []TerraformModule   `json:"modules,omitempty"`
	Variables []TerraformVariable `json:"variables,omitempty"`
	Outputs   []TerraformOutput   `json:"outputs,omitempty"`
	Providers []TerraformProvider `json:"providers,omitempty"`
	DataSources []TerraformData   `json:"data_sources,omitempty"`
	Terraform *TerraformBlock     `json:"terraform,omitempty"`
}

// TerraformBlock represents the terraform {} block
type TerraformBlock struct {
	RequiredVersion string                     `json:"required_version,omitempty"`
	RequiredProviders map[string]ProviderRequirement `json:"required_providers,omitempty"`
	Backend         *Backend                    `json:"backend,omitempty"`
}

// ProviderRequirement represents a required provider configuration
type ProviderRequirement struct {
	Source  string `json:"source"`
	Version string `json:"version,omitempty"`
}

// Backend represents the terraform backend configuration
type Backend struct {
	Type          string                 `json:"type"` // e.g., "s3", "gcs"
	Configuration map[string]interface{} `json:"configuration"`
}

// ParseResult holds the result of parsing a Terraform file
type ParseResult struct {
	Parsed      *ParsedTerraform
	Diagnostics hcl.Diagnostics
	Errors      []error
}

// HasErrors returns true if there are any errors
func (r *ParseResult) HasErrors() bool {
	if r == nil {
		return false
	}
	return len(r.Errors) > 0 || r.Diagnostics.HasErrors()
}

// ErrorMessages returns all error messages
func (r *ParseResult) ErrorMessages() []string {
	if r == nil {
		return nil
	}
	messages := make([]string, 0)
	for _, err := range r.Errors {
		messages = append(messages, err.Error())
	}
	for _, diag := range r.Diagnostics {
		if diag.Severity == hcl.DiagError {
			messages = append(messages, diag.Error())
		}
	}
	return messages
}
