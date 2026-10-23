package middlewares

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"

	"code-base-golang/internal/constants"
	"code-base-golang/internal/pkg/ctxmeta"
	"code-base-golang/internal/pkg/response"
)

type recoveryOptions struct {
	writer io.Writer
}

// RecoveryOption configures the Recovery middleware.
type RecoveryOption func(*recoveryOptions)

// WithRecoveryWriter overrides the panic stack trace log destination (default is os.Stderr).
func WithRecoveryWriter(w io.Writer) RecoveryOption {
	return func(o *recoveryOptions) {
		if w != nil {
			o.writer = w
		}
	}
}

// Recovery returns a Gin middleware that recovers from any panics, logs the stack trace server-side,
// and returns a unified HTTP 500 JSON error envelope without exposing internal details.
func (m *Middlewares) Recovery(opts ...RecoveryOption) gin.HandlerFunc {
	cfg := &recoveryOptions{
		writer: os.Stderr,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// 1. Get request context metadata
				reqID := ""
				if c.Request != nil {
					reqID = ctxmeta.GetRequestID(c.Request.Context())
				}
				if reqID == "" {
					if id, exists := c.Get("request_id"); exists {
						if idStr, ok := id.(string); ok {
							reqID = idStr
						}
					}
				}
				if reqID == "" {
					reqID = c.Writer.Header().Get("X-Request-ID")
				}

				now := time.Now().UTC()
				timestamp := now.Format(time.RFC3339)
				method := ""
				path := ""
				if c.Request != nil {
					method = c.Request.Method
					path = c.Request.URL.Path
				}

				stack := debug.Stack()
				logMsg := fmt.Sprintf("[PANIC RECOVERED] time=%s request_id=%s method=%s path=%s error=%v\nstack:\n%s\n",
					timestamp, reqID, method, path, r, string(stack))
				_, _ = fmt.Fprint(cfg.writer, logMsg)

				// 3. Respond with standardized unified JSON error envelope
				c.AbortWithStatusJSON(http.StatusInternalServerError, response.BaseResponse{
					Status:    constants.ResponseStatusError,
					Code:      constants.ResponseCodeInternalError,
					Message:   constants.ErrInternalServerError.Message,
					Timestamp: now,
				})
			}
		}()

		c.Next()
	}
}
