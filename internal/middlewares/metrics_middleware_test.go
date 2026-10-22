package middlewares

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"code-base-golang/internal/config"
	"code-base-golang/internal/pkg/metrics"
)

func TestMetricsMiddleware(t *testing.T) {
	mws := New(config.Config{})
	router := gin.New()
	router.Use(mws.Metrics())

	router.GET("/v1/test-metric", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	router.GET("/metrics", func(c *gin.Context) {
		metrics.Handler().ServeHTTP(c.Writer, c.Request)
	})

	// Perform a request to trigger metrics
	req := httptest.NewRequest(http.MethodGet, "/v1/test-metric", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Scrape metrics
	scrapeReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	scrapeW := httptest.NewRecorder()
	router.ServeHTTP(scrapeW, scrapeReq)

	if scrapeW.Code != http.StatusOK {
		t.Fatalf("expected 200 from /metrics, got %d", scrapeW.Code)
	}

	body := scrapeW.Body.String()
	if !strings.Contains(body, `http_requests_total{method="GET",path="/v1/test-metric",status="200"}`) {
		t.Errorf("expected metric to record request, body: %s", body)
	}
}
