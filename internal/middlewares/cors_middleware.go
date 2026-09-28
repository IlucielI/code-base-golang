package middlewares

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CORSConfig defines configuration options for the CORS middleware.
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           time.Duration
}

// DefaultCORSConfig returns standard permissive default options.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID", "Idempotency-Key"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
}

// CORS returns a Gin middleware handler implementing Cross-Origin Resource Sharing.
func CORS(cfg ...CORSConfig) gin.HandlerFunc {
	config := DefaultCORSConfig()
	if len(cfg) > 0 {
		config = cfg[0]
	}

	methods := strings.Join(config.AllowedMethods, ", ")
	headers := strings.Join(config.AllowedHeaders, ", ")
	exposeHeaders := strings.Join(config.ExposeHeaders, ", ")
	maxAge := fmt.Sprintf("%d", int(config.MaxAge.Seconds()))

	allowAllOrigins := false
	originSet := make(map[string]struct{}, len(config.AllowedOrigins))
	for _, o := range config.AllowedOrigins {
		if o == "*" {
			allowAllOrigins = true
		}
		originSet[strings.ToLower(strings.TrimSpace(o))] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		if origin != "" {
			matched := false
			if allowAllOrigins {
				c.Header("Access-Control-Allow-Origin", origin)
				matched = true
			} else if _, ok := originSet[strings.ToLower(origin)]; ok {
				c.Header("Access-Control-Allow-Origin", origin)
				matched = true
			}

			if matched {
				c.Header("Vary", "Origin")

				if config.AllowCredentials {
					c.Header("Access-Control-Allow-Credentials", "true")
				}
				if len(config.AllowedMethods) > 0 {
					c.Header("Access-Control-Allow-Methods", methods)
				}
				if len(config.AllowedHeaders) > 0 {
					c.Header("Access-Control-Allow-Headers", headers)
				}
				if len(config.ExposeHeaders) > 0 {
					c.Header("Access-Control-Expose-Headers", exposeHeaders)
				}
				if config.MaxAge > 0 {
					c.Header("Access-Control-Max-Age", maxAge)
				}
			}
		}

		// Handle preflight OPTIONS request directly with 204 No Content
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
