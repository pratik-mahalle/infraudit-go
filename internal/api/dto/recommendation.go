package dto

import "time"

// RecommendationDTO represents a recommendation in API responses
type RecommendationDTO struct {
	ID          int64     `json:"id"`
	Type        string    `json:"type"`
	Priority    string    `json:"priority"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Savings     float64   `json:"savings,omitempty"`
	Effort      string    `json:"effort"`
	Impact      string    `json:"impact"`
	Category    string    `json:"category"`
	Resources   []string  `json:"resources,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

// CreateRecommendationRequest represents a recommendation creation request
type CreateRecommendationRequest struct {
	Type        string   `json:"type" validate:"required"`
	Priority    string   `json:"priority" validate:"required,oneof=critical high medium low"`
	Title       string   `json:"title" validate:"required"`
	Description string   `json:"description" validate:"required"`
	Savings     float64  `json:"savings,omitempty"`
	Effort      string   `json:"effort" validate:"required,oneof=low medium high"`
	Impact      string   `json:"impact" validate:"required,oneof=high medium low"`
	Category    string   `json:"category" validate:"required"`
	Resources   []string `json:"resources,omitempty"`
}

// UpdateRecommendationRequest represents a recommendation update request
type UpdateRecommendationRequest struct {
	Type        *string   `json:"type,omitempty"`
	Priority    *string   `json:"priority,omitempty"`
	Title       *string   `json:"title,omitempty"`
	Description *string   `json:"description,omitempty"`
	Savings     *float64  `json:"savings,omitempty"`
	Effort      *string   `json:"effort,omitempty"`
	Impact      *string   `json:"impact,omitempty"`
	Category    *string   `json:"category,omitempty"`
	Resources   *[]string `json:"resources,omitempty"`
}

// RecommendationListRequest represents recommendation list query parameters
type RecommendationListRequest struct {
	Type     string `json:"type,omitempty"`
	Priority string `json:"priority,omitempty"`
	Category string `json:"category,omitempty"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"pageSize,omitempty"`
}
