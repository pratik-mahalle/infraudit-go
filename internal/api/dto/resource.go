package dto

// ResourceDTO represents a cloud resource in API responses
type ResourceDTO struct {
	ID         string `json:"id"`
	Provider   string `json:"provider"`
	ResourceID string `json:"resource_id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Region     string `json:"region"`
	Status     string `json:"status"`
}

// CreateResourceRequest represents a resource creation request
type CreateResourceRequest struct {
	Provider   string `json:"provider" validate:"required,oneof=aws gcp azure"`
	ResourceID string `json:"resource_id" validate:"required"`
	Name       string `json:"name" validate:"required"`
	Type       string `json:"type" validate:"required"`
	Region     string `json:"region" validate:"required"`
	Status     string `json:"status,omitempty"`
}

// UpdateResourceRequest represents a resource update request
type UpdateResourceRequest struct {
	Name   *string `json:"name,omitempty"`
	Type   *string `json:"type,omitempty"`
	Region *string `json:"region,omitempty"`
	Status *string `json:"status,omitempty"`
}

// ResourceListRequest represents resource list query parameters
type ResourceListRequest struct {
	Provider string `json:"provider,omitempty"`
	Type     string `json:"type,omitempty"`
	Region   string `json:"region,omitempty"`
	Status   string `json:"status,omitempty"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
}
