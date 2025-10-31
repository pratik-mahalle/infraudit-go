package client

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// AnomalyService handles anomaly detection-related API calls
type AnomalyService struct {
	client *Client
}

// AnomalyListOptions contains options for listing anomalies
type AnomalyListOptions struct {
	ListOptions
	AnomalyType *string `json:"anomaly_type,omitempty"` // cost_spike, unusual_usage
	Severity    *string `json:"severity,omitempty"`     // critical, high, medium, low
	Status      *string `json:"status,omitempty"`       // detected, investigating, resolved
	ResourceID  *int64  `json:"resource_id,omitempty"`
}

// UpdateAnomalyRequest represents a request to update an anomaly
type UpdateAnomalyRequest struct {
	Status *string `json:"status,omitempty"` // detected, investigating, resolved
}

// List retrieves a list of anomalies
func (s *AnomalyService) List(ctx context.Context, opts *AnomalyListOptions) ([]Anomaly, error) {
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
		if opts.AnomalyType != nil {
			query.Set("anomaly_type", *opts.AnomalyType)
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

	path := "/api/anomalies"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	var anomalies []Anomaly
	if err := s.client.doRequest(ctx, "GET", path, nil, &anomalies); err != nil {
		return nil, err
	}

	return anomalies, nil
}

// Get retrieves a single anomaly by ID
func (s *AnomalyService) Get(ctx context.Context, id int64) (*Anomaly, error) {
	path := fmt.Sprintf("/api/anomalies/%d", id)

	var anomaly Anomaly
	if err := s.client.doRequest(ctx, "GET", path, nil, &anomaly); err != nil {
		return nil, err
	}

	return &anomaly, nil
}

// Resolve marks an anomaly as resolved
func (s *AnomalyService) Resolve(ctx context.Context, id int64) (*Anomaly, error) {
	path := fmt.Sprintf("/api/anomalies/%d", id)
	status := "resolved"

	var anomaly Anomaly
	if err := s.client.doRequest(ctx, "PATCH", path, UpdateAnomalyRequest{Status: &status}, &anomaly); err != nil {
		return nil, err
	}

	return &anomaly, nil
}

// Investigate marks an anomaly as under investigation
func (s *AnomalyService) Investigate(ctx context.Context, id int64) (*Anomaly, error) {
	path := fmt.Sprintf("/api/anomalies/%d", id)
	status := "investigating"

	var anomaly Anomaly
	if err := s.client.doRequest(ctx, "PATCH", path, UpdateAnomalyRequest{Status: &status}, &anomaly); err != nil {
		return nil, err
	}

	return &anomaly, nil
}

// Delete deletes an anomaly
func (s *AnomalyService) Delete(ctx context.Context, id int64) error {
	path := fmt.Sprintf("/api/anomalies/%d", id)
	return s.client.doRequest(ctx, "DELETE", path, nil, nil)
}
