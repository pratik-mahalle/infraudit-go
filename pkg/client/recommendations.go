package client

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// RecommendationService handles recommendation-related API calls
type RecommendationService struct {
	client *Client
}

// RecommendationListOptions contains options for listing recommendations
type RecommendationListOptions struct {
	ListOptions
	Type       *string `json:"type,omitempty"`   // cost, performance, security
	Impact     *string `json:"impact,omitempty"` // high, medium, low
	Status     *string `json:"status,omitempty"` // pending, applied, dismissed
	ResourceID *int64  `json:"resource_id,omitempty"`
}

// UpdateRecommendationRequest represents a request to update a recommendation
type UpdateRecommendationRequest struct {
	Status *string `json:"status,omitempty"` // pending, applied, dismissed
}

// List retrieves a list of recommendations
func (s *RecommendationService) List(ctx context.Context, opts *RecommendationListOptions) ([]Recommendation, error) {
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
		if opts.Type != nil {
			query.Set("type", *opts.Type)
		}
		if opts.Impact != nil {
			query.Set("impact", *opts.Impact)
		}
		if opts.Status != nil {
			query.Set("status", *opts.Status)
		}
		if opts.ResourceID != nil {
			query.Set("resource_id", strconv.FormatInt(*opts.ResourceID, 10))
		}
	}

	path := "/api/recommendations"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	var recommendations []Recommendation
	if err := s.client.doRequest(ctx, "GET", path, nil, &recommendations); err != nil {
		return nil, err
	}

	return recommendations, nil
}

// Get retrieves a single recommendation by ID
func (s *RecommendationService) Get(ctx context.Context, id int64) (*Recommendation, error) {
	path := fmt.Sprintf("/api/recommendations/%d", id)

	var recommendation Recommendation
	if err := s.client.doRequest(ctx, "GET", path, nil, &recommendation); err != nil {
		return nil, err
	}

	return &recommendation, nil
}

// Apply marks a recommendation as applied
func (s *RecommendationService) Apply(ctx context.Context, id int64) (*Recommendation, error) {
	path := fmt.Sprintf("/api/recommendations/%d", id)
	status := "applied"

	var recommendation Recommendation
	if err := s.client.doRequest(ctx, "PATCH", path, UpdateRecommendationRequest{Status: &status}, &recommendation); err != nil {
		return nil, err
	}

	return &recommendation, nil
}

// Dismiss marks a recommendation as dismissed
func (s *RecommendationService) Dismiss(ctx context.Context, id int64) (*Recommendation, error) {
	path := fmt.Sprintf("/api/recommendations/%d", id)
	status := "dismissed"

	var recommendation Recommendation
	if err := s.client.doRequest(ctx, "PATCH", path, UpdateRecommendationRequest{Status: &status}, &recommendation); err != nil {
		return nil, err
	}

	return &recommendation, nil
}

// Delete deletes a recommendation
func (s *RecommendationService) Delete(ctx context.Context, id int64) error {
	path := fmt.Sprintf("/api/recommendations/%d", id)
	return s.client.doRequest(ctx, "DELETE", path, nil, nil)
}
