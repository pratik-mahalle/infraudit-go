package client

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// ProviderService handles provider-related API calls
type ProviderService struct {
	client *Client
}

// CreateProviderRequest represents a request to create a provider
type CreateProviderRequest struct {
	Name         string                 `json:"name"`
	ProviderType string                 `json:"provider_type"` // aws, gcp, azure
	Credentials  map[string]interface{} `json:"credentials"`
	Config       map[string]interface{} `json:"config,omitempty"`
}

// UpdateProviderRequest represents a request to update a provider
type UpdateProviderRequest struct {
	Name        *string                `json:"name,omitempty"`
	Credentials map[string]interface{} `json:"credentials,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

// ProviderListOptions contains options for listing providers
type ProviderListOptions struct {
	ListOptions
	ProviderType *string `json:"provider_type,omitempty"`
	Status       *string `json:"status,omitempty"`
}

// SyncResult represents the result of a provider sync operation
type SyncResult struct {
	ProviderID       int64  `json:"provider_id"`
	ResourcesFound   int    `json:"resources_found"`
	ResourcesCreated int    `json:"resources_created"`
	ResourcesUpdated int    `json:"resources_updated"`
	Errors           []string `json:"errors,omitempty"`
	Status           string `json:"status"`
}

// List retrieves a list of providers
func (s *ProviderService) List(ctx context.Context, opts *ProviderListOptions) ([]Provider, error) {
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
		if opts.ProviderType != nil {
			query.Set("provider_type", *opts.ProviderType)
		}
		if opts.Status != nil {
			query.Set("status", *opts.Status)
		}
	}

	path := "/api/providers"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	var providers []Provider
	if err := s.client.doRequest(ctx, "GET", path, nil, &providers); err != nil {
		return nil, err
	}

	return providers, nil
}

// Get retrieves a single provider by ID
func (s *ProviderService) Get(ctx context.Context, id int64) (*Provider, error) {
	path := fmt.Sprintf("/api/providers/%d", id)

	var provider Provider
	if err := s.client.doRequest(ctx, "GET", path, nil, &provider); err != nil {
		return nil, err
	}

	return &provider, nil
}

// Create creates a new provider
func (s *ProviderService) Create(ctx context.Context, req CreateProviderRequest) (*Provider, error) {
	var provider Provider
	if err := s.client.doRequest(ctx, "POST", "/api/providers", req, &provider); err != nil {
		return nil, err
	}

	return &provider, nil
}

// Update updates an existing provider
func (s *ProviderService) Update(ctx context.Context, id int64, req UpdateProviderRequest) (*Provider, error) {
	path := fmt.Sprintf("/api/providers/%d", id)

	var provider Provider
	if err := s.client.doRequest(ctx, "PATCH", path, req, &provider); err != nil {
		return nil, err
	}

	return &provider, nil
}

// Delete deletes a provider
func (s *ProviderService) Delete(ctx context.Context, id int64) error {
	path := fmt.Sprintf("/api/providers/%d", id)
	return s.client.doRequest(ctx, "DELETE", path, nil, nil)
}

// Sync triggers a sync operation for a provider
func (s *ProviderService) Sync(ctx context.Context, id int64) (*SyncResult, error) {
	path := fmt.Sprintf("/api/providers/%d/sync", id)

	var result SyncResult
	if err := s.client.doRequest(ctx, "POST", path, nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// TestConnection tests the connection to a provider
func (s *ProviderService) TestConnection(ctx context.Context, id int64) (bool, error) {
	path := fmt.Sprintf("/api/providers/%d/test", id)

	var result struct {
		Success bool   `json:"success"`
		Message string `json:"message,omitempty"`
	}

	if err := s.client.doRequest(ctx, "POST", path, nil, &result); err != nil {
		return false, err
	}

	return result.Success, nil
}
