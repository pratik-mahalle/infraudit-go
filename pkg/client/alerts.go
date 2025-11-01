package client

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// AlertService handles alert-related API calls
type AlertService struct {
	client *Client
}

// CreateAlertRequest represents a request to create an alert
type CreateAlertRequest struct {
	ResourceID  *int64                 `json:"resource_id,omitempty"`
	Type        string                 `json:"type"`     // security, compliance, performance
	Severity    string                 `json:"severity"` // critical, high, medium, low
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateAlertRequest represents a request to update an alert
type UpdateAlertRequest struct {
	Status      *string                `json:"status,omitempty"` // open, acknowledged, resolved
	Description *string                `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// AlertListOptions contains options for listing alerts
type AlertListOptions struct {
	ListOptions
	Type       *string `json:"type,omitempty"`
	Severity   *string `json:"severity,omitempty"`
	Status     *string `json:"status,omitempty"`
	ResourceID *int64  `json:"resource_id,omitempty"`
}

// List retrieves a list of alerts
func (s *AlertService) List(ctx context.Context, opts *AlertListOptions) ([]Alert, error) {
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
		if opts.Type != nil {
			query.Set("type", *opts.Type)
		}
		if opts.Severity != nil {
			query.Set("severity", *opts.Severity)
		}
		if opts.Status != nil {
			query.Set("status", *opts.Status)
		}
		if opts.ResourceID != nil {
			query.Set("resource_id", strconv.FormatInt(*opts.ResourceID, 10))
		}
	}

	path := "/api/alerts"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	var alerts []Alert
	if err := s.client.doRequest(ctx, "GET", path, nil, &alerts); err != nil {
		return nil, err
	}

	return alerts, nil
}

// Get retrieves a single alert by ID
func (s *AlertService) Get(ctx context.Context, id int64) (*Alert, error) {
	path := fmt.Sprintf("/api/alerts/%d", id)

	var alert Alert
	if err := s.client.doRequest(ctx, "GET", path, nil, &alert); err != nil {
		return nil, err
	}

	return &alert, nil
}

// Create creates a new alert
func (s *AlertService) Create(ctx context.Context, req CreateAlertRequest) (*Alert, error) {
	var alert Alert
	if err := s.client.doRequest(ctx, "POST", "/api/alerts", req, &alert); err != nil {
		return nil, err
	}

	return &alert, nil
}

// Update updates an existing alert
func (s *AlertService) Update(ctx context.Context, id int64, req UpdateAlertRequest) (*Alert, error) {
	path := fmt.Sprintf("/api/alerts/%d", id)

	var alert Alert
	if err := s.client.doRequest(ctx, "PATCH", path, req, &alert); err != nil {
		return nil, err
	}

	return &alert, nil
}

// Acknowledge acknowledges an alert
func (s *AlertService) Acknowledge(ctx context.Context, id int64) (*Alert, error) {
	status := "acknowledged"
	return s.Update(ctx, id, UpdateAlertRequest{Status: &status})
}

// Resolve resolves an alert
func (s *AlertService) Resolve(ctx context.Context, id int64) (*Alert, error) {
	status := "resolved"
	return s.Update(ctx, id, UpdateAlertRequest{Status: &status})
}

// Delete deletes an alert
func (s *AlertService) Delete(ctx context.Context, id int64) error {
	path := fmt.Sprintf("/api/alerts/%d", id)
	return s.client.doRequest(ctx, "DELETE", path, nil, nil)
}
