package terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// Parser handles parsing of Terraform HCL files
type Parser struct {
	parser *hclparse.Parser
}

// NewParser creates a new Terraform parser
func NewParser() *Parser {
	return &Parser{
		parser: hclparse.NewParser(),
	}
}

// ParseFile parses a single Terraform file
func (p *Parser) ParseFile(filename string) (*ParseResult, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return p.Parse(content, filename)
}

// Parse parses Terraform HCL content
func (p *Parser) Parse(content []byte, filename string) (*ParseResult, error) {
	result := &ParseResult{
		Parsed: &ParsedTerraform{
			Resources:         make([]TerraformResource, 0),
			Modules:           make([]TerraformModule, 0),
			Variables:         make([]TerraformVariable, 0),
			Outputs:           make([]TerraformOutput, 0),
			Providers:         make([]TerraformProvider, 0),
			DataSources:       make([]TerraformData, 0),
		},
		Errors: make([]error, 0),
	}

	// Parse the HCL file
	file, diags := p.parser.ParseHCL(content, filename)
	result.Diagnostics = diags

	if diags.HasErrors() {
		return result, fmt.Errorf("HCL parsing failed: %s", diags.Error())
	}

	if file == nil || file.Body == nil {
		return result, fmt.Errorf("empty HCL file")
	}

	// Parse the body
	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return result, fmt.Errorf("unexpected body type")
	}

	// Parse different block types
	for _, block := range body.Blocks {
		switch block.Type {
		case "resource":
			if resource, err := p.parseResourceBlock(block); err == nil {
				result.Parsed.Resources = append(result.Parsed.Resources, *resource)
			} else {
				result.Errors = append(result.Errors, err)
			}

		case "module":
			if module, err := p.parseModuleBlock(block); err == nil {
				result.Parsed.Modules = append(result.Parsed.Modules, *module)
			} else {
				result.Errors = append(result.Errors, err)
			}

		case "variable":
			if variable, err := p.parseVariableBlock(block); err == nil {
				result.Parsed.Variables = append(result.Parsed.Variables, *variable)
			} else {
				result.Errors = append(result.Errors, err)
			}

		case "output":
			if output, err := p.parseOutputBlock(block); err == nil {
				result.Parsed.Outputs = append(result.Parsed.Outputs, *output)
			} else {
				result.Errors = append(result.Errors, err)
			}

		case "provider":
			if provider, err := p.parseProviderBlock(block); err == nil {
				result.Parsed.Providers = append(result.Parsed.Providers, *provider)
			} else {
				result.Errors = append(result.Errors, err)
			}

		case "data":
			if data, err := p.parseDataBlock(block); err == nil {
				result.Parsed.DataSources = append(result.Parsed.DataSources, *data)
			} else {
				result.Errors = append(result.Errors, err)
			}

		case "terraform":
			if tfBlock, err := p.parseTerraformBlock(block); err == nil {
				result.Parsed.Terraform = tfBlock
			} else {
				result.Errors = append(result.Errors, err)
			}
		}
	}

	return result, nil
}

// ParseDirectory parses all Terraform files in a directory
func (p *Parser) ParseDirectory(dir string) (*ParseResult, error) {
	combinedResult := &ParseResult{
		Parsed: &ParsedTerraform{
			Resources:   make([]TerraformResource, 0),
			Modules:     make([]TerraformModule, 0),
			Variables:   make([]TerraformVariable, 0),
			Outputs:     make([]TerraformOutput, 0),
			Providers:   make([]TerraformProvider, 0),
			DataSources: make([]TerraformData, 0),
		},
		Errors: make([]error, 0),
	}

	// Find all .tf files
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".tf") {
			return nil
		}

		// Parse the file
		result, err := p.ParseFile(path)
		if err != nil {
			combinedResult.Errors = append(combinedResult.Errors, fmt.Errorf("%s: %w", path, err))
			return nil // Continue processing other files
		}

		// Merge results
		combinedResult.Parsed.Resources = append(combinedResult.Parsed.Resources, result.Parsed.Resources...)
		combinedResult.Parsed.Modules = append(combinedResult.Parsed.Modules, result.Parsed.Modules...)
		combinedResult.Parsed.Variables = append(combinedResult.Parsed.Variables, result.Parsed.Variables...)
		combinedResult.Parsed.Outputs = append(combinedResult.Parsed.Outputs, result.Parsed.Outputs...)
		combinedResult.Parsed.Providers = append(combinedResult.Parsed.Providers, result.Parsed.Providers...)
		combinedResult.Parsed.DataSources = append(combinedResult.Parsed.DataSources, result.Parsed.DataSources...)

		// Use the first terraform block found
		if result.Parsed.Terraform != nil && combinedResult.Parsed.Terraform == nil {
			combinedResult.Parsed.Terraform = result.Parsed.Terraform
		}

		combinedResult.Diagnostics = append(combinedResult.Diagnostics, result.Diagnostics...)
		combinedResult.Errors = append(combinedResult.Errors, result.Errors...)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return combinedResult, nil
}

