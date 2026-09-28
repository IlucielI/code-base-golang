package routes

import (
	_ "embed"
	"log"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"

	"code-base-golang/internal/config"
	"code-base-golang/internal/controllers"
	"code-base-golang/internal/middlewares"
)

//go:embed routes.yaml
var routesYAML []byte

type routeItem struct {
	Method    string `yaml:"method"`
	Path      string `yaml:"path"`
	Handler   string `yaml:"handler"`
	Auth      bool   `yaml:"auth,omitempty"`
	BasicAuth bool   `yaml:"basic_auth,omitempty"`
}

type routeConfig struct {
	Routes []routeItem `yaml:"routes"`
}

// NewRouter constructs the central Gin engine and wires controllers and middleware.
func NewRouter(cfg config.Config, ctrls *controllers.Controllers, authValidator ...middlewares.AuthValidator) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	corsCfg := middlewares.CORSConfig{
		AllowedOrigins:   cfg.CORSAllowedOrigins,
		AllowedMethods:   cfg.CORSAllowedMethods,
		AllowedHeaders:   cfg.CORSAllowedHeaders,
		AllowCredentials: cfg.CORSAllowCredentials,
		MaxAge:           cfg.CORSMaxAge,
	}

	router.Use(
		middlewares.StructuredLogger(),
		middlewares.Recovery(),
		middlewares.CORS(corsCfg),
		middlewares.RateLimit(cfg.RateLimiterLimit, cfg.RateLimiterBurst, cfg.RateLimiterEnabled),
		middlewares.ClientMeta(),
	)

	// Central controllers container
	if ctrls == nil {
		ctrls = controllers.New(cfg, nil)
	}
	ctrlsVal := reflect.ValueOf(ctrls)

	// Auth validator for protected routes
	var validator middlewares.AuthValidator
	if len(authValidator) > 0 && authValidator[0] != nil {
		validator = authValidator[0]
	}

	// Load and parse embedded YAML routes
	var rCfg routeConfig
	if err := yaml.Unmarshal(routesYAML, &rCfg); err != nil {
		log.Fatalf("failed to parse routes.yaml: %v", err)
	}

	// Auto-dispatch routes dynamically via reflection (zero manual mapping)
	for _, r := range rCfg.Routes {
		methodName := strings.TrimSpace(r.Handler)
		method := ctrlsVal.MethodByName(methodName)

		if !method.IsValid() {
			log.Fatalf("route configuration error: method '%s' not found on Controllers for route %s %s", methodName, r.Method, r.Path)
		}

		fn, ok := method.Interface().(func(*gin.Context))
		if !ok {
			log.Fatalf("route configuration error: method '%s' must have signature func(*gin.Context)", methodName)
		}

		var handlers []gin.HandlerFunc
		if r.BasicAuth {
			handlers = append(handlers, middlewares.BasicAuth(cfg.BasicAuthUsername, cfg.BasicAuthPassword))
		}
		if r.Auth {
			handlers = append(handlers, middlewares.Auth(validator))
		}
		handlers = append(handlers, fn)

		httpMethod := strings.ToUpper(strings.TrimSpace(r.Method))
		router.Handle(httpMethod, r.Path, handlers...)
	}

	return router
}
