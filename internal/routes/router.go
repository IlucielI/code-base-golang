package routes

import (
	_ "embed"
	"log"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"gopkg.in/yaml.v3"

	"code-base-golang/internal/config"
	"code-base-golang/internal/controllers"
	"code-base-golang/internal/middlewares"
)

//go:embed routes.yaml
var routesYAML []byte

type routeItem struct {
	Method     string `yaml:"method"`
	Path       string `yaml:"path"`
	Handler    string `yaml:"handler"`
	Auth       bool   `yaml:"auth,omitempty"`
	BasicAuth  bool   `yaml:"basic_auth,omitempty"`
	Permission string `yaml:"permission,omitempty"`
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

	// Auth validator for protected routes
	var validator middlewares.AuthValidator
	if len(authValidator) > 0 && authValidator[0] != nil {
		validator = authValidator[0]
	}

	// Central middlewares container
	mws := middlewares.New(cfg, validator)

	// OpenTelemetry HTTP server instrumentation (only when tracing is enabled).
	if cfg.TracingEnabled {
		router.Use(otelgin.Middleware(cfg.TracingServiceName))
	}

	// Global middleware stack
	router.Use(
		mws.Logger(),
		mws.Recovery(),
		mws.CORS(),
		mws.RateLimit(),
		mws.ClientMeta(),
		mws.BodyLimit(),
	)

	// Central controllers container. A fully-wired controller (with a live
	// service) is a hard precondition: failing fast here surfaces wiring bugs
	// instead of silently constructing a controller with a nil service.
	if ctrls == nil {
		log.Fatal("router configuration error: controllers must not be nil")
	}
	ctrlsVal := reflect.ValueOf(ctrls)

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
			handlers = append(handlers, mws.BasicAuth())
		}
		if r.Permission != "" {
			if !r.Auth {
				handlers = append(handlers, mws.Auth())
			}
			handlers = append(handlers, mws.RequirePermission(r.Permission))
		} else if r.Auth {
			handlers = append(handlers, mws.Auth())
		}
		handlers = append(handlers, fn)

		httpMethod := strings.ToUpper(strings.TrimSpace(r.Method))
		router.Handle(httpMethod, r.Path, handlers...)
	}

	return router
}
