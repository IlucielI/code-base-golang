package modules

import (
	"testing"

	"code-base-golang/internal/adapters"
	"code-base-golang/internal/config"
)

func TestModules_New(t *testing.T) {
	cfg := config.Config{AppName: "test-app", Version: "1.0.0"}
	adp := adapters.New(cfg, nil, nil, nil, nil, nil)
	mods := New(adp)

	if mods == nil {
		t.Fatal("expected non-nil modules")
	}
	if mods.Health == nil {
		t.Error("expected non-nil health controller")
	}
	if mods.Docs == nil {
		t.Error("expected non-nil docs controller")
	}
	if mods.Metrics == nil {
		t.Error("expected non-nil metrics controller")
	}

	handlers := mods.Handlers()
	if len(handlers) != 3 {
		t.Errorf("expected 3 handlers, got %d", len(handlers))
	}
}
