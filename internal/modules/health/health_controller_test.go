package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"code-base-golang/internal/config"
	"code-base-golang/internal/constants"
	"code-base-golang/internal/pkg/handler"
	"code-base-golang/internal/pkg/response"
)

func TestController_HealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := config.Config{Version: "1.0.0", GitHash: "abcdef"}
	base := handler.NewBase(cfg)
	uc := NewUsecase(&mockChecker{})
	ctrl := NewController(base, uc)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest(http.MethodGet, "/v1/health", nil)

	ctrl.HealthCheck(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp response.APIResponse[Data]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Status != constants.ResponseStatusSuccess {
		t.Errorf("expected success status, got %s", resp.Status)
	}
	if resp.Data.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", resp.Data.Version)
	}
}
