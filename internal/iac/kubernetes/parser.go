package kubernetes

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Parser handles parsing of Kubernetes manifest files
type Parser struct{}

// NewParser creates a new Kubernetes parser
func NewParser() *Parser {
	return &Parser{}
}

// ParseFile parses a single Kubernetes manifest file
func (p *Parser) ParseFile(filename string) (*ParseResult, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return p.Parse(content, filename)
}

// Parse parses Kubernetes YAML content (can contain multiple resources separated by ---)
func (p *Parser) Parse(content []byte, filename string) (*ParseResult, error) {
	result := &ParseResult{
		Parsed: &ParsedKubernetes{
			Resources: make([]KubernetesResource, 0),
		},
		Errors: make([]error, 0),
	}

	// Split by YAML document separator (---)
	documents := p.splitYAMLDocuments(content)

	for i, doc := range documents {
		// Skip empty documents
		if len(bytes.TrimSpace(doc)) == 0 {
			continue
		}

		// Parse the document
		resource, err := p.parseDocument(doc)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("document %d in %s: %w", i+1, filename, err))
			continue
		}

		if resource != nil {
			result.Parsed.Resources = append(result.Parsed.Resources, *resource)
		}
	}

	return result, nil
}

// ParseDirectory parses all Kubernetes manifest files in a directory
func (p *Parser) ParseDirectory(dir string) (*ParseResult, error) {
	combinedResult := &ParseResult{
		Parsed: &ParsedKubernetes{
			Resources: make([]KubernetesResource, 0),
		},
		Errors: make([]error, 0),
	}

	// Find all YAML files
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}

		// Parse the file
		result, err := p.ParseFile(path)
		if err != nil {
			combinedResult.Errors = append(combinedResult.Errors, fmt.Errorf("%s: %w", path, err))
			return nil // Continue processing other files
		}

		// Merge results
		combinedResult.Parsed.Resources = append(combinedResult.Parsed.Resources, result.Parsed.Resources...)
		combinedResult.Errors = append(combinedResult.Errors, result.Errors...)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return combinedResult, nil
}

// splitYAMLDocuments splits a YAML file into multiple documents
func (p *Parser) splitYAMLDocuments(content []byte) [][]byte {
	documents := make([][]byte, 0)
	scanner := bufio.NewScanner(bytes.NewReader(content))

	var currentDoc bytes.Buffer

	for scanner.Scan() {
		line := scanner.Text()

		// Check if this is a document separator
		if strings.TrimSpace(line) == "---" {
			// Save current document if it has content
			if currentDoc.Len() > 0 {
				documents = append(documents, currentDoc.Bytes())
				currentDoc.Reset()
			}
			continue
		}

		currentDoc.WriteString(line)
		currentDoc.WriteByte('\n')
	}

	// Add the last document
	if currentDoc.Len() > 0 {
		documents = append(documents, currentDoc.Bytes())
	}

	return documents
}

// parseDocument parses a single YAML document into a Kubernetes resource
func (p *Parser) parseDocument(doc []byte) (*KubernetesResource, error) {
	var k8sRes K8sResource

	if err := yaml.Unmarshal(doc, &k8sRes); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate required fields
	if k8sRes.APIVersion == "" {
		return nil, fmt.Errorf("apiVersion is required")
	}
	if k8sRes.Kind == "" {
		return nil, fmt.Errorf("kind is required")
	}
	if k8sRes.Metadata.Name == "" {
		return nil, fmt.Errorf("metadata.name is required")
	}

	// Convert to our internal format
	resource := &KubernetesResource{
		APIVersion:  k8sRes.APIVersion,
		Kind:        k8sRes.Kind,
		Name:        k8sRes.Metadata.Name,
		Namespace:   k8sRes.Metadata.Namespace,
		Labels:      k8sRes.Metadata.Labels,
		Annotations: k8sRes.Metadata.Annotations,
		Spec:        k8sRes.Spec,
		Data:        k8sRes.Data,
	}

	// Set default namespace if not specified
	if resource.Namespace == "" && p.requiresNamespace(resource.Kind) {
		resource.Namespace = "default"
	}

	return resource, nil
}

// requiresNamespace checks if a resource kind requires a namespace
func (p *Parser) requiresNamespace(kind string) bool {
	// Cluster-scoped resources don't have a namespace
	clusterScoped := map[string]bool{
		"Namespace":             true,
		"ClusterRole":           true,
		"ClusterRoleBinding":    true,
		"PersistentVolume":      true,
		"StorageClass":          true,
		"CustomResourceDefinition": true,
		"Node":                  true,
	}

	return !clusterScoped[kind]
}

