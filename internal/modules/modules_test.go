package modules_test

import (
	"testing"

	"code-base-golang/internal/config"
	"code-base-golang/internal/modules"
)

func TestModules_New(t *testing.T) {
	cfg := config.Config{AppName: "test-app", Version: "1.0.0"}
	mods := modules.New(cfg, nil, nil, nil, nil, nil)

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

func TestModules_Handlers_NilSafe(t *testing.T) {
	var m *modules.Modules
	if h := m.Handlers(); h != nil {
		t.Errorf("expected nil from nil modules handlers, got %v", h)
	}
}