// parseResourceBlock parses a resource block
func (p *Parser) parseResourceBlock(block *hclsyntax.Block) (*TerraformResource, error) {
	if len(block.Labels) < 2 {
		return nil, fmt.Errorf("resource block requires 2 labels (type and name)")
	}

	resourceType := block.Labels[0]
	resourceName := block.Labels[1]

	// Extract provider from resource type (e.g., "aws_instance" -> "aws")
	provider := extractProviderFromType(resourceType)

	resource := &TerraformResource{
		Type:          resourceType,
		Name:          resourceName,
		Address:       fmt.Sprintf("%s.%s", resourceType, resourceName),
		Provider:      provider,
		Configuration: make(map[string]interface{}),
		Dependencies:  make([]string, 0),
	}

	// Parse attributes and nested blocks
	attrs, err := p.parseBlockContent(block.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse resource content: %w", err)
	}

	// Extract special meta-arguments
	if count, ok := attrs["count"]; ok {
		resource.Count = count
		delete(attrs, "count")
	}

	if forEach, ok := attrs["for_each"]; ok {
		resource.ForEach = forEach
		delete(attrs, "for_each")
	}

	if dependsOn, ok := attrs["depends_on"]; ok {
		if deps, ok := dependsOn.([]interface{}); ok {
			for _, dep := range deps {
				if depStr, ok := dep.(string); ok {
					resource.DependsOn = append(resource.DependsOn, depStr)
				}
			}
		}
		delete(attrs, "depends_on")
	}

	// Parse lifecycle block
	if lifecycleData, ok := attrs["lifecycle"]; ok {
		if lifecycleMap, ok := lifecycleData.(map[string]interface{}); ok {
			resource.Lifecycle = &Lifecycle{}
			if cdb, ok := lifecycleMap["create_before_destroy"].(bool); ok {
				resource.Lifecycle.CreateBeforeDestroy = cdb
			}
			if pd, ok := lifecycleMap["prevent_destroy"].(bool); ok {
				resource.Lifecycle.PreventDestroy = pd
			}
			if ic, ok := lifecycleMap["ignore_changes"].([]interface{}); ok {
				for _, change := range ic {
					if changeStr, ok := change.(string); ok {
						resource.Lifecycle.IgnoreChanges = append(resource.Lifecycle.IgnoreChanges, changeStr)
					}
				}
			}
		}
		delete(attrs, "lifecycle")
	}

	resource.Configuration = attrs

	return resource, nil
}

// parseModuleBlock parses a module block
func (p *Parser) parseModuleBlock(block *hclsyntax.Block) (*TerraformModule, error) {
	if len(block.Labels) < 1 {
		return nil, fmt.Errorf("module block requires a name label")
	}

	module := &TerraformModule{
		Name:          block.Labels[0],
		Configuration: make(map[string]interface{}),
	}

	attrs, err := p.parseBlockContent(block.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse module content: %w", err)
	}

	// Extract source and version
	if source, ok := attrs["source"].(string); ok {
		module.Source = source
		delete(attrs, "source")
	}

	if version, ok := attrs["version"].(string); ok {
		module.Version = version
		delete(attrs, "version")
	}

	module.Configuration = attrs

	return module, nil
}

