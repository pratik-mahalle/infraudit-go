package kubernetes

import (
	"fmt"
	"reflect"

	"github.com/pratik-mahalle/infraudit/internal/domain/iac"
)

// ResourceMapper maps Kubernetes resources to InfraAudit domain resources
type ResourceMapper struct{}

// NewResourceMapper creates a new resource mapper
func NewResourceMapper() *ResourceMapper {
	return &ResourceMapper{}
}

// MapToIaCResources converts parsed Kubernetes resources to IaC resources
func (m *ResourceMapper) MapToIaCResources(parsed *ParsedKubernetes, definitionID string, userID string) []iac.IaCResource {
	resources := make([]iac.IaCResource, 0, len(parsed.Resources))

	for _, k8sRes := range parsed.Resources {
		iacRes := m.mapKubernetesResource(k8sRes, definitionID, userID)
		resources = append(resources, iacRes)
	}

	return resources
}

// mapKubernetesResource maps a single Kubernetes resource
func (m *ResourceMapper) mapKubernetesResource(k8sRes KubernetesResource, definitionID string, userID string) iac.IaCResource {
	// Build resource address
	address := m.buildResourceAddress(k8sRes)

	// Combine spec and data into configuration
	config := make(map[string]interface{})
	if k8sRes.Spec != nil {
		config["spec"] = k8sRes.Spec
	}
	if k8sRes.Data != nil {
		config["data"] = k8sRes.Data
	}
	if k8sRes.Labels != nil {
		config["labels"] = k8sRes.Labels
	}
	if k8sRes.Annotations != nil {
		config["annotations"] = k8sRes.Annotations
	}

	return iac.IaCResource{
		IaCDefinitionID: definitionID,
		UserID:          userID,
		ResourceType:    m.mapResourceType(k8sRes.Kind),
		ResourceName:    k8sRes.Name,
		ResourceAddress: address,
		Provider:        "kubernetes",
		Configuration:   config,
	}
}

// buildResourceAddress builds a unique address for a Kubernetes resource
func (m *ResourceMapper) buildResourceAddress(k8sRes KubernetesResource) string {
	if k8sRes.Namespace != "" {
		return fmt.Sprintf("%s/%s/%s", k8sRes.Kind, k8sRes.Namespace, k8sRes.Name)
	}
	return fmt.Sprintf("%s/%s", k8sRes.Kind, k8sRes.Name)
}

// mapResourceType maps Kubernetes kinds to InfraAudit resource types
func (m *ResourceMapper) mapResourceType(kind string) string {
	// Map of Kubernetes kinds to InfraAudit types
	typeMap := map[string]string{
		// Workloads
		"Pod":                   "k8s_pod",
		"Deployment":            "k8s_deployment",
		"StatefulSet":           "k8s_statefulset",
		"DaemonSet":             "k8s_daemonset",
		"ReplicaSet":            "k8s_replicaset",
		"Job":                   "k8s_job",
		"CronJob":               "k8s_cronjob",

		// Services
		"Service":               "k8s_service",
		"Ingress":               "k8s_ingress",
		"IngressClass":          "k8s_ingress_class",

		// Config & Storage
		"ConfigMap":             "k8s_configmap",
		"Secret":                "k8s_secret",
		"PersistentVolume":      "k8s_persistent_volume",
		"PersistentVolumeClaim": "k8s_persistent_volume_claim",
		"StorageClass":          "k8s_storage_class",

		// RBAC
		"ServiceAccount":        "k8s_service_account",
		"Role":                  "k8s_role",
		"RoleBinding":           "k8s_role_binding",
		"ClusterRole":           "k8s_cluster_role",
		"ClusterRoleBinding":    "k8s_cluster_role_binding",

		// Networking
		"NetworkPolicy":         "k8s_network_policy",
		"Endpoints":             "k8s_endpoints",

		// Cluster
		"Namespace":             "k8s_namespace",
		"Node":                  "k8s_node",
		"ResourceQuota":         "k8s_resource_quota",
		"LimitRange":            "k8s_limit_range",

		// Autoscaling
		"HorizontalPodAutoscaler": "k8s_hpa",
		"VerticalPodAutoscaler":   "k8s_vpa",
		"PodDisruptionBudget":     "k8s_pdb",

		// Custom Resources
		"CustomResourceDefinition": "k8s_crd",
	}

	// Check if we have a mapping
	if mapped, ok := typeMap[kind]; ok {
		return mapped
	}

	// Return kind prefixed with k8s_ if no specific mapping
	return fmt.Sprintf("k8s_%s", kind)
}

