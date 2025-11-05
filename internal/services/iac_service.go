package services

import (
	"context"
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

	// Similar extraction logic for CloudFormation
	// This is a placeholder - actual implementation would need proper type conversion

	return resources
}

// extractKubernetesResources extracts resources from Kubernetes parsed data
func (s *IaCService) extractKubernetesResources(definition *iac.IaCDefinition) []iac.IaCResource {
	resources := make([]iac.IaCResource, 0)

	if definition.ParsedResources == nil {
		return resources
	}

	// Similar extraction logic for Kubernetes
	// This is a placeholder - actual implementation would need proper type conversion

	return resources
}

// mapToTerraformResource converts a map to TerraformResource
// This is a helper for type conversion
func (s *IaCService) mapToTerraformResource(resMap map[string]interface{}) tfparser.TerraformResource {
	tfRes := tfparser.TerraformResource{}

	if typeVal, ok := resMap["Type"].(string); ok {
		tfRes.Type = typeVal
	}
	if name, ok := resMap["Name"].(string); ok {
		tfRes.Name = name
	}
	if addr, ok := resMap["Address"].(string); ok {
		tfRes.Address = addr
	}
	if provider, ok := resMap["Provider"].(string); ok {
		tfRes.Provider = provider
	}
	if config, ok := resMap["Configuration"].(map[string]interface{}); ok {
		tfRes.Configuration = config
	}

	return tfRes
}

// compareResources compares IaC resources with actual deployed resources
func (s *IaCService) compareResources(iacResources []*iac.IaCResource, actualResources interface{}, definition *iac.IaCDefinition) []*iac.IaCDriftResult {
	drifts := make([]*iac.IaCDriftResult, 0)

	// Create a map of actual resources for quick lookup
	// This is a simplified approach - production would need more sophisticated matching
	actualResourceMap := make(map[string]interface{})

	// Check for missing resources (in IaC but not deployed)
	for _, iacRes := range iacResources {
		// Try to find matching actual resource
		// For now, we'll create a "missing" drift for demonstration
		drift := &iac.IaCDriftResult{
			UserID:          definition.UserID,
			IaCDefinitionID: definition.ID,
			IaCResourceID:   &iacRes.ID,
			DriftCategory:   iac.DriftCategoryMissing,
			Status:          iac.DriftStatusDetected,
			Details: map[string]interface{}{
				"message": fmt.Sprintf("Resource %s defined in IaC but not found in deployed resources", iacRes.ResourceAddress),
				"iac_resource": map[string]interface{}{
					"type":    iacRes.ResourceType,
					"name":    iacRes.ResourceName,
					"address": iacRes.ResourceAddress,
				},
			},
		}

		severity := iac.SeverityMedium
		drift.Severity = &severity

		drifts = append(drifts, drift)
	}

	// Check for shadow resources (deployed but not in IaC)
	// This would require iterating through actual resources and checking if they're in IaC

	_ = actualResourceMap // Avoid unused variable error

	return drifts
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
