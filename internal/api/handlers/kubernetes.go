package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/pratik-mahalle/infraudit/internal/api/dto"
	"github.com/pratik-mahalle/infraudit/internal/api/middleware"
	"github.com/pratik-mahalle/infraudit/internal/pkg/errors"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
	"github.com/pratik-mahalle/infraudit/internal/pkg/utils"
	"github.com/pratik-mahalle/infraudit/internal/pkg/validator"
)

// KubernetesHandler handles Kubernetes-related API endpoints
type KubernetesHandler struct {
	logger    *logger.Logger
	validator *validator.Validator
}

// NewKubernetesHandler creates a new KubernetesHandler
func NewKubernetesHandler(log *logger.Logger, val *validator.Validator) *KubernetesHandler {
	return &KubernetesHandler{
		logger:    log,
		validator: val,
	}
}

// ListClusters returns all registered Kubernetes clusters
// @Summary List Kubernetes clusters
// @Description Get a list of all registered Kubernetes clusters for the user
// @Tags Kubernetes
// @Produce json
// @Success 200 {array} dto.K8sClusterDTO "List of Kubernetes clusters"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /kubernetes/clusters [get]
func (h *KubernetesHandler) ListClusters(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)
	h.logger.Infof("Listing clusters for user %d", userID)

	// For now, return mock data - this will be replaced with actual service calls
	clusters := []dto.K8sClusterDTO{
		{
			ID:          1,
			Name:        "dev-cluster",
			Context:     "dev-cluster-context",
			Server:      "https://dev-cluster.example.com:6443",
			Status:      "healthy",
			Version:     "v1.28.4",
			Nodes:       3,
			Pods:        42,
			Services:    15,
			Deployments: 12,
			Namespaces:  8,
			Provider:    "aws",
			Region:      "us-east-1",
			CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
		},
		{
			ID:          2,
			Name:        "staging-cluster",
			Context:     "staging-cluster-context",
			Server:      "https://staging-cluster.example.com:6443",
			Status:      "healthy",
			Version:     "v1.28.4",
			Nodes:       5,
			Pods:        78,
			Services:    25,
			Deployments: 22,
			Namespaces:  12,
			Provider:    "gcp",
			Region:      "us-central1",
			CreatedAt:   time.Now().Add(-60 * 24 * time.Hour),
		},
		{
			ID:          3,
			Name:        "production-cluster",
			Context:     "production-cluster-context",
			Server:      "https://prod-cluster.example.com:6443",
			Status:      "healthy",
			Version:     "v1.27.8",
			Nodes:       10,
			Pods:        256,
			Services:    50,
			Deployments: 45,
			Namespaces:  20,
			Provider:    "azure",
			Region:      "eastus",
			CreatedAt:   time.Now().Add(-120 * 24 * time.Hour),
		},
	}

	utils.WriteSuccess(w, http.StatusOK, clusters)
}

// GetCluster returns a specific Kubernetes cluster by ID
// @Summary Get Kubernetes cluster
// @Description Get details of a specific Kubernetes cluster
// @Tags Kubernetes
// @Produce json
// @Param id path int true "Cluster ID"
// @Success 200 {object} dto.K8sClusterDTO "Kubernetes cluster details"
// @Failure 404 {object} utils.ErrorResponse "Cluster not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /kubernetes/clusters/{id} [get]
func (h *KubernetesHandler) GetCluster(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.WriteError(w, errors.BadRequest("Invalid cluster ID"))
		return
	}

	// Mock data for now
	cluster := dto.K8sClusterDTO{
		ID:          id,
		Name:        "dev-cluster",
		Context:     "dev-cluster-context",
		Server:      "https://dev-cluster.example.com:6443",
		Status:      "healthy",
		Version:     "v1.28.4",
		Nodes:       3,
		Pods:        42,
		Services:    15,
		Deployments: 12,
		Namespaces:  8,
		Provider:    "aws",
		Region:      "us-east-1",
		CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
	}

	utils.WriteSuccess(w, http.StatusOK, cluster)
}

