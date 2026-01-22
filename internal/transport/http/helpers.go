package http

import (
	"cruise-price-compare/internal/repo"

	"github.com/gin-gonic/gin"
)

// ParsePagination parses pagination parameters from query string
func ParsePagination(c *gin.Context) repo.Pagination {
	params := GetPagination(c)
	return params.ToRepoPagination()
}

// RespondError sends a JSON error response
func RespondError(c *gin.Context, statusCode int, errorCode, message string) {
	c.JSON(statusCode, gin.H{
		"error":   errorCode,
		"message": message,
	})
}
