package detector

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/pratik-mahalle/infraudit/internal/domain/iac"
)

// IaCDriftDetector detects configuration drift between IaC and deployed resources
type IaCDriftDetector struct {
}

// NewIaCDriftDetector creates a new IaC drift detector
func NewIaCDriftDetector() *IaCDriftDetector {
	return &IaCDriftDetector{}
}

// DetectDrifts compares IaC resources with actual deployed resources
func (d *IaCDriftDetector) DetectDrifts(iacResources []*iac.IaCResource, actualResources []*ActualResource) []*iac.IaCDriftResult {
	drifts := make([]*iac.IaCDriftResult, 0)

	// Create maps for efficient lookup
	iacMap := d.createIaCResourceMap(iacResources)
	actualMap := d.createActualResourceMap(actualResources)

	// Check for missing resources (in IaC but not deployed)
	for address, iacRes := range iacMap {
		if _, exists := actualMap[address]; !exists {
			drift := d.createMissingResourceDrift(iacRes)
			drifts = append(drifts, drift)
		}
	}

	// Check for shadow resources (deployed but not in IaC)
	for address, actualRes := range actualMap {
		if _, exists := iacMap[address]; !exists {
			drift := d.createShadowResourceDrift(actualRes)
			drifts = append(drifts, drift)
		}
	}

	// Check for configuration drifts (both exist but differ)
	for address, iacRes := range iacMap {
		if actualRes, exists := actualMap[address]; exists {
			configDrifts := d.compareConfigurations(iacRes, actualRes)
			drifts = append(drifts, configDrifts...)
		}
	}

	return drifts
}

// ActualResource represents a deployed resource
type ActualResource struct {
	ID            string
	ResourceType  string
	ResourceName  string
	Address       string
	Provider      string
	Configuration map[string]interface{}
}

// createIaCResourceMap creates a map of IaC resources by address
func (d *IaCDriftDetector) createIaCResourceMap(resources []*iac.IaCResource) map[string]*iac.IaCResource {
	resMap := make(map[string]*iac.IaCResource)
	for _, res := range resources {
		resMap[res.ResourceAddress] = res
	}
	return resMap
}

// createActualResourceMap creates a map of actual resources by address
func (d *IaCDriftDetector) createActualResourceMap(resources []*ActualResource) map[string]*ActualResource {
	resMap := make(map[string]*ActualResource)
	for _, res := range resources {
		resMap[res.Address] = res
	}
	return resMap
}

// createMissingResourceDrift creates a drift result for a missing resource
func (d *IaCDriftDetector) createMissingResourceDrift(iacRes *iac.IaCResource) *iac.IaCDriftResult {
	severity := iac.SeverityHigh

	drift := &iac.IaCDriftResult{
		UserID:          iacRes.UserID,
		IaCDefinitionID: iacRes.IaCDefinitionID,
		IaCResourceID:   &iacRes.ID,
		DriftCategory:   iac.DriftCategoryMissing,
		Severity:        &severity,
		Status:          iac.DriftStatusDetected,
		Details: map[string]interface{}{
			"message": fmt.Sprintf("Resource defined in IaC but not deployed: %s", iacRes.ResourceAddress),
			"resource_type": iacRes.ResourceType,
			"resource_name": iacRes.ResourceName,
			"provider": iacRes.Provider,
			"recommendation": "Deploy this resource or remove it from IaC definition",
		},
	}

	return drift
}

// createShadowResourceDrift creates a drift result for a shadow resource
func (d *IaCDriftDetector) createShadowResourceDrift(actualRes *ActualResource) *iac.IaCDriftResult {
	severity := iac.SeverityMedium

	drift := &iac.IaCDriftResult{
		ActualResourceID: &actualRes.ID,
		DriftCategory:    iac.DriftCategoryShadow,
		Severity:         &severity,
		Status:           iac.DriftStatusDetected,
		Details: map[string]interface{}{
			"message": fmt.Sprintf("Resource deployed but not defined in IaC: %s", actualRes.Address),
			"resource_type": actualRes.ResourceType,
			"resource_name": actualRes.ResourceName,
			"provider": actualRes.Provider,
			"recommendation": "Add this resource to IaC definition or remove from infrastructure",
		},
	}

	return drift
}