// RegisterCluster registers a new Kubernetes cluster
// @Summary Register Kubernetes cluster
// @Description Register a new Kubernetes cluster with kubeconfig
// @Tags Kubernetes
// @Accept json
// @Produce json
// @Param request body dto.RegisterK8sClusterRequest true "Cluster registration request"
// @Success 201 {object} dto.K8sClusterDTO "Registered cluster"
// @Failure 400 {object} utils.ErrorResponse "Invalid request"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /kubernetes/clusters [post]
func (h *KubernetesHandler) RegisterCluster(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterK8sClusterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, errors.BadRequest("Invalid request body"))
		return
	}

	if errs := h.validator.Validate(req); len(errs) > 0 {
		utils.WriteError(w, errors.ValidationError("Validation failed", errs))
		return
	}

	// Mock response for now
	cluster := dto.K8sClusterDTO{
		ID:        4,
		Name:      req.Name,
		Context:   req.Context,
		Status:    "connecting",
		CreatedAt: time.Now(),
	}

	utils.WriteSuccess(w, http.StatusCreated, cluster)
}

// DeleteCluster deletes a Kubernetes cluster
// @Summary Delete Kubernetes cluster
// @Description Remove a registered Kubernetes cluster
// @Tags Kubernetes
// @Produce json
// @Param id path int true "Cluster ID"
// @Success 200 {object} utils.SuccessResponse "Cluster deleted successfully"
// @Failure 404 {object} utils.ErrorResponse "Cluster not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /kubernetes/clusters/{id} [delete]
func (h *KubernetesHandler) DeleteCluster(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	_, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.WriteError(w, errors.BadRequest("Invalid cluster ID"))
		return
	}

	utils.WriteSuccessWithMessage(w, http.StatusOK, "Cluster deleted successfully", nil)
}

// GetClusterStats returns aggregated stats for all clusters
// @Summary Get Kubernetes cluster statistics
// @Description Get aggregated statistics for all Kubernetes clusters
// @Tags Kubernetes
// @Produce json
// @Success 200 {object} dto.K8sClusterStatsDTO "Cluster statistics"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /kubernetes/stats [get]
func (h *KubernetesHandler) GetClusterStats(w http.ResponseWriter, r *http.Request) {
	stats := dto.K8sClusterStatsDTO{
		TotalClusters:    3,
		HealthyClusters:  3,
		TotalNodes:       18,
		TotalPods:        376,
		TotalServices:    90,
		TotalDeployments: 79,
	}

	utils.WriteSuccess(w, http.StatusOK, stats)
}

// ListNamespaces returns namespaces for a cluster
// @Summary List Kubernetes namespaces
// @Description Get a list of namespaces in a specific cluster
// @Tags Kubernetes
// @Produce json
// @Param clusterId path int true "Cluster ID"
// @Success 200 {array} dto.K8sNamespaceDTO "List of namespaces"
// @Failure 404 {object} utils.ErrorResponse "Cluster not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /kubernetes/clusters/{clusterId}/namespaces [get]
func (h *KubernetesHandler) ListNamespaces(w http.ResponseWriter, r *http.Request) {
	namespaces := []dto.K8sNamespaceDTO{
		{Name: "default", Status: "Active", CreatedAt: time.Now().Add(-30 * 24 * time.Hour)},
		{Name: "kube-system", Status: "Active", CreatedAt: time.Now().Add(-30 * 24 * time.Hour)},
		{Name: "kube-public", Status: "Active", CreatedAt: time.Now().Add(-30 * 24 * time.Hour)},
		{Name: "monitoring", Status: "Active", CreatedAt: time.Now().Add(-7 * 24 * time.Hour)},
		{Name: "logging", Status: "Active", CreatedAt: time.Now().Add(-7 * 24 * time.Hour)},
		{Name: "app", Status: "Active", CreatedAt: time.Now().Add(-14 * 24 * time.Hour)},
	}

	utils.WriteSuccess(w, http.StatusOK, namespaces)
}

// ListDeployments returns deployments for a cluster
// @Summary List Kubernetes deployments
// @Description Get a list of deployments in a specific cluster
// @Tags Kubernetes
// @Produce json
// @Param clusterId path int true "Cluster ID"
// @Param namespace query string false "Filter by namespace"
// @Success 200 {array} dto.K8sDeploymentDTO "List of deployments"
// @Failure 404 {object} utils.ErrorResponse "Cluster not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /kubernetes/clusters/{clusterId}/deployments [get]
func (h *KubernetesHandler) ListDeployments(w http.ResponseWriter, r *http.Request) {
	deployments := []dto.K8sDeploymentDTO{
		{
			Name:              "web-app",
			Namespace:         "app",
			Replicas:          3,
			ReadyReplicas:     3,
			AvailableReplicas: 3,
			Images:            []string{"nginx:1.25"},
			CreatedAt:         time.Now().Add(-7 * 24 * time.Hour),
		},
		{
			Name:              "api-server",
			Namespace:         "app",
			Replicas:          5,
			ReadyReplicas:     5,
			AvailableReplicas: 5,
			Images:            []string{"api-server:v1.2.3"},
			CreatedAt:         time.Now().Add(-5 * 24 * time.Hour),
		},
		{
			Name:              "prometheus",
			Namespace:         "monitoring",
			Replicas:          1,
			ReadyReplicas:     1,
			AvailableReplicas: 1,
			Images:            []string{"prom/prometheus:v2.48.0"},
			CreatedAt:         time.Now().Add(-14 * 24 * time.Hour),
		},
	}

	utils.WriteSuccess(w, http.StatusOK, deployments)
}

