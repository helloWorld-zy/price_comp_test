package obs

import (
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Metrics holds application metrics
type Metrics struct {
	mu sync.RWMutex

	// Request metrics
	TotalRequests    int64
	TotalErrors      int64
	RequestDurations []time.Duration

	// Business metrics
	TotalQuotes       int64
	TotalImports      int64
	SuccessfulImports int64
	FailedImports     int64

	// Endpoint metrics
	EndpointCounts map[string]int64
}

// NewMetrics creates a new metrics collector
func NewMetrics() *Metrics {
	return &Metrics{
		EndpointCounts:   make(map[string]int64),
		RequestDurations: make([]time.Duration, 0, 1000),
	}
}

// RecordRequest records a request metric
func (m *Metrics) RecordRequest(path string, duration time.Duration, status int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalRequests++
	if status >= 400 {
		m.TotalErrors++
	}

	// Keep last 1000 durations for percentile calculation
	if len(m.RequestDurations) >= 1000 {
		m.RequestDurations = m.RequestDurations[1:]
	}
	m.RequestDurations = append(m.RequestDurations, duration)

	key := path
	m.EndpointCounts[key]++
}

// RecordQuote records a quote creation
func (m *Metrics) RecordQuote() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalQuotes++
}

// RecordImport records an import
func (m *Metrics) RecordImport(success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalImports++
	if success {
		m.SuccessfulImports++
	} else {
		m.FailedImports++
	}
}

// GetStats returns current metrics stats
func (m *Metrics) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := map[string]interface{}{
		"total_requests":     m.TotalRequests,
		"total_errors":       m.TotalErrors,
		"total_quotes":       m.TotalQuotes,
		"total_imports":      m.TotalImports,
		"successful_imports": m.SuccessfulImports,
		"failed_imports":     m.FailedImports,
		"error_rate":         float64(0),
	}

	if m.TotalRequests > 0 {
		stats["error_rate"] = float64(m.TotalErrors) / float64(m.TotalRequests)
	}

	// Calculate average duration
	if len(m.RequestDurations) > 0 {
		var total time.Duration
		for _, d := range m.RequestDurations {
			total += d
		}
		stats["avg_request_duration_ms"] = float64(total.Milliseconds()) / float64(len(m.RequestDurations))
	}

	return stats
}

// MetricsMiddleware collects request metrics
func MetricsMiddleware(metrics *Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		metrics.RecordRequest(c.FullPath(), duration, c.Writer.Status())
	}
}

// Default metrics instance
var defaultMetrics = NewMetrics()

// DefaultMetrics returns the default metrics instance
func DefaultMetrics() *Metrics {
	return defaultMetrics
}

// RequestLoggerMiddleware logs each request
func RequestLoggerMiddleware(logger *Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		logger.WithContext(c.Request.Context()).Info("http request",
			"method", c.Request.Method,
			"path", path,
			"query", query,
			"status", status,
			"latency_ms", latency.Milliseconds(),
			"client_ip", c.ClientIP(),
			"user_agent", c.GetHeader("User-Agent"),
			"body_size", strconv.Itoa(c.Writer.Size()),
		)
	}
}