// parseVariableBlock parses a variable block
func (p *Parser) parseVariableBlock(block *hclsyntax.Block) (*TerraformVariable, error) {
	if len(block.Labels) < 1 {
		return nil, fmt.Errorf("variable block requires a name label")
	}

	variable := &TerraformVariable{
		Name:       block.Labels[0],
		Validation: make([]Validation, 0),
	}

	attrs, err := p.parseBlockContent(block.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse variable content: %w", err)
	}

	if varType, ok := attrs["type"].(string); ok {
		variable.Type = varType
	}

	if desc, ok := attrs["description"].(string); ok {
		variable.Description = desc
	}

	if defVal, ok := attrs["default"]; ok {
		variable.Default = defVal
	}

	if sensitive, ok := attrs["sensitive"].(bool); ok {
		variable.Sensitive = sensitive
	}

	return variable, nil
}

// parseOutputBlock parses an output block
func (p *Parser) parseOutputBlock(block *hclsyntax.Block) (*TerraformOutput, error) {
	if len(block.Labels) < 1 {
		return nil, fmt.Errorf("output block requires a name label")
	}

	output := &TerraformOutput{
		Name: block.Labels[0],
	}

	attrs, err := p.parseBlockContent(block.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse output content: %w", err)
	}

	if value, ok := attrs["value"]; ok {
		output.Value = value
	}

	if desc, ok := attrs["description"].(string); ok {
		output.Description = desc
	}

	if sensitive, ok := attrs["sensitive"].(bool); ok {
		output.Sensitive = sensitive
	}

	return output, nil
}

// parseProviderBlock parses a provider block
func (p *Parser) parseProviderBlock(block *hclsyntax.Block) (*TerraformProvider, error) {
	if len(block.Labels) < 1 {
		return nil, fmt.Errorf("provider block requires a name label")
	}

	provider := &TerraformProvider{
		Name:          block.Labels[0],
		Configuration: make(map[string]interface{}),
	}

	attrs, err := p.parseBlockContent(block.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse provider content: %w", err)
	}

	if alias, ok := attrs["alias"].(string); ok {
		provider.Alias = alias
		delete(attrs, "alias")
	}

	provider.Configuration = attrs

	return provider, nil
}

// parseDataBlock parses a data source block
func (p *Parser) parseDataBlock(block *hclsyntax.Block) (*TerraformData, error) {
	if len(block.Labels) < 2 {
		return nil, fmt.Errorf("data block requires 2 labels (type and name)")
	}

	dataType := block.Labels[0]
	dataName := block.Labels[1]
	provider := extractProviderFromType(dataType)

	data := &TerraformData{
		Type:          dataType,
		Name:          dataName,
		Address:       fmt.Sprintf("data.%s.%s", dataType, dataName),
		Provider:      provider,
		Configuration: make(map[string]interface{}),
	}

	attrs, err := p.parseBlockContent(block.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse data content: %w", err)
	}

	data.Configuration = attrs

	return data, nil
}

// parseTerraformBlock parses the terraform {} block
func (p *Parser) parseTerraformBlock(block *hclsyntax.Block) (*TerraformBlock, error) {
	tfBlock := &TerraformBlock{
		RequiredProviders: make(map[string]ProviderRequirement),
	}

	attrs, err := p.parseBlockContent(block.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse terraform block: %w", err)
	}

	if reqVer, ok := attrs["required_version"].(string); ok {
		tfBlock.RequiredVersion = reqVer
	}

	if reqProvs, ok := attrs["required_providers"].(map[string]interface{}); ok {
		for provName, provData := range reqProvs {
			if provMap, ok := provData.(map[string]interface{}); ok {
				req := ProviderRequirement{}
				if source, ok := provMap["source"].(string); ok {
					req.Source = source
				}
				if version, ok := provMap["version"].(string); ok {
					req.Version = version
				}
				tfBlock.RequiredProviders[provName] = req
			}
		}
	}

	return tfBlock, nil
}

