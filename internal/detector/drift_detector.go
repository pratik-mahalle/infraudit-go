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
	// Convert to JSON and compare (handles nested structures)
	j1, _ := json.Marshal(v1)
	j2, _ := json.Marshal(v2)
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
