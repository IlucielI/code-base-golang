package middlewares

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"code-base-golang/internal/pkg/metrics"
)

// Metrics returns a Gin middleware that observes HTTP request counts, latencies,
// and in-flight request gauges for Prometheus scraping.
func (m *Middlewares) Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request != nil && c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()
		metrics.IncInFlight()
		defer metrics.DecInFlight()

		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		path := c.FullPath()
		if path == "" {
			path = "unmatched"
		}

		method := "UNKNOWN"
		if c.Request != nil {
			method = c.Request.Method
		}

		metrics.ObserveHTTPRequest(method, path, status, duration)
	}
}
