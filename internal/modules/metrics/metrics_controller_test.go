package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestController_PrometheusMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := NewController()

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/metrics", nil)

	ctrl.PrometheusMetrics(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from PrometheusMetrics, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "go_goroutines") {
		t.Errorf("expected metrics to include standard go runtime metrics, got: %s", body)
	}
}
