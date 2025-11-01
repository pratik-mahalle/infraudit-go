package detector

import (
	"strings"

	"infraudit/backend/internal/domain/drift"
	"infraudit/backend/internal/domain/resource"
)

// SecurityRule defines criteria for security-relevant configuration changes
type SecurityRule struct {
	Field       string   // Field name or pattern
	Severity    string   // Severity level
	DriftType   string   // Type of drift
	Conditions  []string // Conditions that trigger the rule
	Description string   // Human-readable description
}

// evaluateChange determines the severity and type of a configuration change
func (d *DriftDetector) evaluateChange(resourceType string, change ConfigChange) (string, string) {
	rules := d.getSecurityRules(resourceType)

	for _, rule := range rules {
		if d.matchesRule(change, rule) {
			return rule.Severity, rule.DriftType
		}
	}

	// Default for unclassified changes
	return drift.SeverityLow, drift.TypeConfigurationChange
}

// matchesRule checks if a change matches a security rule
func (d *DriftDetector) matchesRule(change ConfigChange, rule SecurityRule) bool {
	// Check if field path matches
	if !strings.Contains(strings.ToLower(change.Path), strings.ToLower(rule.Field)) {
		return false
	}

	// Check conditions
	for _, condition := range rule.Conditions {
		switch condition {
		case "encryption_disabled":
			if d.isEncryptionDisabled(change) {
				return true
			}
		case "public_access_enabled":
			if d.isPublicAccessEnabled(change) {
				return true
			}
		case "security_group_opened":
			if d.isSecurityGroupOpened(change) {
				return true
			}
		case "permission_escalation":
			if d.isPermissionEscalation(change) {
				return true
			}
		case "value_changed":
			if change.ChangeType == "modified" || change.ChangeType == "removed" {
				return true
			}
		}
	}

	return false
}

// Security condition checkers

func (d *DriftDetector) isEncryptionDisabled(change ConfigChange) bool {
	// Check if encryption was enabled and is now disabled
	if change.ChangeType == "modified" || change.ChangeType == "removed" {
		oldVal, oldOk := change.OldValue.(bool)
		newVal, newOk := change.NewValue.(bool)

		// Encryption field changed from true to false
		if oldOk && newOk && oldVal && !newVal {
			return true
		}

		// Encryption field removed
		if change.ChangeType == "removed" && oldOk && oldVal {
			return true
		}

		// String values: "enabled" -> "disabled", "AES256" -> "none", etc.
		oldStr, oldStrOk := change.OldValue.(string)
		newStr, newStrOk := change.NewValue.(string)
		if oldStrOk && newStrOk {
			oldStr = strings.ToLower(oldStr)
			newStr = strings.ToLower(newStr)
			if (oldStr == "enabled" && newStr == "disabled") ||
				(oldStr == "aes256" && newStr == "none") ||
				(strings.Contains(oldStr, "encrypt") && newStr == "none") {
				return true
			}
		}
	}
	return false
}

func (d *DriftDetector) isPublicAccessEnabled(change ConfigChange) bool {
	// Check if public access was disabled and is now enabled
	if change.ChangeType == "modified" || change.ChangeType == "added" {
		newVal, newOk := change.NewValue.(bool)
		oldVal, oldOk := change.OldValue.(bool)

		// Public access enabled
		if newOk && newVal && (!oldOk || !oldVal) {
			return true
		}

		// String values: "private" -> "public", "no-public-access" -> "public-read"
		newStr, newStrOk := change.NewValue.(string)
		if newStrOk {
			newStr = strings.ToLower(newStr)
			if strings.Contains(newStr, "public") {
				return true
			}
		}
	}
	return false
}

func (d *DriftDetector) isSecurityGroupOpened(change ConfigChange) bool {
	// Check if security group rule was added or modified to allow broad access
	if change.ChangeType == "added" || change.ChangeType == "modified" {
		// Check for 0.0.0.0/0 or ::/0
		newStr, newStrOk := change.NewValue.(string)
		if newStrOk {
			if strings.Contains(newStr, "0.0.0.0/0") || strings.Contains(newStr, "::/0") {
				return true
			}
		}

		// Check in arrays/slices
		if newSlice, ok := change.NewValue.([]interface{}); ok {
			for _, item := range newSlice {
				if strItem, ok := item.(string); ok {
					if strings.Contains(strItem, "0.0.0.0/0") || strings.Contains(strItem, "::/0") {
						return true
					}
				}
			}
		}
	}
	return false
}

func (d *DriftDetector) isPermissionEscalation(change ConfigChange) bool {
	// Check if permissions were expanded (IAM policies)
	if change.ChangeType == "added" || change.ChangeType == "modified" {
		newStr, newStrOk := change.NewValue.(string)
		if newStrOk {
			newStr = strings.ToLower(newStr)
			// Check for admin or wildcard permissions
			if strings.Contains(newStr, "*") ||
				strings.Contains(newStr, "admin") ||
				strings.Contains(newStr, "full") {
				return true
			}
		}
	}
	return false
}

// getSecurityRules returns security rules for a given resource type
func (d *DriftDetector) getSecurityRules(resourceType string) []SecurityRule {
	// Common rules for all resources
	commonRules := []SecurityRule{
		{
			Field:       "encryption",
			Severity:    drift.SeverityCritical,
			DriftType:   drift.TypeEncryption,
			Conditions:  []string{"encryption_disabled"},
			Description: "Encryption was disabled",
		},
		{
			Field:       "public",
			Severity:    drift.SeverityCritical,
			DriftType:   drift.TypeSecurityGroup,
			Conditions:  []string{"public_access_enabled"},
			Description: "Public access was enabled",
		},
	}

	// Resource-specific rules
	switch resourceType {
	case resource.TypeS3Bucket, resource.TypeGCSBucket, resource.TypeAzureStorage:
		return append(commonRules, []SecurityRule{
			{
				Field:       "versioning",
				Severity:    drift.SeverityMedium,
				DriftType:   drift.TypeConfigurationChange,
				Conditions:  []string{"value_changed"},
				Description: "Versioning configuration changed",
			},
			{
				Field:       "acl",
				Severity:    drift.SeverityHigh,
				DriftType:   drift.TypeSecurityGroup,
				Conditions:  []string{"value_changed"},
				Description: "Bucket ACL permissions changed",
			},
			{
				Field:       "policy",
				Severity:    drift.SeverityHigh,
				DriftType:   drift.TypeIAMPolicy,
				Conditions:  []string{"value_changed"},
				Description: "Bucket policy changed",
			},
		}...)

	case resource.TypeEC2Instance, resource.TypeGCEInstance, resource.TypeAzureVM:
		return append(commonRules, []SecurityRule{
			{
				Field:       "security",
				Severity:    drift.SeverityCritical,
				DriftType:   drift.TypeSecurityGroup,
				Conditions:  []string{"security_group_opened"},
				Description: "Security group allows unrestricted access",
			},
			{
				Field:       "iam",
				Severity:    drift.SeverityHigh,
				DriftType:   drift.TypeIAMPolicy,
				Conditions:  []string{"permission_escalation"},
				Description: "IAM role permissions expanded",
			},
			{
				Field:       "ssh",
				Severity:    drift.SeverityMedium,
				DriftType:   drift.TypeSecurityGroup,
				Conditions:  []string{"value_changed"},
				Description: "SSH configuration changed",
			},
			{
				Field:       "network",
				Severity:    drift.SeverityMedium,
				DriftType:   drift.TypeNetworkRule,
				Conditions:  []string{"value_changed"},
				Description: "Network configuration changed",
			},
		}...)

	default:
		return commonRules
	}
}