// ListPods returns pods for a cluster
// @Summary List Kubernetes pods
// @Description Get a list of pods in a specific cluster
// @Tags Kubernetes
// @Produce json
// @Param clusterId path int true "Cluster ID"
// @Param namespace query string false "Filter by namespace"
// @Success 200 {array} dto.K8sPodDTO "List of pods"
// @Failure 404 {object} utils.ErrorResponse "Cluster not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /kubernetes/clusters/{clusterId}/pods [get]
func (h *KubernetesHandler) ListPods(w http.ResponseWriter, r *http.Request) {
	pods := []dto.K8sPodDTO{
		{
			Name:       "web-app-abc123",
			Namespace:  "app",
			Status:     "Running",
			Node:       "node-1",
			Containers: []string{"nginx"},
			CreatedAt:  time.Now().Add(-2 * 24 * time.Hour),
		},
		{
			Name:       "api-server-xyz789",
			Namespace:  "app",
			Status:     "Running",
			Node:       "node-2",
			Containers: []string{"api-server"},
			CreatedAt:  time.Now().Add(-1 * 24 * time.Hour),
		},
		{
			Name:       "prometheus-0",
			Namespace:  "monitoring",
			Status:     "Running",
			Node:       "node-1",
			Containers: []string{"prometheus"},
			CreatedAt:  time.Now().Add(-14 * 24 * time.Hour),
		},
	}

	utils.WriteSuccess(w, http.StatusOK, pods)
}

// ListServices returns services for a cluster
// @Summary List Kubernetes services
// @Description Get a list of services in a specific cluster
// @Tags Kubernetes
// @Produce json
// @Param clusterId path int true "Cluster ID"
// @Param namespace query string false "Filter by namespace"
// @Success 200 {array} dto.K8sServiceDTO "List of services"
// @Failure 404 {object} utils.ErrorResponse "Cluster not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /kubernetes/clusters/{clusterId}/services [get]
func (h *KubernetesHandler) ListServices(w http.ResponseWriter, r *http.Request) {
	services := []dto.K8sServiceDTO{
		{
			Name:      "web-app",
			Namespace: "app",
			Type:      "ClusterIP",
			ClusterIP: "10.0.0.50",
			Ports:     []dto.K8sPortDTO{{Port: 80, TargetPort: 8080, Protocol: "TCP"}},
			CreatedAt: time.Now().Add(-7 * 24 * time.Hour),
		},
		{
			Name:       "api-server",
			Namespace:  "app",
			Type:       "LoadBalancer",
			ClusterIP:  "10.0.0.51",
			ExternalIP: "34.123.45.67",
			Ports:      []dto.K8sPortDTO{{Port: 443, TargetPort: 8443, Protocol: "TCP"}},
			CreatedAt:  time.Now().Add(-5 * 24 * time.Hour),
		},
		{
			Name:      "prometheus",
			Namespace: "monitoring",
			Type:      "ClusterIP",
			ClusterIP: "10.0.0.100",
			Ports:     []dto.K8sPortDTO{{Port: 9090, TargetPort: 9090, Protocol: "TCP"}},
			CreatedAt: time.Now().Add(-14 * 24 * time.Hour),
		},
	}

	utils.WriteSuccess(w, http.StatusOK, services)
}

// SyncCluster triggers a sync for a Kubernetes cluster
// @Summary Sync Kubernetes cluster
// @Description Trigger a sync to refresh cluster data
// @Tags Kubernetes
// @Produce json
// @Param id path int true "Cluster ID"
// @Success 200 {object} utils.SuccessResponse "Sync initiated"
// @Failure 404 {object} utils.ErrorResponse "Cluster not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /kubernetes/clusters/{id}/sync [post]
func (h *KubernetesHandler) SyncCluster(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	_, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.WriteError(w, errors.BadRequest("Invalid cluster ID"))
		return
	}

	utils.WriteSuccessWithMessage(w, http.StatusOK, "Cluster sync initiated", nil)
}