// ExtractResourceIdentifiers extracts identifiers from Kubernetes resource
func (m *ResourceMapper) ExtractResourceIdentifiers(k8sRes KubernetesResource) map[string]string {
	identifiers := make(map[string]string)

	// Add standard identifiers
	identifiers["name"] = k8sRes.Name
	identifiers["kind"] = k8sRes.Kind

	if k8sRes.Namespace != "" {
		identifiers["namespace"] = k8sRes.Namespace
	}

	// Add labels as identifiers (useful for matching)
	for key, value := range k8sRes.Labels {
		identifiers[fmt.Sprintf("label_%s", key)] = value
	}

	// Extract UID if available
	// Note: UID is typically only available in live cluster resources, not in manifests

	return identifiers
}

// CompareResources compares two Kubernetes resource configurations
func (m *ResourceMapper) CompareResources(k8sConfig, actualConfig map[string]interface{}) []iac.FieldChange {
	changes := make([]iac.FieldChange, 0)

	// Check all fields in K8s config
	for key, k8sValue := range k8sConfig {
		actualValue, exists := actualConfig[key]

		if !exists {
			changes = append(changes, iac.FieldChange{
				Field:       key,
				IaCValue:    k8sValue,
				ActualValue: nil,
				ChangeType:  "removed",
			})
			continue
		}

		if !m.valuesEqual(k8sValue, actualValue) {
			changes = append(changes, iac.FieldChange{
				Field:       key,
				IaCValue:    k8sValue,
				ActualValue: actualValue,
				ChangeType:  "modified",
			})
		}
	}

	// Check for fields in actual config not in K8s
	for key, actualValue := range actualConfig {
		if _, exists := k8sConfig[key]; !exists {
			// Skip computed fields that are expected to differ
			if m.isComputedField(key) {
				continue
			}

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

// isComputedField checks if a field is computed by Kubernetes and should be ignored in drift detection
func (m *ResourceMapper) isComputedField(fieldName string) bool {
	computedFields := map[string]bool{
		"uid":                true,
		"resourceVersion":    true,
		"selfLink":           true,
		"creationTimestamp":  true,
		"generation":         true,
		"managedFields":      true,
		"status":             true,
		"clusterIP":          true, // For Services
		"clusterIPs":         true,
		"internalTrafficPolicy": true,
		"ipFamilies":         true,
		"ipFamilyPolicy":     true,
		"sessionAffinity":    true,
	}

	return computedFields[fieldName]
}

// valuesEqual compares two values for equality
func (m *ResourceMapper) valuesEqual(v1, v2 interface{}) bool {
	return reflect.DeepEqual(v1, v2)
}

// NormalizeSpec normalizes a Kubernetes spec by removing computed fields
// This is useful for comparing manifest specs with live cluster specs
func (m *ResourceMapper) NormalizeSpec(spec map[string]interface{}) map[string]interface{} {
	normalized := make(map[string]interface{})

	for key, value := range spec {
		// Skip computed fields
		if m.isComputedField(key) {
			continue
		}

		// Recursively normalize nested maps
		if mapValue, ok := value.(map[string]interface{}); ok {
			normalized[key] = m.NormalizeSpec(mapValue)
		} else if sliceValue, ok := value.([]interface{}); ok {
			// Normalize each item in slice if it's a map
			normalizedSlice := make([]interface{}, len(sliceValue))
			for i, item := range sliceValue {
				if itemMap, ok := item.(map[string]interface{}); ok {
					normalizedSlice[i] = m.NormalizeSpec(itemMap)
				} else {
					normalizedSlice[i] = item
				}
			}
			normalized[key] = normalizedSlice
		} else {
			normalized[key] = value
		}
	}

	return normalized
}

// ExtractSecurityContext extracts security context from a workload resource
func (m *ResourceMapper) ExtractSecurityContext(k8sRes KubernetesResource) map[string]interface{} {
	securityContext := make(map[string]interface{})

	if k8sRes.Spec == nil {
		return securityContext
	}

	// Extract pod security context for workloads
	switch k8sRes.Kind {
	case KindDeployment, KindStatefulSet, KindDaemonSet, KindReplicaSet:
		if template, ok := k8sRes.Spec["template"].(map[string]interface{}); ok {
			if spec, ok := template["spec"].(map[string]interface{}); ok {
				if sc, ok := spec["securityContext"].(map[string]interface{}); ok {
					securityContext["podSecurityContext"] = sc
				}
				// Extract container security contexts
				if containers, ok := spec["containers"].([]interface{}); ok {
					containerContexts := make([]interface{}, 0)
					for _, container := range containers {
						if containerMap, ok := container.(map[string]interface{}); ok {
							if sc, ok := containerMap["securityContext"].(map[string]interface{}); ok {
								containerContexts = append(containerContexts, sc)
							}
						}
					}
					if len(containerContexts) > 0 {
						securityContext["containerSecurityContexts"] = containerContexts
					}
				}
			}
		}

	case KindPod:
		if sc, ok := k8sRes.Spec["securityContext"].(map[string]interface{}); ok {
			securityContext["podSecurityContext"] = sc
		}
	}

	return securityContext
}
