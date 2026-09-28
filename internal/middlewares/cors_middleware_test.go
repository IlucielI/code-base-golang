package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"code-base-golang/internal/middlewares"
)

func TestCORS_PreflightOptions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middlewares.CORS())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status 204 No Content for OPTIONS, got %d", w.Code)
	}

	if origin := w.Header().Get("Access-Control-Allow-Origin"); origin != "http://localhost:3000" {
		t.Errorf("expected Access-Control-Allow-Origin: http://localhost:3000, got %s", origin)
	}

	if methods := w.Header().Get("Access-Control-Allow-Methods"); methods == "" {
		t.Error("expected non-empty Access-Control-Allow-Methods")
	}
}

func TestCORS_SpecificOrigins(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middlewares.CORS(middlewares.CORSConfig{
		AllowedOrigins: []string{"https://app.example.com"},
		AllowedMethods: []string{"GET", "POST"},
		MaxAge:         time.Hour,
	}))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// 1. Allowed origin
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://app.example.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if origin := w.Header().Get("Access-Control-Allow-Origin"); origin != "https://app.example.com" {
		t.Errorf("expected allowed origin https://app.example.com, got %s", origin)
	}

	// 2. Disallowed origin
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.Header.Set("Origin", "https://malicious.com")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if origin := w2.Header().Get("Access-Control-Allow-Origin"); origin != "" {
		t.Errorf("expected empty Access-Control-Allow-Origin for disallowed origin, got %s", origin)
	}
}
