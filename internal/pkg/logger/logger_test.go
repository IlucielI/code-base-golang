package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"code-base-golang/internal/config"
	"code-base-golang/internal/pkg/ctxmeta"
)

func TestLogger_JSONHandlerWithContext(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.Config{AppEnv: "production"}
	l := New(cfg, WithWriter(&buf), WithLevel(slog.LevelDebug))

	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	tracer := tp.Tracer("test-tracer")

	ctx, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	reqID := "test-request-id-12345"
	ctx = ctxmeta.WithRequestID(ctx, reqID)

	userID := uuid.New()
	ctx = ctxmeta.WithAuthUser(ctx, ctxmeta.AuthUser{
		UserID: userID,
		Role:   "admin",
	})

	l.InfoContext(ctx, "hello world", "component", "testing")

	output := buf.String()
	if !strings.Contains(output, "hello world") {
		t.Fatalf("expected log to contain message, got: %s", output)
	}

	var parsed map[string]any
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("failed to parse JSON log output: %v", err)
	}

	if parsed["request_id"] != reqID {
		t.Errorf("expected request_id=%s, got %v", reqID, parsed["request_id"])
	}
	if parsed["user_id"] != userID.String() {
		t.Errorf("expected user_id=%s, got %v", userID.String(), parsed["user_id"])
	}
	if parsed["trace_id"] == nil || parsed["trace_id"] == "" {
		t.Error("expected trace_id in log output")
	}
	if parsed["span_id"] == nil || parsed["span_id"] == "" {
		t.Error("expected span_id in log output")
	}
	if parsed["component"] != "testing" {
		t.Errorf("expected component=testing, got %v", parsed["component"])
	}
}

func TestLogger_DevelopmentTextHandler(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.Config{AppEnv: "development"}
	l := New(cfg, WithWriter(&buf))

	ctx := ctxmeta.WithRequestID(context.Background(), "req-dev-999")
	l.WarnContext(ctx, "dev warning", "tag", "debug")

	out := buf.String()
	if !strings.Contains(out, "dev warning") {
		t.Errorf("expected text log to contain message, got: %s", out)
	}
	if !strings.Contains(out, "request_id=req-dev-999") {
		t.Errorf("expected text log to contain request_id, got: %s", out)
	}
}

func TestLogger_InitSetsDefault(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.Config{AppEnv: "production"}
	l := Init(cfg, WithWriter(&buf))
	if l == nil {
		t.Fatal("expected Init to return non-nil logger")
	}

	slog.Default().Info("default logger test")
	if !strings.Contains(buf.String(), "default logger test") {
		t.Errorf("expected default logger to write to buffer, got: %s", buf.String())
	}
}
