package middlewares

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"code-base-golang/internal/constants"
	"code-base-golang/internal/pkg/ctxmeta"
)

// LogEntry defines the JSON structure emitted for each HTTP access log.
type LogEntry struct {
	Level      string `json:"level"`
	Timestamp  string `json:"timestamp"`
	RequestID  string `json:"request_id"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	StatusCode int    `json:"status_code"`
	Latency    string `json:"latency"`
	LatencyMs  int64  `json:"latency_ms"`
	ClientIP   string `json:"client_ip,omitempty"`
}

type loggerOptions struct {
	writer io.Writer
}

// LoggerOption configures the StructuredLogger middleware.
type LoggerOption func(*loggerOptions)

// WithWriter overrides the output destination (default is os.Stdout).
func WithWriter(w io.Writer) LoggerOption {
	return func(o *loggerOptions) {
		if w != nil {
			o.writer = w
		}
	}
}

// Logger returns the structured logger middleware from the Middlewares container.
func (m *Middlewares) Logger(opts ...LoggerOption) gin.HandlerFunc {
	return m.StructuredLogger(opts...)
}

// StructuredLogger returns a Gin middleware that outputs structured JSON logs for each HTTP request.
func (m *Middlewares) StructuredLogger(opts ...LoggerOption) gin.HandlerFunc {
	cfg := &loggerOptions{
		writer: os.Stdout,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	return func(c *gin.Context) {
		start := time.Now()

		// 1. Extract or generate Request ID
		reqID := c.GetHeader("X-Request-ID")
		if reqID == "" {
			reqID = uuid.New().String()
		} else if _, err := uuid.Parse(reqID); err != nil {
			reqID = uuid.New().String()
		}

		// Inject request_id into response headers and contexts
		c.Header("X-Request-ID", reqID)
		c.Set("request_id", reqID)
		if c.Request != nil {
			c.Request = c.Request.WithContext(ctxmeta.WithRequestID(c.Request.Context(), reqID))
		}

		// 2. Process request
		c.Next()

		// 3. Compute latency & level
		latency := time.Since(start)
		statusCode := c.Writer.Status()

		var level string
		switch {
		case statusCode >= http.StatusInternalServerError:
			level = constants.LogLevelError
		case statusCode >= http.StatusBadRequest:
			level = constants.LogLevelWarn
		default:
			level = constants.LogLevelInfo
		}

		entry := LogEntry{
			Level:      level,
			Timestamp:  time.Now().UTC().Format(time.RFC3339),
			RequestID:  reqID,
			Method:     c.Request.Method,
			Path:       c.Request.URL.Path,
			StatusCode: statusCode,
			Latency:    latency.String(),
			LatencyMs:  latency.Milliseconds(),
			ClientIP:   c.ClientIP(),
		}

		data, err := json.Marshal(entry)
		if err == nil {
			data = append(data, '\n')
			if n, writeErr := cfg.writer.Write(data); writeErr != nil {
				log.Printf("failed to write structured log (bytes: %d): %v", n, writeErr)
			}
		}
	}
}
