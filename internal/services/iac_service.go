package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/pratik-mahalle/infraudit/internal/domain/iac"
	"github.com/pratik-mahalle/infraudit/internal/domain/resource"
	cfparser "github.com/pratik-mahalle/infraudit/internal/iac/cloudformation"
	k8sparser "github.com/pratik-mahalle/infraudit/internal/iac/kubernetes"
	tfparser "github.com/pratik-mahalle/infraudit/internal/iac/terraform"
	"github.com/pratik-mahalle/infraudit/internal/repository/postgres"
)

// IaCService handles IaC-related business logic
type IaCService struct {
	repo            *postgres.IaCRepository
	resourceService *ResourceService
	driftService    *DriftService
}

// NewIaCService creates a new IaC service
func NewIaCService(repo *postgres.IaCRepository, resourceService *ResourceService, driftService *DriftService) *IaCService {
	return &IaCService{
		repo:            repo,
		resourceService: resourceService,
		driftService:    driftService,
	}
}

// UploadAndParse uploads an IaC file and parses it
func (s *IaCService) UploadAndParse(ctx context.Context, userID, name string, iacType iac.IaCType, content string) (*iac.IaCDefinition, error) {
	// Validate IaC type
	if err := s.validateIaCType(iacType); err != nil {
		return nil, err
	}

	// Create definition
	definition := &iac.IaCDefinition{
		UserID:  userID,
		Name:    name,
		IaCType: iacType,
		Content: content,
	}

	if err := definition.Validate(); err != nil {
		return nil, err
	}

	// Parse the IaC content
	parsedResources, err := s.parseIaC(iacType, []byte(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse IaC: %w", err)
	}

	definition.ParsedResources = parsedResources
	now := time.Now()
	definition.LastParsed = &now

	// Save to database
	if err := s.repo.CreateDefinition(ctx, definition); err != nil {
		return nil, err
	}

	// Extract and save individual resources
	if err := s.saveIaCResources(ctx, definition); err != nil {
		return nil, fmt.Errorf("failed to save IaC resources: %w", err)
	}

	return definition, nil
}

// GetDefinition retrieves an IaC definition by ID
func (s *IaCService) GetDefinition(ctx context.Context, userID, definitionID string) (*iac.IaCDefinition, error) {
	return s.repo.GetDefinitionByID(ctx, userID, definitionID)
}

// ListDefinitions lists all IaC definitions for a user
func (s *IaCService) ListDefinitions(ctx context.Context, userID string, iacType *iac.IaCType) ([]*iac.IaCDefinition, error) {
	return s.repo.ListDefinitions(ctx, userID, iacType)
}

// DeleteDefinition deletes an IaC definition
func (s *IaCService) DeleteDefinition(ctx context.Context, userID, definitionID string) error {
	// Delete associated resources first
	if err := s.repo.DeleteResourcesByDefinition(ctx, userID, definitionID); err != nil {
		return err
	}

	return s.repo.DeleteDefinition(ctx, userID, definitionID)
}

// DetectDrift compares IaC definition with actual deployed resources
func (s *IaCService) DetectDrift(ctx context.Context, userID, definitionID string) ([]*iac.IaCDriftResult, error) {
	// Get IaC definition
	definition, err := s.repo.GetDefinitionByID(ctx, userID, definitionID)
	if err != nil {
		return nil, err
	}

	// Get IaC resources
	iacResources, err := s.repo.ListResourcesByDefinition(ctx, userID, definitionID)
	if err != nil {
		return nil, err
	}

	if len(iacResources) == 0 {
		return nil, fmt.Errorf("no resources found in IaC definition")
	}

	// Get actual deployed resources
	// Note: This is a simplified approach - in production, you'd filter by provider and resource types
	// Convert userID from string to int64
	userIDInt, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	actualResources, _, err := s.resourceService.List(ctx, userIDInt, resource.Filter{}, 1000, 0)
	if err != nil {
		return nil, err
	}

	// Perform drift detection
	drifts := s.compareResources(iacResources, actualResources, definition)

	// Save drift results
	for _, drift := range drifts {
		if err := s.repo.CreateDriftResult(ctx, drift); err != nil {
			return nil, fmt.Errorf("failed to save drift result: %w", err)
		}
	}

	return drifts, nil
}

// GetDriftResults retrieves drift results
func (s *IaCService) GetDriftResults(ctx context.Context, userID string, definitionID *string, category *iac.DriftCategory, status *iac.DriftStatus) ([]*iac.IaCDriftResult, error) {
	return s.repo.ListDriftResults(ctx, userID, definitionID, category, status)
}

// UpdateDriftStatus updates the status of a drift
func (s *IaCService) UpdateDriftStatus(ctx context.Context, userID, driftID string, status iac.DriftStatus) error {
	return s.repo.UpdateDriftStatus(ctx, userID, driftID, status)
}

// GetDriftSummary returns a summary of drifts
func (s *IaCService) GetDriftSummary(ctx context.Context, userID string, definitionID *string) (map[string]interface{}, error) {
	return s.repo.GetDriftSummary(ctx, userID, definitionID)
}

// parseIaC parses IaC content based on type
func (s *IaCService) parseIaC(iacType iac.IaCType, content []byte) (map[string]interface{}, error) {
	switch iacType {
	case iac.IaCTypeTerraform:
		return s.parseTerraform(content)

	case iac.IaCTypeCloudFormation:
		return s.parseCloudFormation(content)

	case iac.IaCTypeKubernetes:
		return s.parseKubernetes(content)

	case iac.IaCTypeHelm:
		// Helm charts are treated as Kubernetes manifests after templating
		return s.parseKubernetes(content)

	default:
		return nil, iac.ErrUnsupportedFormat
	}
}

// parseTerraform parses Terraform content
func (s *IaCService) parseTerraform(content []byte) (map[string]interface{}, error) {
	parser := tfparser.NewParser()
	result, err := parser.Parse(content, "main.tf")
	if err != nil {
		return nil, err
	}

	if result.HasErrors() {
		return nil, fmt.Errorf("terraform parsing errors: %v", result.ErrorMessages())
	}

	// Convert to map for storage
	return map[string]interface{}{
		"resources":    result.Parsed.Resources,
		"modules":      result.Parsed.Modules,
		"variables":    result.Parsed.Variables,
		"outputs":      result.Parsed.Outputs,
		"providers":    result.Parsed.Providers,
		"data_sources": result.Parsed.DataSources,
	}, nil
}

// parseCloudFormation parses CloudFormation content
func (s *IaCService) parseCloudFormation(content []byte) (map[string]interface{}, error) {
	parser := cfparser.NewParser()

	// Try JSON first, then YAML
	result, err := parser.ParseJSON(content)
	if err != nil {
		result, err = parser.ParseYAML(content)
		if err != nil {
			return nil, err
		}
	}

	if result.HasErrors() {
		return nil, fmt.Errorf("cloudformation parsing errors: %v", result.ErrorMessages())
	}

	// Convert to map for storage
	return map[string]interface{}{
		"resources":  result.Parsed.Resources,
		"parameters": result.Parsed.Parameters,
		"outputs":    result.Parsed.Outputs,
	}, nil
}

// parseKubernetes parses Kubernetes manifest content
func (s *IaCService) parseKubernetes(content []byte) (map[string]interface{}, error) {
	parser := k8sparser.NewParser()
	result, err := parser.Parse(content, "manifest.yaml")
	if err != nil {
		return nil, err
	}

	if result.HasErrors() {
		return nil, fmt.Errorf("kubernetes parsing errors: %v", result.ErrorMessages())
	}

	// Convert to map for storage
	return map[string]interface{}{
		"resources": result.Parsed.Resources,
		"namespace": result.Parsed.Namespace,
	}, nil
}

// saveIaCResources extracts and saves individual resources from parsed IaC
func (s *IaCService) saveIaCResources(ctx context.Context, definition *iac.IaCDefinition) error {
	var resources []iac.IaCResource

	switch definition.IaCType {
	case iac.IaCTypeTerraform:
		resources = s.extractTerraformResources(definition)

	case iac.IaCTypeCloudFormation:
		resources = s.extractCloudFormationResources(definition)

	case iac.IaCTypeKubernetes, iac.IaCTypeHelm:
		resources = s.extractKubernetesResources(definition)
	}

	// Save each resource
	for _, res := range resources {
		if err := s.repo.CreateResource(ctx, &res); err != nil {
			return err
		}
	}

	return nil
}

// extractTerraformResources extracts resources from Terraform parsed data
func (s *IaCService) extractTerraformResources(definition *iac.IaCDefinition) []iac.IaCResource {
	resources := make([]iac.IaCResource, 0)

	if definition.ParsedResources == nil {
		return resources
	}

	// Extract resources array
	if tfResources, ok := definition.ParsedResources["resources"].([]interface{}); ok {
		mapper := tfparser.NewResourceMapper()

		for _, res := range tfResources {
			// Convert back to TerraformResource
			// This is a simplified approach - in production, you'd need proper type conversion
			if resMap, ok := res.(map[string]interface{}); ok {
				tfRes := s.mapToTerraformResource(resMap)
				iacRes := mapper.MapToIaCResources(
					&tfparser.ParsedTerraform{Resources: []tfparser.TerraformResource{tfRes}},
					definition.ID,
					definition.UserID,
				)
				resources = append(resources, iacRes...)
			}
		}
	}

	return resources
}

// extractCloudFormationResources extracts resources from CloudFormation parsed data
func (s *IaCService) extractCloudFormationResources(definition *iac.IaCDefinition) []iac.IaCResource {
	resources := make([]iac.IaCResource, 0)

	if definition.ParsedResources == nil {
		return resources
	}

	// Extract resources array from parsed data
	cfResourcesInterface, ok := definition.ParsedResources["resources"].([]interface{})
	if !ok {
		return resources
	}

	mapper := cfparser.NewResourceMapper()
	cfResources := make([]cfparser.CloudFormationResource, 0, len(cfResourcesInterface))

	// Reconstruct CloudFormationResource structs from maps
	for _, resInterface := range cfResourcesInterface {
		resMap, ok := resInterface.(map[string]interface{})
		if !ok {
			continue
		}

		cfRes := s.mapToCloudFormationResource(resMap)
		cfResources = append(cfResources, cfRes)
	}

	// Use resource mapper to convert to IaC resources
	if len(cfResources) > 0 {
		iacResources := mapper.MapToIaCResources(
			&cfparser.ParsedCloudFormation{Resources: cfResources},
			definition.ID,
			definition.UserID,
		)
		resources = append(resources, iacResources...)
	}

	return resources
}

// extractKubernetesResources extracts resources from Kubernetes parsed data
func (s *IaCService) extractKubernetesResources(definition *iac.IaCDefinition) []iac.IaCResource {
	resources := make([]iac.IaCResource, 0)

	if definition.ParsedResources == nil {
		return resources
	}

	// Extract resources array from parsed data
	k8sResourcesInterface, ok := definition.ParsedResources["resources"].([]interface{})
	if !ok {
		return resources
	}

	mapper := k8sparser.NewResourceMapper()
	k8sResources := make([]k8sparser.KubernetesResource, 0, len(k8sResourcesInterface))

	// Reconstruct KubernetesResource structs from maps
	for _, resInterface := range k8sResourcesInterface {
		resMap, ok := resInterface.(map[string]interface{})
		if !ok {
			continue
		}

		k8sRes := s.mapToKubernetesResource(resMap)
		k8sResources = append(k8sResources, k8sRes)
	}

	// Use resource mapper to convert to IaC resources
	if len(k8sResources) > 0 {
		iacResources := mapper.MapToIaCResources(
			&k8sparser.ParsedKubernetes{Resources: k8sResources},
			definition.ID,
			definition.UserID,
		)
		resources = append(resources, iacResources...)
	}

	return resources
}

// mapToTerraformResource converts a map to TerraformResource
// This is a helper for type conversion
func (s *IaCService) mapToTerraformResource(resMap map[string]interface{}) tfparser.TerraformResource {
	tfRes := tfparser.TerraformResource{}

	if typeVal, ok := resMap["type"].(string); ok {
		tfRes.Type = typeVal
	}
	if name, ok := resMap["name"].(string); ok {
		tfRes.Name = name
	}
	if addr, ok := resMap["address"].(string); ok {
		tfRes.Address = addr
	}
	if provider, ok := resMap["provider"].(string); ok {
		tfRes.Provider = provider
	}
	if config, ok := resMap["configuration"].(map[string]interface{}); ok {
		tfRes.Configuration = config
	}

	return tfRes
}

// mapToCloudFormationResource converts a map to CloudFormationResource
// This is a helper for type conversion
func (s *IaCService) mapToCloudFormationResource(resMap map[string]interface{}) cfparser.CloudFormationResource {
	cfRes := cfparser.CloudFormationResource{}

	if logicalID, ok := resMap["logical_id"].(string); ok {
		cfRes.LogicalID = logicalID
	}
	if typeVal, ok := resMap["type"].(string); ok {
		cfRes.Type = typeVal
	}
	if provider, ok := resMap["provider"].(string); ok {
		cfRes.Provider = provider
	}
	if properties, ok := resMap["properties"].(map[string]interface{}); ok {
		cfRes.Properties = properties
	}
	if dependsOn, ok := resMap["depends_on"].([]interface{}); ok {
		cfRes.DependsOn = make([]string, 0, len(dependsOn))
		for _, dep := range dependsOn {
			if depStr, ok := dep.(string); ok {
				cfRes.DependsOn = append(cfRes.DependsOn, depStr)
			}
		}
	}
	if condition, ok := resMap["condition"].(string); ok {
		cfRes.Condition = condition
	}
	if deletionPolicy, ok := resMap["deletion_policy"].(string); ok {
		cfRes.DeletionPolicy = deletionPolicy
	}

	return cfRes
}

// mapToKubernetesResource converts a map to KubernetesResource
// This is a helper for type conversion
func (s *IaCService) mapToKubernetesResource(resMap map[string]interface{}) k8sparser.KubernetesResource {
	k8sRes := k8sparser.KubernetesResource{}

	if apiVersion, ok := resMap["api_version"].(string); ok {
		k8sRes.APIVersion = apiVersion
	}
	if kind, ok := resMap["kind"].(string); ok {
		k8sRes.Kind = kind
	}
	if name, ok := resMap["name"].(string); ok {
		k8sRes.Name = name
	}
	if namespace, ok := resMap["namespace"].(string); ok {
		k8sRes.Namespace = namespace
	}
	if labels, ok := resMap["labels"].(map[string]interface{}); ok {
		k8sRes.Labels = make(map[string]string)
		for key, val := range labels {
			if strVal, ok := val.(string); ok {
				k8sRes.Labels[key] = strVal
			}
		}
	}
	if annotations, ok := resMap["annotations"].(map[string]interface{}); ok {
		k8sRes.Annotations = make(map[string]string)
		for key, val := range annotations {
			if strVal, ok := val.(string); ok {
				k8sRes.Annotations[key] = strVal
			}
		}
	}
	if spec, ok := resMap["spec"].(map[string]interface{}); ok {
		k8sRes.Spec = spec
	}
	if data, ok := resMap["data"].(map[string]interface{}); ok {
		k8sRes.Data = data
	}

	return k8sRes
}

// compareResources compares IaC resources with actual deployed resources
func (s *IaCService) compareResources(iacResources []*iac.IaCResource, actualResources interface{}, definition *iac.IaCDefinition) []*iac.IaCDriftResult {
	drifts := make([]*iac.IaCDriftResult, 0)

	// Type assert actualResources to []*resource.Resource
	actualResourceList, ok := actualResources.([]*resource.Resource)
	if !ok {
		// If type assertion fails, log and return empty drifts
		return drifts
	}

	// Build lookup map: key = "provider:type:name" -> actual resource
	// This allows O(1) lookups when matching IaC resources
	actualResourceMap := make(map[string]*resource.Resource)
	for _, actualRes := range actualResourceList {
		// Create multiple lookup keys for better matching chances
		// Key format: "provider:type:name"
		key := fmt.Sprintf("%s:%s:%s", actualRes.Provider, actualRes.Type, actualRes.Name)
		actualResourceMap[key] = actualRes

		// Also create a key with ResourceID for exact ID matching
		if actualRes.ResourceID != "" {
			idKey := fmt.Sprintf("%s:%s:%s", actualRes.Provider, actualRes.Type, actualRes.ResourceID)
			actualResourceMap[idKey] = actualRes
		}
	}

	// Track which actual resources have been matched to detect shadow resources later
	matchedActualResourceIDs := make(map[string]bool)

	// Check for missing and modified resources (in IaC)
	for _, iacRes := range iacResources {
		// Try multiple matching strategies
		var matchedResource *resource.Resource

		// Strategy 1: Match by provider:type:name
		key1 := fmt.Sprintf("%s:%s:%s", iacRes.Provider, iacRes.ResourceType, iacRes.ResourceName)
		if res, exists := actualResourceMap[key1]; exists {
			matchedResource = res
		}

		// Strategy 2: Try matching with resource address (last part after dot)
		if matchedResource == nil && iacRes.ResourceAddress != "" {
			// Extract name from address (e.g., "module.vpc.aws_instance.web" -> "web")
			parts := splitResourceAddress(iacRes.ResourceAddress)
			if len(parts) > 0 {
				addressName := parts[len(parts)-1]
				key2 := fmt.Sprintf("%s:%s:%s", iacRes.Provider, iacRes.ResourceType, addressName)
				if res, exists := actualResourceMap[key2]; exists {
					matchedResource = res
				}
			}
		}

		if matchedResource == nil {
			// Missing resource: defined in IaC but not deployed
			drift := &iac.IaCDriftResult{
				UserID:          definition.UserID,
				IaCDefinitionID: definition.ID,
				IaCResourceID:   &iacRes.ID,
				DriftCategory:   iac.DriftCategoryMissing,
				Status:          iac.DriftStatusDetected,
				Details: map[string]interface{}{
					"message": fmt.Sprintf("Resource %s defined in IaC but not found in deployed resources", iacRes.ResourceAddress),
					"iac_resource": map[string]interface{}{
						"type":          iacRes.ResourceType,
						"name":          iacRes.ResourceName,
						"address":       iacRes.ResourceAddress,
						"provider":      iacRes.Provider,
						"configuration": iacRes.Configuration,
					},
					"recommendation": "Deploy this resource or remove it from IaC definition",
				},
			}

			severity := iac.SeverityMedium
			drift.Severity = &severity
			drifts = append(drifts, drift)
		} else {
			// Mark this actual resource as matched
			matchedActualResourceIDs[matchedResource.ID] = true

			// Check for configuration drift (modified)
			configDrift := s.detectConfigurationDrift(iacRes, matchedResource)
			if configDrift != nil {
				drift := &iac.IaCDriftResult{
					UserID:           definition.UserID,
					IaCDefinitionID:  definition.ID,
					IaCResourceID:    &iacRes.ID,
					ActualResourceID: &matchedResource.ID,
					DriftCategory:    iac.DriftCategoryModified,
					Status:           iac.DriftStatusDetected,
					Details:          configDrift,
				}

				severity := iac.SeverityMedium
				drift.Severity = &severity
				drifts = append(drifts, drift)
			}
		}
	}

	// Check for shadow resources (deployed but not in IaC)
	seenResourceIDs := make(map[string]bool)
	for _, actualRes := range actualResourceMap {
		// Skip if we've already processed this resource or if it was matched
		if seenResourceIDs[actualRes.ID] || matchedActualResourceIDs[actualRes.ID] {
			continue
		}
		seenResourceIDs[actualRes.ID] = true

		drift := &iac.IaCDriftResult{
			UserID:           definition.UserID,
			IaCDefinitionID:  definition.ID,
			ActualResourceID: &actualRes.ID,
			DriftCategory:    iac.DriftCategoryShadow,
			Status:           iac.DriftStatusDetected,
			Details: map[string]interface{}{
				"message": fmt.Sprintf("Resource %s is deployed but not defined in IaC", actualRes.Name),
				"actual_resource": map[string]interface{}{
					"type":        actualRes.Type,
					"name":        actualRes.Name,
					"provider":    actualRes.Provider,
					"resource_id": actualRes.ResourceID,
					"region":      actualRes.Region,
					"status":      actualRes.Status,
				},
				"recommendation": "Add this resource to IaC definition or investigate if it should be removed",
			},
		}

		severity := iac.SeverityLow
		drift.Severity = &severity
		drifts = append(drifts, drift)
	}

	return drifts
}

// splitResourceAddress splits a resource address by dots and returns parts
// Example: "module.vpc.aws_instance.web" -> ["module", "vpc", "aws_instance", "web"]
func splitResourceAddress(address string) []string {
	if address == "" {
		return []string{}
	}
	parts := make([]string, 0)
	current := ""
	for _, char := range address {
		if char == '.' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

// detectConfigurationDrift compares IaC resource configuration with actual resource configuration
// Returns nil if no drift detected, otherwise returns details map
func (s *IaCService) detectConfigurationDrift(iacRes *iac.IaCResource, actualRes *resource.Resource) map[string]interface{} {
	changes := make([]map[string]interface{}, 0)

	// Parse actual resource configuration from JSON string
	var actualConfig map[string]interface{}
	if actualRes.Configuration != "" {
		if err := json.Unmarshal([]byte(actualRes.Configuration), &actualConfig); err != nil {
			// If parsing fails, we can't compare configurations
			return nil
		}
	}

	// Compare configurations field by field
	// Check fields in IaC that differ from actual
	for key, iacValue := range iacRes.Configuration {
		actualValue, exists := actualConfig[key]
		if !exists {
			changes = append(changes, map[string]interface{}{
				"field":       key,
				"iac_value":   iacValue,
				"actual_value": nil,
				"change_type": "missing_in_actual",
			})
		} else if !deepEqual(iacValue, actualValue) {
			changes = append(changes, map[string]interface{}{
				"field":       key,
				"iac_value":   iacValue,
				"actual_value": actualValue,
				"change_type": "modified",
			})
		}
	}

	// Check fields in actual that are not in IaC
	for key, actualValue := range actualConfig {
		if _, exists := iacRes.Configuration[key]; !exists {
			changes = append(changes, map[string]interface{}{
				"field":       key,
				"iac_value":   nil,
				"actual_value": actualValue,
				"change_type": "added_in_actual",
			})
		}
	}

	// If no changes detected, return nil
	if len(changes) == 0 {
		return nil
	}

	// Build drift details
	return map[string]interface{}{
		"message": fmt.Sprintf("Configuration drift detected for resource %s", iacRes.ResourceAddress),
		"iac_resource": map[string]interface{}{
			"type":          iacRes.ResourceType,
			"name":          iacRes.ResourceName,
			"address":       iacRes.ResourceAddress,
			"configuration": iacRes.Configuration,
		},
		"actual_resource": map[string]interface{}{
			"type":          actualRes.Type,
			"name":          actualRes.Name,
			"resource_id":   actualRes.ResourceID,
			"configuration": actualConfig,
		},
		"changes":        changes,
		"change_count":   len(changes),
		"recommendation": "Update IaC definition to match actual state or apply IaC to fix drift",
	}
}

// deepEqual performs a deep equality check between two values
// This is a simplified version - production would use reflect.DeepEqual or similar
func deepEqual(a, b interface{}) bool {
	// Convert both to JSON strings for comparison
	// This handles nested structures but may not be perfect for all cases
	aJSON, aErr := json.Marshal(a)
	bJSON, bErr := json.Marshal(b)

	if aErr != nil || bErr != nil {
		return false
	}

	return string(aJSON) == string(bJSON)
}

// validateIaCType validates the IaC type
func (s *IaCService) validateIaCType(iacType iac.IaCType) error {
	validTypes := []iac.IaCType{
		iac.IaCTypeTerraform,
		iac.IaCTypeCloudFormation,
		iac.IaCTypeKubernetes,
		iac.IaCTypeHelm,
	}

	for _, t := range validTypes {
		if iacType == t {
			return nil
		}
	}

	return iac.ErrInvalidIaCType
}
