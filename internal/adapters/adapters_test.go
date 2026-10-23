package adapters

import (
	"context"
	"testing"

	"code-base-golang/internal/config"
)

func TestNewAdapters(t *testing.T) {
	cfg := config.Config{AppName: "test-app"}
	adp := New(cfg, nil, nil, nil, nil, nil)

	if adp == nil {
		t.Fatal("expected non-nil Adapters")
	}
	if adp.Config.AppName != "test-app" {
		t.Errorf("expected AppName 'test-app', got %s", adp.Config.AppName)
	}

	ctx := context.Background()
	if err := adp.PingDatabase(ctx); err == nil {
		t.Error("expected error pinging nil database")
	}
	if err := adp.PingRedis(ctx); err == nil {
		t.Error("expected error pinging nil redis")
	}
	if err := adp.PingStorage(ctx); err == nil {
		t.Error("expected error pinging nil storage")
	}
	if err := adp.PingPublisher(ctx); err == nil {
		t.Error("expected error pinging nil publisher")
	}
	if err := adp.PingMailer(ctx); err == nil {
		t.Error("expected error pinging nil mailer")
	}
}

func TestAdapters_Ping_NilReceiver(t *testing.T) {
	var adp *Adapters
	ctx := context.Background()

	if err := adp.PingDatabase(ctx); err == nil {
		t.Error("expected error for nil receiver PingDatabase")
	}
	if err := adp.PingRedis(ctx); err == nil {
		t.Error("expected error for nil receiver PingRedis")
	}
	if err := adp.PingStorage(ctx); err == nil {
		t.Error("expected error for nil receiver PingStorage")
	}
	if err := adp.PingPublisher(ctx); err == nil {
		t.Error("expected error for nil receiver PingPublisher")
	}
	if err := adp.PingMailer(ctx); err == nil {
		t.Error("expected error for nil receiver PingMailer")
	}
}
