package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMetrics_CollectionAndHandler(t *testing.T) {
	EnsureInit()

	IncInFlight()
	ObserveHTTPRequest("GET", "/v1/health", "200", 0.045)
	DecInFlight()

	h := Handler()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from metrics handler, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "http_requests_total") {
		t.Errorf("expected metrics output to contain http_requests_total, got: %s", body)
	}
	if !strings.Contains(body, "http_request_duration_seconds") {
		t.Errorf("expected metrics output to contain http_request_duration_seconds, got: %s", body)
	}
	if !strings.Contains(body, "http_requests_in_flight") {
		t.Errorf("expected metrics output to contain http_requests_in_flight, got: %s", body)
	}
	if !strings.Contains(body, `path="/v1/health"`) {
		t.Errorf("expected metrics output to contain path label, got: %s", body)
	}
}
