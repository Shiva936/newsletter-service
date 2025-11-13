package dtos

import (
	"newsletter-service/internal/constants"
)

// PaginationRequest represents pagination parameters
type PaginationRequest struct {
	Page     int `form:"page" json:"page" binding:"omitempty,min=1"`                   // Page number (starts from 1)
	PageSize int `form:"page_size" json:"page_size" binding:"omitempty,min=1,max=100"` // Items per page (max 100)
}

// PaginationResponse represents pagination metadata
type PaginationResponse struct {
	Page       int   `json:"page"`        // Current page number
	PageSize   int   `json:"page_size"`   // Items per page
	TotalItems int64 `json:"total_items"` // Total number of items
	TotalPages int   `json:"total_pages"` // Total number of pages
	HasNext    bool  `json:"has_next"`    // Whether there's a next page
	HasPrev    bool  `json:"has_prev"`    // Whether there's a previous page
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse[T any] struct {
	Data       []T                `json:"data"`       // The actual data
	Pagination PaginationResponse `json:"pagination"` // Pagination metadata
}

// GetDefaults returns default pagination values
func (p *PaginationRequest) GetDefaults() (page int, pageSize int) {
	page = constants.DefaultPage
	pageSize = constants.DefaultPageSize // Default page size

	if p.Page > 0 {
		page = p.Page
	}
	if p.PageSize > 0 {
		pageSize = p.PageSize
	}

	return page, pageSize
}

// CalculateOffset calculates the offset for database queries
func (p *PaginationRequest) CalculateOffset() int {
	page, pageSize := p.GetDefaults()
	return (page - 1) * pageSize
}

// CreatePaginationResponse creates pagination metadata
func CreatePaginationResponse(page, pageSize int, totalItems int64) PaginationResponse {
	totalPages := int((totalItems + int64(pageSize) - 1) / int64(pageSize))

	return PaginationResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}