// parseBlockContent parses the content of an HCL block
func (p *Parser) parseBlockContent(body *hclsyntax.Body) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Parse attributes
	for name, attr := range body.Attributes {
		value, err := p.evalExpression(attr.Expr)
		if err != nil {
			// Store the expression as string if evaluation fails
			result[name] = fmt.Sprintf("${%s}", string(attr.Expr.Range().SliceBytes(body.SrcRange.SliceBytes([]byte{}))))
			continue
		}
		result[name] = value
	}

	// Parse nested blocks
	for _, block := range body.Blocks {
		blockContent, err := p.parseBlockContent(block.Body)
		if err != nil {
			continue
		}

		if existing, ok := result[block.Type]; ok {
			// Convert to array if multiple blocks of same type
			if existingSlice, ok := existing.([]interface{}); ok {
				result[block.Type] = append(existingSlice, blockContent)
			} else {
				result[block.Type] = []interface{}{existing, blockContent}
			}
		} else {
			result[block.Type] = blockContent
		}
	}

	return result, nil
}

// evalExpression evaluates an HCL expression to a Go value
func (p *Parser) evalExpression(expr hclsyntax.Expression) (interface{}, error) {
	// Simple evaluation for basic types
	switch e := expr.(type) {
	case *hclsyntax.LiteralValueExpr:
		return ctyValueToInterface(e.Val), nil

	case *hclsyntax.TemplateExpr:
		// Handle string templates
		if len(e.Parts) == 1 {
			if lit, ok := e.Parts[0].(*hclsyntax.LiteralValueExpr); ok {
				return ctyValueToInterface(lit.Val), nil
			}
		}
		// Return the template as string
		return fmt.Sprintf("${template}"), nil

	case *hclsyntax.TupleConsExpr:
		// Handle arrays
		result := make([]interface{}, 0, len(e.Exprs))
		for _, item := range e.Exprs {
			val, err := p.evalExpression(item)
			if err == nil {
				result = append(result, val)
			}
		}
		return result, nil

	case *hclsyntax.ObjectConsExpr:
		// Handle objects/maps
		result := make(map[string]interface{})
		for _, item := range e.Items {
			keyExpr, ok := item.KeyExpr.(*hclsyntax.ObjectConsKeyExpr)
			if !ok {
				continue
			}

			key, err := p.evalExpression(keyExpr.Wrapped)
			if err != nil {
				continue
			}

			value, err := p.evalExpression(item.ValueExpr)
			if err != nil {
				continue
			}

			if keyStr, ok := key.(string); ok {
				result[keyStr] = value
			}
		}
		return result, nil

	default:
		// For complex expressions, return a placeholder
		return fmt.Sprintf("${expression}"), nil
	}
}

// ctyValueToInterface converts a cty.Value to a Go interface{}
func ctyValueToInterface(val cty.Value) interface{} {
	if val.IsNull() {
		return nil
	}

	switch val.Type() {
	case cty.String:
		return val.AsString()
	case cty.Number:
		num, _ := val.AsBigFloat().Float64()
		return num
	case cty.Bool:
		return val.True()
	default:
		if val.Type().IsListType() || val.Type().IsTupleType() {
			result := make([]interface{}, 0, val.LengthInt())
			iter := val.ElementIterator()
			for iter.Next() {
				_, v := iter.Element()
				result = append(result, ctyValueToInterface(v))
			}
			return result
		}

		if val.Type().IsMapType() || val.Type().IsObjectType() {
			result := make(map[string]interface{})
			iter := val.ElementIterator()
			for iter.Next() {
				k, v := iter.Element()
				result[k.AsString()] = ctyValueToInterface(v)
			}
			return result
		}

		return nil
	}
}

// extractProviderFromType extracts the provider name from a resource type
// e.g., "aws_instance" -> "aws", "google_compute_instance" -> "gcp"
func extractProviderFromType(resourceType string) string {
	parts := strings.SplitN(resourceType, "_", 2)
	if len(parts) == 0 {
		return "unknown"
	}

	provider := parts[0]

	// Normalize provider names
	switch provider {
	case "google":
		return "gcp"
	case "azurerm", "azuread", "azapi":
		return "azure"
	default:
		return provider
	}
}
