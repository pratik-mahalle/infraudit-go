package detector

import (
	"encoding/json"
	"fmt"

	"github.com/pratik-mahalle/infraudit/internal/domain/drift"
)

// DriftDetector analyzes configuration changes and identifies security drifts
type DriftDetector struct{}

// NewDriftDetector creates a new drift detector
func NewDriftDetector() *DriftDetector {
	return &DriftDetector{}
}

// ConfigChange represents a change in configuration
type ConfigChange struct {
	Field     string      `json:"field"`
	OldValue  interface{} `json:"old_value"`
	NewValue  interface{} `json:"new_value"`
	Path      string      `json:"path"` // JSON path to the field
	ChangeType string     `json:"change_type"` // added, removed, modified
}

// DetectionResult contains the result of drift detection
type DetectionResult struct {
	HasDrift  bool           `json:"has_drift"`
	DriftType string         `json:"drift_type"`
	Severity  string         `json:"severity"`
	Details   string         `json:"details"`
	Changes   []ConfigChange `json:"changes"`
}

// DetectDrift compares baseline and current configurations
func (d *DriftDetector) DetectDrift(resourceType string, baselineConfig, currentConfig string) (*DetectionResult, error) {
	// Parse JSON configurations
	var baseline, current map[string]interface{}

	if err := json.Unmarshal([]byte(baselineConfig), &baseline); err != nil {
		return nil, fmt.Errorf("failed to parse baseline config: %w", err)
	}

	if err := json.Unmarshal([]byte(currentConfig), &current); err != nil {
		return nil, fmt.Errorf("failed to parse current config: %w", err)
	}

	// Compare configurations
	changes := d.compareConfigs(baseline, current, "")

	if len(changes) == 0 {
		return &DetectionResult{
			HasDrift: false,
		}, nil
	}

	// Analyze changes for security relevance
	return d.analyzeChanges(resourceType, changes), nil
}

// compareConfigs recursively compares two configuration maps
func (d *DriftDetector) compareConfigs(baseline, current map[string]interface{}, path string) []ConfigChange {
	var changes []ConfigChange

	// Check for modified and removed fields
	for key, oldVal := range baseline {
		currentPath := key
		if path != "" {
			currentPath = path + "." + key
		}

		newVal, exists := current[key]
		if !exists {
			changes = append(changes, ConfigChange{
				Field:      key,
				OldValue:   oldVal,
				NewValue:   nil,
				Path:       currentPath,
				ChangeType: "removed",
			})
			continue
		}

		// If both are maps, recurse
		oldMap, oldIsMap := oldVal.(map[string]interface{})
		newMap, newIsMap := newVal.(map[string]interface{})
		if oldIsMap && newIsMap {
			changes = append(changes, d.compareConfigs(oldMap, newMap, currentPath)...)
			continue
		}

		// Compare values
		if !d.valuesEqual(oldVal, newVal) {
			changes = append(changes, ConfigChange{
				Field:      key,
				OldValue:   oldVal,
				NewValue:   newVal,
				Path:       currentPath,
				ChangeType: "modified",
			})
		}
	}

	// Check for added fields
	for key, newVal := range current {
		currentPath := key
		if path != "" {
			currentPath = path + "." + key
		}

		if _, exists := baseline[key]; !exists {
			changes = append(changes, ConfigChange{
				Field:      key,
				OldValue:   nil,
				NewValue:   newVal,
				Path:       currentPath,
				ChangeType: "added",
			})
		}
	}

	return changes
}

// valuesEqual compares two values for equality
func (d *DriftDetector) valuesEqual(v1, v2 interface{}) bool {
	// Fast path for nil comparison
	if v1 == nil && v2 == nil {
		return true
	}
	if v1 == nil || v2 == nil {
		return false
	}

	// Fast path for primitive types
	switch v1Type := v1.(type) {
	case string:
		v2Str, ok := v2.(string)
		return ok && v1Type == v2Str
	case float64:
		v2Float, ok := v2.(float64)
		return ok && v1Type == v2Float
	case bool:
		v2Bool, ok := v2.(bool)
		return ok && v1Type == v2Bool
	case int:
		v2Int, ok := v2.(int)
		return ok && v1Type == v2Int
	}

	// For slices, compare elements
	if v1Slice, ok := v1.([]interface{}); ok {
		v2Slice, ok := v2.([]interface{})
		if !ok || len(v1Slice) != len(v2Slice) {
			return false
		}
		for i := range v1Slice {
			if !d.valuesEqual(v1Slice[i], v2Slice[i]) {
				return false
			}
		}
		return true
	}

	// For maps, compare recursively
	if v1Map, ok := v1.(map[string]interface{}); ok {
		v2Map, ok := v2.(map[string]interface{})
		if !ok || len(v1Map) != len(v2Map) {
			return false
		}
		for key, val1 := range v1Map {
			val2, exists := v2Map[key]
			if !exists || !d.valuesEqual(val1, val2) {
				return false
			}
		}
		return true
	}

	// Fallback to JSON comparison for complex types
	j1, err1 := json.Marshal(v1)
	j2, err2 := json.Marshal(v2)
	if err1 != nil || err2 != nil {
		return false
	}
	return string(j1) == string(j2)
}

// analyzeChanges determines drift type and severity based on changes
func (d *DriftDetector) analyzeChanges(resourceType string, changes []ConfigChange) *DetectionResult {
	result := &DetectionResult{
		HasDrift: true,
		Changes:  changes,
		Severity: drift.SeverityLow, // Default
	}

	// Analyze each change for security impact
	highestSeverity := drift.SeverityLow
	driftTypes := make(map[string]bool)

	for _, change := range changes {
		severity, driftType := d.evaluateChange(resourceType, change)
		driftTypes[driftType] = true

		if d.severityLevel(severity) > d.severityLevel(highestSeverity) {
			highestSeverity = severity
		}
	}

	result.Severity = highestSeverity

	// Determine primary drift type
	if driftTypes[drift.TypeEncryption] {
		result.DriftType = drift.TypeEncryption
	} else if driftTypes[drift.TypeSecurityGroup] {
		result.DriftType = drift.TypeSecurityGroup
	} else if driftTypes[drift.TypeIAMPolicy] {
		result.DriftType = drift.TypeIAMPolicy
	} else if driftTypes[drift.TypeNetworkRule] {
		result.DriftType = drift.TypeNetworkRule
	} else {
		result.DriftType = drift.TypeConfigurationChange
	}

	result.Details = d.generateDetails(changes, result.Severity)

	return result
}

// severityLevel converts severity string to numeric level
func (d *DriftDetector) severityLevel(severity string) int {
	switch severity {
	case drift.SeverityCritical:
		return 4
	case drift.SeverityHigh:
		return 3
	case drift.SeverityMedium:
		return 2
	case drift.SeverityLow:
		return 1
	default:
		return 0
	}
}

// generateDetails creates a human-readable description of changes
func (d *DriftDetector) generateDetails(changes []ConfigChange, severity string) string {
	if len(changes) == 0 {
		return "No changes detected"
	}

	detail := fmt.Sprintf("%d configuration change(s) detected with %s severity:\n", len(changes), severity)
	for i, change := range changes {
		if i >= 5 { // Limit to first 5 changes
			detail += fmt.Sprintf("... and %d more changes", len(changes)-5)
			break
		}
		detail += fmt.Sprintf("- %s: %s (was: %v, now: %v)\n",
			change.Path, change.ChangeType, change.OldValue, change.NewValue)
	}

	return detail
}
