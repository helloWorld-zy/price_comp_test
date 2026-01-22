package obs

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// TraceIDHeader is the HTTP header name for trace ID
	TraceIDHeader = "X-Trace-ID"
	// TraceIDKey is the gin context key for trace ID
	TraceIDKey = "trace_id"
)

// TraceMiddleware adds trace ID to each request
func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get trace ID from header or generate new one
		traceID := c.GetHeader(TraceIDHeader)
		if traceID == "" {
			traceID = uuid.New().String()
		}

		// Store in gin context
		c.Set(TraceIDKey, traceID)

		// Set response header
		c.Header(TraceIDHeader, traceID)

		// Add to request context
		ctx := WithTraceID(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// GetTraceID retrieves trace ID from gin context
func GetTraceID(c *gin.Context) string {
	if traceID, exists := c.Get(TraceIDKey); exists {
		if s, ok := traceID.(string); ok {
			return s
		}
	}
	return ""
}