// compareConfigurations compares IaC and actual resource configurations
func (d *IaCDriftDetector) compareConfigurations(iacRes *iac.IaCResource, actualRes *ActualResource) []*iac.IaCDriftResult {
	drifts := make([]*iac.IaCDriftResult, 0)

	// Deep compare configurations
	changes := d.deepCompare("", iacRes.Configuration, actualRes.Configuration)

	if len(changes) == 0 {
		// No drift - create compliant result
		severity := iac.SeverityInfo
		drift := &iac.IaCDriftResult{
			UserID:           iacRes.UserID,
			IaCDefinitionID:  iacRes.IaCDefinitionID,
			IaCResourceID:    &iacRes.ID,
			ActualResourceID: &actualRes.ID,
			DriftCategory:    iac.DriftCategoryCompliant,
			Severity:         &severity,
			Status:           iac.DriftStatusDetected,
			Details: map[string]interface{}{
				"message": "Resource configuration matches IaC definition",
			},
		}
		drifts = append(drifts, drift)
		return drifts
	}

	// Analyze changes and determine severity
	severity := d.calculateDriftSeverity(iacRes.ResourceType, changes)

	drift := &iac.IaCDriftResult{
		UserID:           iacRes.UserID,
		IaCDefinitionID:  iacRes.IaCDefinitionID,
		IaCResourceID:    &iacRes.ID,
		ActualResourceID: &actualRes.ID,
		DriftCategory:    iac.DriftCategoryModified,
		Severity:         &severity,
		Status:           iac.DriftStatusDetected,
		Details: map[string]interface{}{
			"message": fmt.Sprintf("Configuration drift detected for %s", iacRes.ResourceAddress),
			"changes": changes,
			"change_count": len(changes),
			"recommendation": d.generateRecommendation(changes),
		},
	}

	drifts = append(drifts, drift)

	return drifts
}

// deepCompare performs deep comparison of two configurations
func (d *IaCDriftDetector) deepCompare(path string, iacValue, actualValue interface{}) []iac.FieldChange {
	changes := make([]iac.FieldChange, 0)

	// If both are nil, no change
	if iacValue == nil && actualValue == nil {
		return changes
	}

	// If one is nil and other isn't
	if iacValue == nil && actualValue != nil {
		changes = append(changes, iac.FieldChange{
			Field:       path,
			IaCValue:    nil,
			ActualValue: actualValue,
			ChangeType:  "added",
		})
		return changes
	}

	if iacValue != nil && actualValue == nil {
		changes = append(changes, iac.FieldChange{
			Field:       path,
			IaCValue:    iacValue,
			ActualValue: nil,
			ChangeType:  "removed",
		})
		return changes
	}

	// Compare based on type
	switch iacVal := iacValue.(type) {
	case map[string]interface{}:
		if actualMap, ok := actualValue.(map[string]interface{}); ok {
			changes = append(changes, d.compareMap(path, iacVal, actualMap)...)
		} else {
			changes = append(changes, iac.FieldChange{
				Field:       path,
				IaCValue:    iacValue,
				ActualValue: actualValue,
				ChangeType:  "modified",
			})
		}

	case []interface{}:
		if actualSlice, ok := actualValue.([]interface{}); ok {
			changes = append(changes, d.compareSlice(path, iacVal, actualSlice)...)
		} else {
			changes = append(changes, iac.FieldChange{
				Field:       path,
				IaCValue:    iacValue,
				ActualValue: actualValue,
				ChangeType:  "modified",
			})
		}

	default:
		// Primitive types - direct comparison
		if !d.valuesEqual(iacValue, actualValue) {
			changes = append(changes, iac.FieldChange{
				Field:       path,
				IaCValue:    iacValue,
				ActualValue: actualValue,
				ChangeType:  "modified",
			})
		}
	}

	return changes
}

// compareMap compares two maps
func (d *IaCDriftDetector) compareMap(path string, iacMap, actualMap map[string]interface{}) []iac.FieldChange {
	changes := make([]iac.FieldChange, 0)

	// Check all fields in IaC config
	for key, iacValue := range iacMap {
		fieldPath := d.joinPath(path, key)

		if actualValue, exists := actualMap[key]; exists {
			changes = append(changes, d.deepCompare(fieldPath, iacValue, actualValue)...)
		} else {
			changes = append(changes, iac.FieldChange{
				Field:       fieldPath,
				IaCValue:    iacValue,
				ActualValue: nil,
				ChangeType:  "removed",
			})
		}
	}

	// Check for fields in actual config not in IaC
	for key, actualValue := range actualMap {
		if _, exists := iacMap[key]; !exists {
			// Skip computed fields that are expected to differ
			if d.isComputedField(key) {
				continue
			}

			fieldPath := d.joinPath(path, key)
			changes = append(changes, iac.FieldChange{
				Field:       fieldPath,
				IaCValue:    nil,
				ActualValue: actualValue,
				ChangeType:  "added",
			})
		}
	}

	return changes
}

