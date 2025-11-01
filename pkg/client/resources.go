package client

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// ResourceService handles resource-related API calls
type ResourceService struct {
	client *Client
}

// CreateResourceRequest represents a request to create a resource
type CreateResourceRequest struct {
	ProviderID   int64                  `json:"provider_id"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   string                 `json:"resource_id"`
	Name         string                 `json:"name"`
	Region       string                 `json:"region,omitempty"`
	Status       string                 `json:"status,omitempty"`
	Tags         map[string]string      `json:"tags,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateResourceRequest represents a request to update a resource
type UpdateResourceRequest struct {
	Name     *string                `json:"name,omitempty"`
	Status   *string                `json:"status,omitempty"`
	Tags     map[string]string      `json:"tags,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ResourceListOptions contains options for listing resources
type ResourceListOptions struct {
	ListOptions
	ProviderID   *int64  `json:"provider_id,omitempty"`
	ResourceType *string `json:"resource_type,omitempty"`
	Status       *string `json:"status,omitempty"`
	Region       *string `json:"region,omitempty"`
}

// List retrieves a list of resources
func (s *ResourceService) List(ctx context.Context, opts *ResourceListOptions) ([]Resource, error) {
	query := url.Values{}

	if opts != nil {
		if opts.Page > 0 {
			query.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			query.Set("page_size", strconv.Itoa(opts.PageSize))
		}
		if opts.Sort != "" {
			query.Set("sort", opts.Sort)
		}
		if opts.Order != "" {
			query.Set("order", opts.Order)
		}
		if opts.Search != "" {
			query.Set("search", opts.Search)
		}
		if opts.ProviderID != nil {
			query.Set("provider_id", strconv.FormatInt(*opts.ProviderID, 10))
		}
		if opts.ResourceType != nil {
			query.Set("resource_type", *opts.ResourceType)
		}
		if opts.Status != nil {
			query.Set("status", *opts.Status)
		}
		if opts.Region != nil {
			query.Set("region", *opts.Region)
		}
	}

	path := "/api/resources"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	var resources []Resource
	if err := s.client.doRequest(ctx, "GET", path, nil, &resources); err != nil {
		return nil, err
	}

	return resources, nil
}

// Get retrieves a single resource by ID
func (s *ResourceService) Get(ctx context.Context, id int64) (*Resource, error) {
	path := fmt.Sprintf("/api/resources/%d", id)

	var resource Resource
	if err := s.client.doRequest(ctx, "GET", path, nil, &resource); err != nil {
		return nil, err
	}

	return &resource, nil
}

// Create creates a new resource
func (s *ResourceService) Create(ctx context.Context, req CreateResourceRequest) (*Resource, error) {
	var resource Resource
	if err := s.client.doRequest(ctx, "POST", "/api/resources", req, &resource); err != nil {
		return nil, err
	}

	return &resource, nil
}

// Update updates an existing resource
func (s *ResourceService) Update(ctx context.Context, id int64, req UpdateResourceRequest) (*Resource, error) {
	path := fmt.Sprintf("/api/resources/%d", id)

	var resource Resource
	if err := s.client.doRequest(ctx, "PATCH", path, req, &resource); err != nil {
		return nil, err
	}

	return &resource, nil
}

// Delete deletes a resource
func (s *ResourceService) Delete(ctx context.Context, id int64) error {
	path := fmt.Sprintf("/api/resources/%d", id)
	return s.client.doRequest(ctx, "DELETE", path, nil, nil)
}

// GetCost retrieves cost information for a resource
func (s *ResourceService) GetCost(ctx context.Context, id int64) (*ResourceCost, error) {
	path := fmt.Sprintf("/api/resources/%d/cost", id)

	var cost ResourceCost
	if err := s.client.doRequest(ctx, "GET", path, nil, &cost); err != nil {
		return nil, err
	}

	return &cost, nil
}
