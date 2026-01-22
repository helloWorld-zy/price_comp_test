package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Error codes
const (
	ErrCodeInternal        = "INTERNAL_ERROR"
	ErrCodeValidation      = "VALIDATION_ERROR"
	ErrCodeNotFound        = "NOT_FOUND"
	ErrCodeUnauthorized    = "UNAUTHORIZED"
	ErrCodeForbidden       = "FORBIDDEN"
	ErrCodeBadRequest      = "BAD_REQUEST"
	ErrCodeConflict        = "CONFLICT"
	ErrCodeTooManyRequests = "TOO_MANY_REQUESTS"
)

// APIError represents an API error response
type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// ErrorResponse wraps an API error
type ErrorResponse struct {
	Error APIError `json:"error"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(code, message string, details interface{}) ErrorResponse {
	return ErrorResponse{
		Error: APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

// RespondWithError sends an error response
func RespondWithError(c *gin.Context, status int, code, message string) {
	c.JSON(status, NewErrorResponse(code, message, nil))
}

// RespondWithErrorDetails sends an error response with details
func RespondWithErrorDetails(c *gin.Context, status int, code, message string, details interface{}) {
	c.JSON(status, NewErrorResponse(code, message, details))
}

// RespondInternalError sends a 500 internal error
func RespondInternalError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, NewErrorResponse(ErrCodeInternal, "internal server error", nil))
}

// RespondValidationError sends a 400 validation error
func RespondValidationError(c *gin.Context, message string, details interface{}) {
	c.JSON(http.StatusBadRequest, NewErrorResponse(ErrCodeValidation, message, details))
}

// RespondNotFound sends a 404 not found error
func RespondNotFound(c *gin.Context, resource string) {
	c.JSON(http.StatusNotFound, NewErrorResponse(ErrCodeNotFound, resource+" not found", nil))
}

// RespondUnauthorized sends a 401 unauthorized error
func RespondUnauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, NewErrorResponse(ErrCodeUnauthorized, message, nil))
}

// RespondForbidden sends a 403 forbidden error
func RespondForbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, NewErrorResponse(ErrCodeForbidden, message, nil))
}

// RespondBadRequest sends a 400 bad request error
func RespondBadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, NewErrorResponse(ErrCodeBadRequest, message, nil))
}

// RespondConflict sends a 409 conflict error
func RespondConflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, NewErrorResponse(ErrCodeConflict, message, nil))
}

// ErrorHandler is a gin middleware that handles errors
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there were any errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			// Handle specific error types
			var status int
			var code string
			var message string

			switch {
			case errors.Is(err.Err, ErrNotFound):
				status = http.StatusNotFound
				code = ErrCodeNotFound
				message = "resource not found"
			case errors.Is(err.Err, ErrValidation):
				status = http.StatusBadRequest
				code = ErrCodeValidation
				message = err.Error()
			case errors.Is(err.Err, ErrUnauthorized):
				status = http.StatusUnauthorized
				code = ErrCodeUnauthorized
				message = "unauthorized"
			case errors.Is(err.Err, ErrForbidden):
				status = http.StatusForbidden
				code = ErrCodeForbidden
				message = "forbidden"
			default:
				status = http.StatusInternalServerError
				code = ErrCodeInternal
				message = "internal server error"
			}

			c.JSON(status, NewErrorResponse(code, message, nil))
		}
	}
}

// Sentinel errors for error handling
var (
	ErrNotFound     = errors.New("not found")
	ErrValidation   = errors.New("validation error")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
)
