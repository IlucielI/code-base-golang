package middlewares

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"code-base-golang/internal/config"
	"code-base-golang/internal/constants"
	"code-base-golang/internal/dtos"
)

// Middlewares is the central container for route-level and global middleware methods.
type Middlewares struct {
	cfg         config.Config
	limiter     *RateLimiter
	idempotency IdempotencyStore
}

// New initializes the central middlewares container with dependencies.
func New(cfg config.Config) *Middlewares {
	var limiter *RateLimiter
	if cfg.RateLimiterEnabled && cfg.RateLimiterLimit > 0 && cfg.RateLimiterBurst > 0 {
		limiter = NewRateLimiter(cfg.RateLimiterLimit, cfg.RateLimiterBurst)
	}
	return &Middlewares{
		cfg:     cfg,
		limiter: limiter,
	}
}

// WithIdempotencyStore fluently sets the idempotency store backend.
func (m *Middlewares) WithIdempotencyStore(store IdempotencyStore) *Middlewares {
	m.idempotency = store
	return m
}

// SetIdempotencyStore sets the idempotency store backend.
func (m *Middlewares) SetIdempotencyStore(store IdempotencyStore) {
	m.idempotency = store
}

// abortError aborts the Gin context with a standardized JSON error envelope.
func abortError(c *gin.Context, statusCode int, code, message string) {
	status := constants.ResponseStatusFail
	if statusCode >= http.StatusInternalServerError {
		status = constants.ResponseStatusError
	}
	c.AbortWithStatusJSON(statusCode, dtos.BaseResponse{
		Status:    status,
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
	})
}

// extractBearerToken parses and extracts the Bearer token string from an Authorization header.
func extractBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("missing authorization header")
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("invalid authorization header format, expected 'Bearer <token>'")
	}
	tokenStr := strings.TrimSpace(parts[1])
	if tokenStr == "" {
		return "", errors.New("empty bearer token")
	}
	return tokenStr, nil
}
