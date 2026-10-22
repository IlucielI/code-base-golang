package routes

import (
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestRouteItem_UnmarshalYAML(t *testing.T) {
	yamlData := `
routes:
  - method: GET
    path: /v1/test/public
    handler: HealthCheck

  - method: GET
    path: /v1/test/auth
    handler: HealthCheck
    middlewares: [auth]

  - method: PUT
    path: /v1/test/permission
    handler: HealthCheck
    middlewares: [auth]
    permission: "users:write"

  - method: POST
    path: /v1/test/optional
    handler: HealthCheck
    middlewares: [optional_auth]

  - method: GET
    path: /v1/test/single-middleware
    handler: HealthCheck
    middleware: auth
`

	var cfg routeConfig
	if err := yaml.Unmarshal([]byte(yamlData), &cfg); err != nil {
		t.Fatalf("failed to unmarshal yaml: %v", err)
	}

	if len(cfg.Routes) != 5 {
		t.Fatalf("expected 5 routes, got %d", len(cfg.Routes))
	}

	// 1. Public route
	r0 := cfg.Routes[0]
	if r0.Method != "GET" || r0.Path != "/v1/test/public" || len(r0.Middlewares) != 0 || r0.Permission != "" {
		t.Errorf("r0 mismatch: %+v", r0)
	}

	// 2. Auth route
	r1 := cfg.Routes[1]
	if !reflect.DeepEqual(r1.Middlewares, []string{"auth"}) {
		t.Errorf("r1 middlewares expected [auth], got %v", r1.Middlewares)
	}

	// 3. Permission route
	r2 := cfg.Routes[2]
	if !reflect.DeepEqual(r2.Middlewares, []string{"auth"}) || r2.Permission != "users:write" {
		t.Errorf("r2 mismatch: %+v", r2)
	}

	// 4. Optional auth
	r3 := cfg.Routes[3]
	if !reflect.DeepEqual(r3.Middlewares, []string{"optional_auth"}) {
		t.Errorf("r3 mismatch: %+v", r3)
	}

	// 5. Single scalar middleware
	r4 := cfg.Routes[4]
	if !reflect.DeepEqual(r4.Middlewares, []string{"auth"}) {
		t.Errorf("r4 mismatch: %+v", r4)
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"auth", "Auth"},
		{"basic_auth", "BasicAuth"},
		{"optional_auth", "OptionalAuth"},
		{"client_meta", "ClientMeta"},
		{"body_limit", "BodyLimit"},
		{"rate_limit", "RateLimit"},
		{"AdminAuth", "AdminAuth"},
		{"", ""},
	}

	for _, tt := range tests {
		got := toPascalCase(tt.input)
		if got != tt.expected {
			t.Errorf("toPascalCase(%q) = %q, expected %q", tt.input, got, tt.expected)
		}
	}
}
