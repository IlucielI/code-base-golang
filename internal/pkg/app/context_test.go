package app_test

import (
	"testing"

	"code-base-golang/internal/config"
	"code-base-golang/internal/pkg/app"
)

func TestNewContext(t *testing.T) {
	cfg := config.Config{AppName: "test-app"}
	ctx := app.NewContext(cfg, nil, nil, nil, nil, nil)

	if ctx == nil {
		t.Fatal("expected non-nil app.Context")
	}
	if ctx.Config.AppName != "test-app" {
		t.Errorf("expected AppName 'test-app', got %s", ctx.Config.AppName)
	}
}
