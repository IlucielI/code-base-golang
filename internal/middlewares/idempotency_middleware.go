package middlewares

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"code-base-golang/internal/constants"
)

// IdempotencyStore defines the storage contract required by the Idempotency middleware.
type IdempotencyStore interface {
	SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error)
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	Delete(ctx context.Context, keys ...string) error
}

// CachedResponse stores HTTP response status, headers, and body for idempotency replay.
type CachedResponse struct {
	StatusCode  int                 `json:"status_code"`
	ContentType string              `json:"content_type"`
	Headers     map[string][]string `json:"headers"`
	Body        []byte              `json:"body"`
}

type responseInterceptor struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (w *responseInterceptor) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseInterceptor) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func (w *responseInterceptor) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseInterceptor) Status() int {
	if w.statusCode != 0 {
		return w.statusCode
	}
	return w.ResponseWriter.Status()
}

// Idempotency returns a Gin middleware that ensures mutating requests with an Idempotency-Key
// header are executed only once and duplicate requests replay the cached response.
func (m *Middlewares) Idempotency() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Extract idempotency key from headers (support both standard and X- prefixed)
		key := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
		if key == "" {
			key = strings.TrimSpace(c.GetHeader("X-Idempotency-Key"))
		}

		// Header is optional; if omitted, proceed normally
		if key == "" || m.idempotency == nil {
			c.Next()
			return
		}

		ctx := c.Request.Context()
		lockKey := fmt.Sprintf("idempotency:lock:%s", key)
		respKey := fmt.Sprintf("idempotency:response:%s", key)

		// 2. Check if cached response already exists
		cachedJSON, err := m.idempotency.Get(ctx, respKey)
		if err == nil && cachedJSON != "" {
			var cached CachedResponse
			if unmarshalErr := json.Unmarshal([]byte(cachedJSON), &cached); unmarshalErr == nil {
				for hKey, hVals := range cached.Headers {
					for _, v := range hVals {
						c.Writer.Header().Add(hKey, v)
					}
				}
				c.Writer.Header().Set("X-Cache-Lookup", "HIT")
				c.Data(cached.StatusCode, cached.ContentType, cached.Body)
				c.Abort()
				return
			}
		}

		// 3. Acquire in-flight lock to prevent concurrent duplicate execution
		lockTTL := 30 * time.Second
		acquired, err := m.idempotency.SetNX(ctx, lockKey, "in_progress", lockTTL)
		if err != nil || !acquired {
			abortError(c, http.StatusConflict, constants.ResponseCodeConflict, "request with this idempotency key is currently processing")
			return
		}

		// Release in-flight lock when processing finishes
		defer func() {
			_ = m.idempotency.Delete(context.Background(), lockKey)
		}()

		// 4. Wrap response writer to capture status, headers, and body
		interceptor := &responseInterceptor{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = interceptor

		c.Next()

		// 5. Store response in cache for successful or client error responses (< 500)
		statusCode := interceptor.Status()
		if statusCode < http.StatusInternalServerError {
			ttl := m.cfg.IdempotencyTTL
			if ttl <= 0 {
				ttl = 24 * time.Hour
			}

			cached := CachedResponse{
				StatusCode:  statusCode,
				ContentType: c.Writer.Header().Get("Content-Type"),
				Headers:     c.Writer.Header().Clone(),
				Body:        interceptor.body.Bytes(),
			}

			if data, marshalErr := json.Marshal(cached); marshalErr == nil {
				_ = m.idempotency.Set(context.Background(), respKey, string(data), ttl)
			}
		}
	}
}
