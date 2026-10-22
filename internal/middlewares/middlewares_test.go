package middlewares_test

import (
	"testing"

	"code-base-golang/internal/config"
	"code-base-golang/internal/middlewares"
)

func TestMiddlewaresContainer(t *testing.T) {
	cfg := config.Config{
		CORSAllowedOrigins: []string{"*"},
		RateLimiterEnabled: true,
		RateLimiterLimit:   20,
		RateLimiterBurst:   30,
		MaxRequestBodySize: 1024 * 1024,
		BasicAuthUsername:  "admin",
		BasicAuthPassword:  "secret",
	}

	container := middlewares.New(cfg)

	if authFn := container.Auth(); authFn == nil {
		t.Error("expected container.Auth() to return non-nil handler")
	}

	if optFn := container.OptionalAuth(); optFn == nil {
		t.Error("expected container.OptionalAuth() to return non-nil handler")
	}

	if permFn := container.RequirePermission("users:read"); permFn == nil {
		t.Error("expected container.RequirePermission() to return non-nil handler")
	}

	if corsFn := container.CORS(); corsFn == nil {
		t.Error("expected container.CORS() to return non-nil handler")
	}

	if rlFn := container.RateLimit(); rlFn == nil {
		t.Error("expected container.RateLimit() to return non-nil handler")
	}

	if logFn := container.Logger(); logFn == nil {
		t.Error("expected container.Logger() to return non-nil handler")
	}

	if recFn := container.Recovery(); recFn == nil {
		t.Error("expected container.Recovery() to return non-nil handler")
	}

	if metaFn := container.ClientMeta(); metaFn == nil {
		t.Error("expected container.ClientMeta() to return non-nil handler")
	}

	if bodyFn := container.BodyLimit(); bodyFn == nil {
		t.Error("expected container.BodyLimit() to return non-nil handler")
	}

	if basicFn := container.BasicAuth(); basicFn == nil {
		t.Error("expected container.BasicAuth() to return non-nil handler")
	}

	if basicExplicit := container.BasicAuth("custom", "pass"); basicExplicit == nil {
		t.Error("expected container.BasicAuth(user, pass) to return non-nil handler")
	}
}
