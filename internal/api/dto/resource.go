package dto

import "time"

// ResourceDTO represents a cloud resource in API responses
// Uses camelCase for frontend compatibility
type ResourceDTO struct {
	ID            int64             `json:"id"`
	ResourceID    string            `json:"resourceId,omitempty"`
	Provider      string            `json:"provider"`
	Name          string            `json:"name"`
	Type          string            `json:"type"`
	Region        string            `json:"region"`
	Status        string            `json:"status"`
	Tags          map[string]string `json:"tags,omitempty"`
	Cost          float64           `json:"cost"`
	Configuration string            `json:"configuration,omitempty"`
	CreatedAt     time.Time         `json:"createdAt"`
	UpdatedAt     time.Time         `json:"updatedAt,omitempty"`
}

// CreateResourceRequest represents a resource creation request
type CreateResourceRequest struct {
	Provider      string            `json:"provider" validate:"required,oneof=aws gcp azure"`
	ResourceID    string            `json:"resourceId" validate:"required"`
	Name          string            `json:"name" validate:"required"`
	Type          string            `json:"type" validate:"required"`
	Region        string            `json:"region" validate:"required"`
	Status        string            `json:"status,omitempty"`
	Tags          map[string]string `json:"tags,omitempty"`
	Cost          float64           `json:"cost,omitempty"`
	Configuration string            `json:"configuration,omitempty"`
}

// UpdateResourceRequest represents a resource update request
type UpdateResourceRequest struct {
	Name          *string            `json:"name,omitempty"`
	Type          *string            `json:"type,omitempty"`
	Region        *string            `json:"region,omitempty"`
	Status        *string            `json:"status,omitempty"`
	Tags          *map[string]string `json:"tags,omitempty"`
	Cost          *float64           `json:"cost,omitempty"`
	Configuration *string            `json:"configuration,omitempty"`
}

// ResourceListRequest represents resource list query parameters
type ResourceListRequest struct {
	Provider string `json:"provider,omitempty"`
	Type     string `json:"type,omitempty"`
	Region   string `json:"region,omitempty"`
	Status   string `json:"status,omitempty"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"pageSize,omitempty"`
}

// ResourceSummaryDTO represents resource summary statistics
type ResourceSummaryDTO struct {
	Total      int            `json:"total"`
	ByProvider map[string]int `json:"byProvider"`
	ByType     map[string]int `json:"byType"`
	ByStatus   map[string]int `json:"byStatus"`
	TotalCost  float64        `json:"totalCost"`
}
