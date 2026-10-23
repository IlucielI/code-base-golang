package adapters_test

import (
	"testing"

	"code-base-golang/internal/adapters"
	"code-base-golang/internal/config"
)

func TestNewAdapters(t *testing.T) {
	cfg := config.Config{AppName: "test-app"}
	adp := adapters.New(cfg, nil, nil, nil, nil, nil)

	if adp == nil {
		t.Fatal("expected non-nil Adapters")
	}
	if adp.Config.AppName != "test-app" {
		t.Errorf("expected AppName 'test-app', got %s", adp.Config.AppName)
	}
}
