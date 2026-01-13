package dto

import "time"

// K8sClusterDTO represents a Kubernetes cluster
type K8sClusterDTO struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Context     string     `json:"context,omitempty"`
	Server      string     `json:"server,omitempty"`
	Status      string     `json:"status"`
	Version     string     `json:"version,omitempty"`
	Nodes       int        `json:"nodes"`
	Pods        int        `json:"pods"`
	Services    int        `json:"services"`
	Deployments int        `json:"deployments"`
	Namespaces  int        `json:"namespaces"`
	Provider    string     `json:"provider,omitempty"`
	Region      string     `json:"region,omitempty"`
	LastSynced  *time.Time `json:"lastSynced,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
}

// RegisterK8sClusterRequest represents a request to register a cluster
type RegisterK8sClusterRequest struct {
	Name       string `json:"name" validate:"required"`
	Kubeconfig string `json:"kubeconfig" validate:"required"`
	Context    string `json:"context,omitempty"`
}

// K8sNamespaceDTO represents a Kubernetes namespace
type K8sNamespaceDTO struct {
	Name      string            `json:"name"`
	Status    string            `json:"status"`
	Labels    map[string]string `json:"labels,omitempty"`
	CreatedAt time.Time         `json:"createdAt"`
}

// K8sDeploymentDTO represents a Kubernetes deployment
type K8sDeploymentDTO struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	Replicas          int32             `json:"replicas"`
	ReadyReplicas     int32             `json:"readyReplicas"`
	AvailableReplicas int32             `json:"availableReplicas"`
	Labels            map[string]string `json:"labels,omitempty"`
	Images            []string          `json:"images"`
	CreatedAt         time.Time         `json:"createdAt"`
}

// K8sPodDTO represents a Kubernetes pod
type K8sPodDTO struct {
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	Status     string            `json:"status"`
	Node       string            `json:"node,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	Containers []string          `json:"containers"`
	CreatedAt  time.Time         `json:"createdAt"`
}

// K8sServiceDTO represents a Kubernetes service
type K8sServiceDTO struct {
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	Type       string            `json:"type"`
	ClusterIP  string            `json:"clusterIP,omitempty"`
	ExternalIP string            `json:"externalIP,omitempty"`
	Ports      []K8sPortDTO      `json:"ports"`
	Labels     map[string]string `json:"labels,omitempty"`
	CreatedAt  time.Time         `json:"createdAt"`
}

// K8sPortDTO represents a port in a Kubernetes service
type K8sPortDTO struct {
	Name       string `json:"name,omitempty"`
	Port       int32  `json:"port"`
	TargetPort int32  `json:"targetPort"`
	Protocol   string `json:"protocol"`
}

// K8sClusterStatsDTO represents cluster statistics
type K8sClusterStatsDTO struct {
	TotalClusters    int `json:"totalClusters"`
	HealthyClusters  int `json:"healthyClusters"`
	TotalNodes       int `json:"totalNodes"`
	TotalPods        int `json:"totalPods"`
	TotalServices    int `json:"totalServices"`
	TotalDeployments int `json:"totalDeployments"`
}
