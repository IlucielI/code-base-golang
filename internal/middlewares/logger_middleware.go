package middlewares

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"code-base-golang/internal/pkg/ctxmeta"
	"code-base-golang/internal/pkg/logger"
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

// StructuredLogger returns a Gin middleware that outputs structured logs for each HTTP request using slog.
func (m *Middlewares) StructuredLogger(opts ...LoggerOption) gin.HandlerFunc {
	cfg := &loggerOptions{
		writer: os.Stdout,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	appLogger := logger.New(m.cfg, logger.WithWriter(cfg.writer), logger.WithLevel(slog.LevelDebug))

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

		slogLevel := slog.LevelInfo
		switch {
		case statusCode >= http.StatusInternalServerError:
			slogLevel = slog.LevelError
		case statusCode >= http.StatusBadRequest:
			slogLevel = slog.LevelWarn
		}

		appLogger.Log(c.Request.Context(), slogLevel, "HTTP request completed",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status_code", statusCode),
			slog.String("latency", latency.String()),
			slog.Int64("latency_ms", latency.Milliseconds()),
			slog.String("client_ip", c.ClientIP()),
		)
	}
}
