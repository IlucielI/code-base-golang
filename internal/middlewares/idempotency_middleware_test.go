package middlewares

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"code-base-golang/internal/config"
)

// mockIdempotencyStore provides an in-memory thread-safe store for testing.
type mockIdempotencyStore struct {
	mu   sync.Mutex
	data map[string]string
}

func newMockIdempotencyStore() *mockIdempotencyStore {
	return &mockIdempotencyStore{
		data: make(map[string]string),
	}
}

func (s *mockIdempotencyStore) SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.data[key]; exists {
		return false, nil
	}
	s.data[key] = value.(string)
	return true, nil
}

func (s *mockIdempotencyStore) Get(ctx context.Context, key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	val, exists := s.data[key]
	if !exists {
		return "", errors.New("key not found")
	}
	return val, nil
}

func (s *mockIdempotencyStore) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value.(string)
	return nil
}

func (s *mockIdempotencyStore) Delete(ctx context.Context, keys ...string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, k := range keys {
		delete(s.data, k)
	}
	return nil
}

func TestIdempotencyMiddleware(t *testing.T) {
	store := newMockIdempotencyStore()
	cfg := config.Config{IdempotencyTTL: 1 * time.Hour}
	mws := New(cfg).WithIdempotencyStore(store)

	t.Run("omitted_header_proceeds_normally", func(t *testing.T) {
		router := gin.New()
		router.Use(mws.Idempotency())

		var executionCount int32
		router.POST("/orders", func(c *gin.Context) {
			atomic.AddInt32(&executionCount, 1)
			c.JSON(http.StatusCreated, gin.H{"order_id": "123"})
		})

		req := httptest.NewRequest(http.MethodPost, "/orders", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created, got %d", w.Code)
		}
		if atomic.LoadInt32(&executionCount) != 1 {
			t.Errorf("expected 1 execution, got %d", executionCount)
		}
	})

	t.Run("caches_and_replays_response_for_same_key", func(t *testing.T) {
		router := gin.New()
		router.Use(mws.Idempotency())

		var executionCount int32
		router.POST("/payments", func(c *gin.Context) {
			count := atomic.AddInt32(&executionCount, 1)
			c.Header("X-Custom-Tx", "tx-999")
			c.JSON(http.StatusOK, gin.H{"status": "paid", "exec": count})
		})

		idemKey := "key-uuid-unique-123"

		// First request: should execute handler
		req1 := httptest.NewRequest(http.MethodPost, "/payments", nil)
		req1.Header.Set("Idempotency-Key", idemKey)
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)

		if w1.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w1.Code)
		}
		if atomic.LoadInt32(&executionCount) != 1 {
			t.Fatalf("expected handler executed once, got %d", executionCount)
		}
		if w1.Header().Get("X-Cache-Lookup") == "HIT" {
			t.Error("expected first request not to be a cache HIT")
		}

		// Second request with same key: should replay cached response without re-executing handler
		req2 := httptest.NewRequest(http.MethodPost, "/payments", nil)
		req2.Header.Set("X-Idempotency-Key", idemKey) // test X- prefix fallback too
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		if w2.Code != http.StatusOK {
			t.Fatalf("expected 200 on replay, got %d", w2.Code)
		}
		if atomic.LoadInt32(&executionCount) != 1 {
			t.Errorf("expected handler execution count to stay 1, got %d", executionCount)
		}
		if w2.Header().Get("X-Cache-Lookup") != "HIT" {
			t.Error("expected X-Cache-Lookup: HIT on replayed request")
		}
		if w2.Header().Get("X-Custom-Tx") != "tx-999" {
			t.Errorf("expected replayed custom header, got %s", w2.Header().Get("X-Custom-Tx"))
		}
		if w2.Body.String() != w1.Body.String() {
			t.Errorf("expected identical body: %s vs %s", w1.Body.String(), w2.Body.String())
		}
	})

	t.Run("concurrent_request_returns_conflict_409", func(t *testing.T) {
		router := gin.New()
		router.Use(mws.Idempotency())

		router.POST("/transfer", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true})
		})

		key := "in-flight-key-456"
		// Simulate lock already acquired
		store.data["idempotency:lock:"+key] = "in_progress"

		req := httptest.NewRequest(http.MethodPost, "/transfer", nil)
		req.Header.Set("Idempotency-Key", key)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Fatalf("expected 409 Conflict, got %d", w.Code)
		}
	})

	t.Run("does_not_cache_5xx_errors", func(t *testing.T) {
		router := gin.New()
		router.Use(mws.Idempotency())

		var executionCount int32
		router.POST("/fail", func(c *gin.Context) {
			atomic.AddInt32(&executionCount, 1)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "panic"})
		})

		key := "error-key-789"
		req1 := httptest.NewRequest(http.MethodPost, "/fail", nil)
		req1.Header.Set("Idempotency-Key", key)
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)

		if w1.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w1.Code)
		}

		// Subsequent request should not replay 500 from cache; handler should run again
		req2 := httptest.NewRequest(http.MethodPost, "/fail", nil)
		req2.Header.Set("Idempotency-Key", key)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		if atomic.LoadInt32(&executionCount) != 2 {
			t.Errorf("expected handler executed twice for 5xx, got %d", executionCount)
		}
	})
}
