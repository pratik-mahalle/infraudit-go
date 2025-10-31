package client

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// DriftService handles drift detection-related API calls
type DriftService struct {
	client *Client
}

// DriftListOptions contains options for listing drifts
type DriftListOptions struct {
	ListOptions
	DriftType  *string `json:"drift_type,omitempty"`  // configuration, security, compliance
	Severity   *string `json:"severity,omitempty"`    // critical, high, medium, low
	Status     *string `json:"status,omitempty"`      // detected, investigating, resolved
	ResourceID *int64  `json:"resource_id,omitempty"`
}

// UpdateDriftRequest represents a request to update a drift
type UpdateDriftRequest struct {
	Status *string `json:"status,omitempty"` // detected, investigating, resolved
}

// List retrieves a list of drifts
func (s *DriftService) List(ctx context.Context, opts *DriftListOptions) ([]Drift, error) {
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
		if opts.DriftType != nil {
			query.Set("drift_type", *opts.DriftType)
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

	path := "/api/drifts"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	var drifts []Drift
	if err := s.client.doRequest(ctx, "GET", path, nil, &drifts); err != nil {
		return nil, err
	}

	return drifts, nil
}

// Get retrieves a single drift by ID
func (s *DriftService) Get(ctx context.Context, id int64) (*Drift, error) {
	path := fmt.Sprintf("/api/drifts/%d", id)

	var drift Drift
	if err := s.client.doRequest(ctx, "GET", path, nil, &drift); err != nil {
		return nil, err
	}

	return &drift, nil
}

// Resolve marks a drift as resolved
func (s *DriftService) Resolve(ctx context.Context, id int64) (*Drift, error) {
	path := fmt.Sprintf("/api/drifts/%d", id)
	status := "resolved"

	var drift Drift
	if err := s.client.doRequest(ctx, "PATCH", path, UpdateDriftRequest{Status: &status}, &drift); err != nil {
		return nil, err
	}

	return &drift, nil
}

// Investigate marks a drift as under investigation
func (s *DriftService) Investigate(ctx context.Context, id int64) (*Drift, error) {
	path := fmt.Sprintf("/api/drifts/%d", id)
	status := "investigating"

	var drift Drift
	if err := s.client.doRequest(ctx, "PATCH", path, UpdateDriftRequest{Status: &status}, &drift); err != nil {
		return nil, err
	}

	return &drift, nil
}

// Delete deletes a drift
func (s *DriftService) Delete(ctx context.Context, id int64) error {
	path := fmt.Sprintf("/api/drifts/%d", id)
	return s.client.doRequest(ctx, "DELETE", path, nil, nil)
}
