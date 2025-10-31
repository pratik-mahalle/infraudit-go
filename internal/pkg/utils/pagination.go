package utils

import (
	"strconv"

	"github.com/go-chi/chi/v5"
)

// PaginationParams contains pagination parameters
type PaginationParams struct {
	Page     int
	PageSize int
	Offset   int
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalItems int64       `json:"total_items"`
	TotalPages int         `json:"total_pages"`
}

// DefaultPageSize is the default number of items per page
const DefaultPageSize = 20

// MaxPageSize is the maximum number of items per page
const MaxPageSize = 100

// ParsePaginationParams parses pagination parameters from query string
func ParsePaginationParams(r *chi.Context) PaginationParams {
	page := parseIntQuery(r.URLParam("page"), 1)
	pageSize := parseIntQuery(r.URLParam("page_size"), DefaultPageSize)

	// Enforce limits
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	offset := (page - 1) * pageSize

	return PaginationParams{
		Page:     page,
		PageSize: pageSize,
		Offset:   offset,
	}
}

// NewPaginatedResponse creates a new paginated response
func NewPaginatedResponse(data interface{}, page, pageSize int, totalItems int64) PaginatedResponse {
	totalPages := int(totalItems) / pageSize
	if int(totalItems)%pageSize != 0 {
		totalPages++
	}

	return PaginatedResponse{
		Data:       data,
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}

func parseIntQuery(value string, defaultValue int) int {
	if value == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return i
}