// compareSlice compares two slices
func (d *IaCDriftDetector) compareSlice(path string, iacSlice, actualSlice []interface{}) []iac.FieldChange {
	changes := make([]iac.FieldChange, 0)

	// Simple length comparison for now
	if len(iacSlice) != len(actualSlice) {
		changes = append(changes, iac.FieldChange{
			Field:       path,
			IaCValue:    iacSlice,
			ActualValue: actualSlice,
			ChangeType:  "modified",
		})
		return changes
	}

	// Element-by-element comparison
	for i := 0; i < len(iacSlice); i++ {
		elementPath := fmt.Sprintf("%s[%d]", path, i)
		changes = append(changes, d.deepCompare(elementPath, iacSlice[i], actualSlice[i])...)
	}

	return changes
}

// joinPath joins path segments
func (d *IaCDriftDetector) joinPath(base, field string) string {
	if base == "" {
		return field
	}
	return fmt.Sprintf("%s.%s", base, field)
}

// valuesEqual checks if two values are equal
func (d *IaCDriftDetector) valuesEqual(v1, v2 interface{}) bool {
	// Convert to JSON for comparison to handle nested structures
	j1, err1 := json.Marshal(v1)
	j2, err2 := json.Marshal(v2)

	if err1 != nil || err2 != nil {
		// Fallback to reflect.DeepEqual
		return reflect.DeepEqual(v1, v2)
	}

	return string(j1) == string(j2)
}

// isComputedField checks if a field is computed and should be ignored
func (d *IaCDriftDetector) isComputedField(field string) bool {
	computedFields := map[string]bool{
		"id":                     true,
		"arn":                    true,
		"self_link":              true,
		"created_at":             true,
		"updated_at":             true,
		"creation_timestamp":     true,
		"uid":                    true,
		"resource_version":       true,
		"generation":             true,
		"managed_fields":         true,
		"status":                 true,
		"instance_id":            true,
		"public_ip":              true,
		"public_dns":             true,
		"private_ip":             true,
		"private_dns":            true,
	}

	return computedFields[field]
}

// calculateDriftSeverity determines the severity based on resource type and changes
func (d *IaCDriftDetector) calculateDriftSeverity(resourceType string, changes []iac.FieldChange) iac.Severity {
	// Check for security-critical changes
	for _, change := range changes {
		field := strings.ToLower(change.Field)

		// Critical security fields
		if strings.Contains(field, "encryption") ||
		   strings.Contains(field, "public_access") ||
		   strings.Contains(field, "security_group") ||
		   strings.Contains(field, "iam_policy") ||
		   strings.Contains(field, "acl") {
			return iac.SeverityCritical
		}

		// High severity fields
		if strings.Contains(field, "password") ||
		   strings.Contains(field, "secret") ||
		   strings.Contains(field, "key") ||
		   strings.Contains(field, "network") ||
		   strings.Contains(field, "firewall") {
			return iac.SeverityHigh
		}
	}

	// Default to medium for configuration changes
	if len(changes) > 5 {
		return iac.SeverityHigh
	}

	return iac.SeverityMedium
}

// generateRecommendation generates a recommendation based on changes
func (d *IaCDriftDetector) generateRecommendation(changes []iac.FieldChange) string {
	if len(changes) == 0 {
		return "No action required"
	}

	recommendations := []string{
		"Review the configuration differences",
		"Update IaC definition to match actual state, or",
		"Re-apply IaC to correct the drift",
	}

	// Check for critical fields
	for _, change := range changes {
		field := strings.ToLower(change.Field)
		if strings.Contains(field, "encryption") {
			recommendations = append(recommendations, "CRITICAL: Encryption configuration has changed")
		}
		if strings.Contains(field, "public") && strings.Contains(field, "access") {
			recommendations = append(recommendations, "CRITICAL: Public access configuration has changed")
		}
	}

	return strings.Join(recommendations, "; ")
}
