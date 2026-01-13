package kubernetes

// K8sResource represents a generic Kubernetes resource
type K8sResource struct {
	APIVersion string                 `yaml:"apiVersion" json:"apiVersion"`
	Kind       string                 `yaml:"kind" json:"kind"`
	Metadata   K8sMetadata            `yaml:"metadata" json:"metadata"`
	Spec       map[string]interface{} `yaml:"spec,omitempty" json:"spec,omitempty"`
	Data       map[string]interface{} `yaml:"data,omitempty" json:"data,omitempty"`
	StringData map[string]string      `yaml:"stringData,omitempty" json:"stringData,omitempty"`
}

// K8sMetadata represents Kubernetes metadata
type K8sMetadata struct {
	Name        string            `yaml:"name" json:"name"`
	Namespace   string            `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty" json:"annotations,omitempty"`
	UID         string            `yaml:"uid,omitempty" json:"uid,omitempty"`
}

// ParsedKubernetes represents parsed Kubernetes manifests
type ParsedKubernetes struct {
	Resources []KubernetesResource `json:"resources"`
	Namespace string               `json:"namespace,omitempty"`
}

// KubernetesResource represents a parsed Kubernetes resource
type KubernetesResource struct {
	APIVersion  string                 `json:"api_version"`
	Kind        string                 `json:"kind"`
	Name        string                 `json:"name"`
	Namespace   string                 `json:"namespace,omitempty"`
	Labels      map[string]string      `json:"labels,omitempty"`
	Annotations map[string]string      `json:"annotations,omitempty"`
	Spec        map[string]interface{} `json:"spec,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
}

// ParseResult holds the result of parsing Kubernetes manifests
type ParseResult struct {
	Parsed *ParsedKubernetes
	Errors []error
}

// HasErrors returns true if there are any errors
func (r *ParseResult) HasErrors() bool {
	if r == nil {
		return false
	}
	return len(r.Errors) > 0
}

// ErrorMessages returns all error messages
func (r *ParseResult) ErrorMessages() []string {
	if r == nil {
		return nil
	}
	messages := make([]string, 0, len(r.Errors))
	for _, err := range r.Errors {
		messages = append(messages, err.Error())
	}
	return messages
}

// Common Kubernetes resource kinds
const (
	KindPod                   = "Pod"
	KindDeployment            = "Deployment"
	KindStatefulSet           = "StatefulSet"
	KindDaemonSet             = "DaemonSet"
	KindReplicaSet            = "ReplicaSet"
	KindService               = "Service"
	KindIngress               = "Ingress"
	KindConfigMap             = "ConfigMap"
	KindSecret                = "Secret"
	KindPersistentVolume      = "PersistentVolume"
	KindPersistentVolumeClaim = "PersistentVolumeClaim"
	KindServiceAccount        = "ServiceAccount"
	KindRole                  = "Role"
	KindRoleBinding           = "RoleBinding"
	KindClusterRole           = "ClusterRole"
	KindClusterRoleBinding    = "ClusterRoleBinding"
	KindNamespace             = "Namespace"
	KindNetworkPolicy         = "NetworkPolicy"
	KindHorizontalPodAutoscaler = "HorizontalPodAutoscaler"
)