// GetResourceByName finds a resource by name and kind
func (p *Parser) GetResourceByName(parsed *ParsedKubernetes, name, kind string) *KubernetesResource {
	for i := range parsed.Resources {
		res := &parsed.Resources[i]
		if res.Name == name && res.Kind == kind {
			return res
		}
	}
	return nil
}

// GetResourcesByKind returns all resources of a specific kind
func (p *Parser) GetResourcesByKind(parsed *ParsedKubernetes, kind string) []KubernetesResource {
	resources := make([]KubernetesResource, 0)
	for _, res := range parsed.Resources {
		if res.Kind == kind {
			resources = append(resources, res)
		}
	}
	return resources
}

// GetResourcesByNamespace returns all resources in a namespace
func (p *Parser) GetResourcesByNamespace(parsed *ParsedKubernetes, namespace string) []KubernetesResource {
	resources := make([]KubernetesResource, 0)
	for _, res := range parsed.Resources {
		if res.Namespace == namespace {
			resources = append(resources, res)
		}
	}
	return resources
}

// ValidateResource performs basic validation on a Kubernetes resource
func (p *Parser) ValidateResource(resource *KubernetesResource) []error {
	errors := make([]error, 0)

	// Required fields
	if resource.APIVersion == "" {
		errors = append(errors, fmt.Errorf("apiVersion is required"))
	}
	if resource.Kind == "" {
		errors = append(errors, fmt.Errorf("kind is required"))
	}
	if resource.Name == "" {
		errors = append(errors, fmt.Errorf("metadata.name is required"))
	}

	// Namespace validation
	if p.requiresNamespace(resource.Kind) && resource.Namespace == "" {
		errors = append(errors, fmt.Errorf("namespace is required for %s", resource.Kind))
	}

	// Name validation (RFC 1123 DNS subdomain)
	if !isValidK8sName(resource.Name) {
		errors = append(errors, fmt.Errorf("invalid resource name: %s", resource.Name))
	}

	// Kind-specific validation
	switch resource.Kind {
	case KindDeployment, KindStatefulSet, KindDaemonSet:
		if resource.Spec == nil {
			errors = append(errors, fmt.Errorf("%s requires spec", resource.Kind))
		}

	case KindService:
		if resource.Spec == nil {
			errors = append(errors, fmt.Errorf("Service requires spec"))
		}

	case KindConfigMap, KindSecret:
		if resource.Data == nil {
			errors = append(errors, fmt.Errorf("%s requires data", resource.Kind))
		}
	}

	return errors
}

// isValidK8sName checks if a name is valid according to Kubernetes naming rules
// RFC 1123 DNS subdomain: lowercase alphanumeric characters, '-' or '.', max 253 chars
func isValidK8sName(name string) bool {
	if len(name) == 0 || len(name) > 253 {
		return false
	}

	for i, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '.') {
			return false
		}
		// Must start and end with alphanumeric
		if i == 0 || i == len(name)-1 {
			if !(c >= 'a' && c <= 'z') && !(c >= '0' && c <= '9') {
				return false
			}
		}
	}

	return true
}

// ExtractContainerImages extracts all container images from workload resources
func (p *Parser) ExtractContainerImages(resource *KubernetesResource) []string {
	images := make([]string, 0)

	if resource.Spec == nil {
		return images
	}

	// Check for pod template in workloads (Deployment, StatefulSet, DaemonSet)
	var containers []interface{}

	switch resource.Kind {
	case KindDeployment, KindStatefulSet, KindDaemonSet, KindReplicaSet:
		// spec.template.spec.containers
		if template, ok := resource.Spec["template"].(map[string]interface{}); ok {
			if spec, ok := template["spec"].(map[string]interface{}); ok {
				if c, ok := spec["containers"].([]interface{}); ok {
					containers = c
				}
			}
		}

	case KindPod:
		// spec.containers
		if c, ok := resource.Spec["containers"].([]interface{}); ok {
			containers = c
		}
	}

	// Extract image from each container
	for _, container := range containers {
		if containerMap, ok := container.(map[string]interface{}); ok {
			if image, ok := containerMap["image"].(string); ok {
				images = append(images, image)
			}
		}
	}

	return images
}
