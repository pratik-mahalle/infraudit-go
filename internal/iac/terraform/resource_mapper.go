package terraform

import (
	"fmt"

	"github.com/pratik-mahalle/infraudit/internal/domain/iac"
)

// ResourceMapper maps Terraform resources to InfraAudit domain resources
type ResourceMapper struct{}

// NewResourceMapper creates a new resource mapper
func NewResourceMapper() *ResourceMapper {
	return &ResourceMapper{}
}

// MapToIaCResources converts parsed Terraform resources to IaC resources
func (m *ResourceMapper) MapToIaCResources(parsed *ParsedTerraform, definitionID string, userID string) []iac.IaCResource {
	resources := make([]iac.IaCResource, 0)

	// Map regular resources
	for _, tfRes := range parsed.Resources {
		iacRes := m.mapTerraformResource(tfRes, definitionID, userID)
		resources = append(resources, iacRes)
	}

	// Map data sources as resources (for comparison purposes)
	for _, tfData := range parsed.DataSources {
		iacRes := m.mapDataSource(tfData, definitionID, userID)
		resources = append(resources, iacRes)
	}

	return resources
}

// mapTerraformResource maps a single Terraform resource
func (m *ResourceMapper) mapTerraformResource(tfRes TerraformResource, definitionID string, userID string) iac.IaCResource {
	return iac.IaCResource{
		IaCDefinitionID: definitionID,
		UserID:          userID,
		ResourceType:    m.mapResourceType(tfRes.Type),
		ResourceName:    tfRes.Name,
		ResourceAddress: tfRes.Address,
		Provider:        tfRes.Provider,
		Configuration:   tfRes.Configuration,
	}
}

// mapDataSource maps a Terraform data source
func (m *ResourceMapper) mapDataSource(tfData TerraformData, definitionID string, userID string) iac.IaCResource {
	return iac.IaCResource{
		IaCDefinitionID: definitionID,
		UserID:          userID,
		ResourceType:    m.mapResourceType(tfData.Type),
		ResourceName:    tfData.Name,
		ResourceAddress: tfData.Address,
		Provider:        tfData.Provider,
		Configuration:   tfData.Configuration,
	}
}

// mapResourceType maps Terraform resource types to InfraAudit resource types
// This normalizes the resource type names across providers
func (m *ResourceMapper) mapResourceType(tfType string) string {
	// Map of Terraform types to InfraAudit types
	typeMap := map[string]string{
		// AWS EC2
		"aws_instance":               "ec2_instance",
		"aws_ebs_volume":             "ebs_volume",
		"aws_security_group":         "security_group",
		"aws_network_interface":      "network_interface",

		// AWS S3
		"aws_s3_bucket":              "s3_bucket",
		"aws_s3_bucket_policy":       "s3_bucket_policy",
		"aws_s3_bucket_acl":          "s3_bucket_acl",

		// AWS IAM
		"aws_iam_role":               "iam_role",
		"aws_iam_policy":             "iam_policy",
		"aws_iam_user":               "iam_user",
		"aws_iam_group":              "iam_group",

		// AWS VPC
		"aws_vpc":                    "vpc",
		"aws_subnet":                 "subnet",
		"aws_route_table":            "route_table",
		"aws_internet_gateway":       "internet_gateway",
		"aws_nat_gateway":            "nat_gateway",

		// AWS RDS
		"aws_db_instance":            "rds_instance",
		"aws_db_cluster":             "rds_cluster",

		// AWS Lambda
		"aws_lambda_function":        "lambda_function",

		// GCP Compute
		"google_compute_instance":    "gce_instance",
		"google_compute_disk":        "gce_disk",
		"google_compute_firewall":    "gce_firewall",
		"google_compute_network":     "gce_network",

		// GCP Storage
		"google_storage_bucket":      "gcs_bucket",
		"google_storage_bucket_iam":  "gcs_bucket_iam",

		// GCP IAM
		"google_service_account":     "gcp_service_account",
		"google_project_iam_binding": "gcp_iam_binding",

		// Azure Compute
		"azurerm_virtual_machine":             "azure_vm",
		"azurerm_linux_virtual_machine":       "azure_vm",
		"azurerm_windows_virtual_machine":     "azure_vm",
		"azurerm_virtual_machine_scale_set":   "azure_vmss",

		// Azure Storage
		"azurerm_storage_account":    "azure_storage_account",
		"azurerm_storage_container":  "azure_storage_container",

		// Azure Network
		"azurerm_virtual_network":    "azure_vnet",
		"azurerm_subnet":             "azure_subnet",
		"azurerm_network_security_group": "azure_nsg",

		// Kubernetes
		"kubernetes_deployment":      "k8s_deployment",
		"kubernetes_service":         "k8s_service",
		"kubernetes_pod":             "k8s_pod",
		"kubernetes_namespace":       "k8s_namespace",
		"kubernetes_config_map":      "k8s_configmap",
		"kubernetes_secret":          "k8s_secret",
		"kubernetes_ingress":         "k8s_ingress",
	}

	// Check if we have a mapping
	if mapped, ok := typeMap[tfType]; ok {
		return mapped
	}

	// Return original type if no mapping found
	return tfType
}

