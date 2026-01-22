package http

import (
	"strconv"

	"cruise-price-compare/internal/repo"

	"github.com/gin-gonic/gin"
)

// DefaultPageSize is the default number of items per page
const DefaultPageSize = 20

// MaxPageSize is the maximum number of items per page
const MaxPageSize = 100

// PaginationParams extracts pagination parameters from query string
type PaginationParams struct {
	Page     int
	PageSize int
}

// GetPagination extracts pagination from gin context
func GetPagination(c *gin.Context) PaginationParams {
	page := 1
	pageSize := DefaultPageSize

	if p := c.Query("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			page = n
		}
	}

	if ps := c.Query("page_size"); ps != "" {
		if n, err := strconv.Atoi(ps); err == nil && n > 0 {
			pageSize = n
			if pageSize > MaxPageSize {
				pageSize = MaxPageSize
			}
		}
	}

	return PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}
}

// ToRepoPagination converts PaginationParams to repo.Pagination
func (p PaginationParams) ToRepoPagination() repo.Pagination {
	return repo.Pagination{
		Page:     p.Page,
		PageSize: p.PageSize,
	}
}

// PaginatedResponse is the standard paginated response format
type PaginatedResponse struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// NewPaginatedResponse creates a paginated response from repo result
func NewPaginatedResponse[T any](result repo.PaginatedResult[T]) PaginatedResponse {
	return PaginatedResponse{
		Items:      result.Items,
		Total:      result.Total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
	}
}

// SuccessResponse is the standard success response format
type SuccessResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// RespondOK sends a 200 OK response with data
func RespondOK(c *gin.Context, data interface{}) {
	c.JSON(200, SuccessResponse{Data: data})
}

// RespondCreated sends a 201 Created response with data
func RespondCreated(c *gin.Context, data interface{}) {
	c.JSON(201, SuccessResponse{Data: data})
}

// RespondNoContent sends a 204 No Content response
func RespondNoContent(c *gin.Context) {
	c.Status(204)
}

// RespondMessage sends a 200 OK response with message
func RespondMessage(c *gin.Context, message string) {
	c.JSON(200, SuccessResponse{Message: message})
}

// ParseUint64Param parses a uint64 path parameter
func ParseUint64Param(c *gin.Context, name string) (uint64, bool) {
	s := c.Param(name)
	if s == "" {
		return 0, false
	}

	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, false
	}

	return n, true
}

// ParseUint64Query parses a uint64 query parameter
func ParseUint64Query(c *gin.Context, name string) *uint64 {
	s := c.Query(name)
	if s == "" {
		return nil
	}

	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return nil
	}

	return &n
}
