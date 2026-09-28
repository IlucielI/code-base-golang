package middlewares

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"code-base-golang/internal/constants"
	"code-base-golang/internal/dtos"
)

type clientVisitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter manages per-IP token bucket rate limiting.
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*clientVisitor
	rate     rate.Limit
	burst    int
}

// NewRateLimiter creates a new thread-safe in-memory rate limiter with periodic cleanup.
func NewRateLimiter(r float64, b int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*clientVisitor),
		rate:     rate.Limit(r),
		burst:    b,
	}

	// Periodic cleanup of clients inactive for more than 5 minutes
	go func() {
		for {
			time.Sleep(2 * time.Minute)
			rl.mu.Lock()
			now := time.Now()
			for ip, v := range rl.visitors {
				if now.Sub(v.lastSeen) > 5*time.Minute {
					delete(rl.visitors, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()

	return rl
}

func (rl *RateLimiter) getVisitor(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rl.rate, rl.burst)
		rl.visitors[ip] = &clientVisitor{
			limiter:  limiter,
			lastSeen: time.Now(),
		}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

// RateLimit creates a Gin middleware that enforces per-client IP rate limits.
func RateLimit(limit float64, burst int, enabled ...bool) gin.HandlerFunc {
	isEnabled := true
	if len(enabled) > 0 {
		isEnabled = enabled[0]
	}

	if !isEnabled || limit <= 0 || burst <= 0 {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	limiter := NewRateLimiter(limit, burst)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		clientLimiter := limiter.getVisitor(ip)

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%.0f", limit))

		if !clientLimiter.Allow() {
			c.Header("Retry-After", "1")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, dtos.BaseResponse{
				Status:    constants.ResponseStatusFail,
				Code:      constants.ResponseCodeTooManyRequests,
				Message:   constants.ErrTooManyRequests.Message,
				Timestamp: time.Now().UTC(),
			})
			return
		}

		c.Next()
	}
}
