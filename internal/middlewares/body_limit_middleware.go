package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	// DefaultMaxRequestBodySize is 2MB, providing generous headroom for rich JSON payloads
	// while preventing memory exhaustion DoS attacks.
	DefaultMaxRequestBodySize = 2 * 1024 * 1024
)

// BodyLimit returns the request body limit middleware from the Middlewares container,
// falling back to m.cfg.MaxRequestBodySize or DefaultMaxRequestBodySize if no explicit limit is provided.
func (m *Middlewares) BodyLimit(maxBytes ...int64) gin.HandlerFunc {
	limit := m.cfg.MaxRequestBodySize
	if len(maxBytes) > 0 && maxBytes[0] > 0 {
		limit = maxBytes[0]
	}
	if limit <= 0 {
		limit = DefaultMaxRequestBodySize
	}

	return func(c *gin.Context) {
		if c.Request != nil && c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		}
		c.Next()
	}
}