// GetProviderFromType extracts and normalizes the provider from resource type
func (m *ResourceMapper) GetProviderFromType(resourceType string) string {
	return extractProviderFromType(resourceType)
}

// CompareResources compares two resource configurations and returns differences
// This is a helper for drift detection
func (m *ResourceMapper) CompareResources(iacConfig, actualConfig map[string]interface{}) []iac.FieldChange {
	changes := make([]iac.FieldChange, 0)

	// Check all fields in IaC config
	for key, iacValue := range iacConfig {
		actualValue, exists := actualConfig[key]

		if !exists {
			changes = append(changes, iac.FieldChange{
				Field:       key,
				IaCValue:    iacValue,
				ActualValue: nil,
				ChangeType:  "removed",
			})
			continue
		}

		if !m.valuesEqual(iacValue, actualValue) {
			changes = append(changes, iac.FieldChange{
				Field:       key,
				IaCValue:    iacValue,
				ActualValue: actualValue,
				ChangeType:  "modified",
			})
		}
	}

	// Check for fields in actual config not in IaC
	for key, actualValue := range actualConfig {
		if _, exists := iacConfig[key]; !exists {
			changes = append(changes, iac.FieldChange{
				Field:       key,
				IaCValue:    nil,
				ActualValue: actualValue,
				ChangeType:  "added",
			})
		}
	}

	return changes
}

// valuesEqual compares two values for equality
func (m *ResourceMapper) valuesEqual(v1, v2 interface{}) bool {
	return fmt.Sprintf("%v", v1) == fmt.Sprintf("%v", v2)
}

// ExtractResourceIdentifiers extracts cloud-provider specific identifiers from Terraform config
// This helps match IaC resources with actual deployed resources
func (m *ResourceMapper) ExtractResourceIdentifiers(tfRes TerraformResource) map[string]string {
	identifiers := make(map[string]string)

	// Add the Terraform address as an identifier
	identifiers["terraform_address"] = tfRes.Address

	// Extract common identifiers based on resource type
	switch {
	case tfRes.Type == "aws_instance":
		if id, ok := tfRes.Configuration["id"].(string); ok {
			identifiers["instance_id"] = id
		}
		if tags, ok := tfRes.Configuration["tags"].(map[string]interface{}); ok {
			if name, ok := tags["Name"].(string); ok {
				identifiers["name"] = name
			}
		}

	case tfRes.Type == "aws_s3_bucket":
		if bucket, ok := tfRes.Configuration["bucket"].(string); ok {
			identifiers["bucket_name"] = bucket
		}

	case tfRes.Type == "google_compute_instance":
		if name, ok := tfRes.Configuration["name"].(string); ok {
			identifiers["name"] = name
		}
		if zone, ok := tfRes.Configuration["zone"].(string); ok {
			identifiers["zone"] = zone
		}

	case tfRes.Type == "google_storage_bucket":
		if name, ok := tfRes.Configuration["name"].(string); ok {
			identifiers["bucket_name"] = name
		}

	case tfRes.Type == "azurerm_virtual_machine" ||
	     tfRes.Type == "azurerm_linux_virtual_machine" ||
	     tfRes.Type == "azurerm_windows_virtual_machine":
		if name, ok := tfRes.Configuration["name"].(string); ok {
			identifiers["name"] = name
		}
		if rg, ok := tfRes.Configuration["resource_group_name"].(string); ok {
			identifiers["resource_group"] = rg
		}
	}

	return identifiers
}
