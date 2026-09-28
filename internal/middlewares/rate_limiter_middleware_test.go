package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"code-base-golang/internal/middlewares"
)

func TestRateLimit_AllowedAndBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	// Allow 1 request per second with burst 2
	router.Use(middlewares.RateLimit(1.0, 2))
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// First request: OK (within burst 2)
	req1 := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("expected status 200 on req 1, got %d", w1.Code)
	}

	// Second request: OK (within burst 2)
	req2 := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected status 200 on req 2, got %d", w2.Code)
	}

	// Third request immediate: should exceed burst and get 429
	req3 := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	if w3.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status 429 on req 3, got %d", w3.Code)
	}
}

func TestRateLimit_Disabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	// Disabled
	router.Use(middlewares.RateLimit(1.0, 1, false))
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200 when disabled, got %d on iteration %d", w.Code, i)
		}
	}
}
