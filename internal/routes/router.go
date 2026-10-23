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
	"code-base-golang/internal/middlewares"
	"code-base-golang/internal/modules"
	"code-base-golang/internal/pkg/strutil"
)

//go:embed routes.yaml
var routesYAML []byte

type routeItem struct {
	Method      string   `yaml:"method"`
	Path        string   `yaml:"path"`
	Handler     string   `yaml:"handler"`
	Permission  string   `yaml:"permission,omitempty"`
	Middlewares []string `yaml:"middlewares,omitempty"`
}

// UnmarshalYAML supports both plural slice 'middlewares: [auth]' and singular scalar 'middleware: auth'.
func (r *routeItem) UnmarshalYAML(node *yaml.Node) error {
	type rawRouteItem routeItem
	var raw rawRouteItem
	if err := node.Decode(&raw); err != nil {
		return err
	}
	*r = routeItem(raw)

	for i := 0; i < len(node.Content)-1; i += 2 {
		key := node.Content[i].Value
		if key == "middleware" {
			var single string
			if err := node.Content[i+1].Decode(&single); err == nil && single != "" {
				if len(r.Middlewares) == 0 {
					r.Middlewares = []string{single}
				}
			}
		}
	}
	return nil
}

type routeConfig struct {
	Routes []routeItem `yaml:"routes"`
}

// NewRouter constructs the central Gin engine and wires modules and middleware.
func NewRouter(cfg config.Config, mods *modules.Modules, idempotencyStore ...middlewares.IdempotencyStore) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Central middlewares container
	mws := middlewares.New(cfg)
	if len(idempotencyStore) > 0 && idempotencyStore[0] != nil {
		mws.SetIdempotencyStore(idempotencyStore[0])
	}
	mwsVal := reflect.ValueOf(mws)

	// OpenTelemetry HTTP server instrumentation (only when tracing is enabled).
	if cfg.TracingEnabled {
		router.Use(otelgin.Middleware(cfg.TracingServiceName))
	}

	// Global middleware stack
	router.Use(
		mws.Metrics(),
		mws.Logger(),
		mws.Recovery(),
		mws.CORS(),
		mws.RateLimit(),
		mws.ClientMeta(),
		mws.BodyLimit(),
	)

	// Central modules container verification.
	if mods == nil {
		log.Fatal("router configuration error: modules must not be nil")
	}
	modsVal := reflect.ValueOf(mods).Elem()

	// Load and parse embedded YAML routes
	var rCfg routeConfig
	if err := yaml.Unmarshal(routesYAML, &rCfg); err != nil {
		log.Fatalf("failed to parse routes.yaml: %v", err)
	}

	// Auto-dispatch routes dynamically across registered modules via reflection (zero manual mapping)
	for _, r := range rCfg.Routes {
		methodName := strings.TrimSpace(r.Handler)
		var method reflect.Value
		for i := 0; i < modsVal.NumField(); i++ {
			fieldVal := modsVal.Field(i)
			if (fieldVal.Kind() == reflect.Pointer || fieldVal.Kind() == reflect.Interface) && fieldVal.IsNil() {
				continue
			}
			m := fieldVal.MethodByName(methodName)
			if m.IsValid() {
				method = m
				break
			}
		}

		if !method.IsValid() {
			log.Fatalf("route configuration error: method '%s' not found across modules for route %s %s", methodName, r.Method, r.Path)
		}

		fn, ok := method.Interface().(func(*gin.Context))
		if !ok {
			log.Fatalf("route configuration error: method '%s' must have signature func(*gin.Context)", methodName)
		}

		var handlers []gin.HandlerFunc

		// Dynamically dispatch route-level middlewares defined in routes.yaml via Middlewares container
		for _, mw := range r.Middlewares {
			mwClean := strings.TrimSpace(mw)
			if mwClean == "" {
				continue
			}
			mwMethodName := strutil.ToPascalCase(mwClean)
			mwMethod := mwsVal.MethodByName(mwMethodName)

			if !mwMethod.IsValid() {
				log.Fatalf("route configuration error: middleware '%s' (method '%s') not found on Middlewares for route %s %s", mwClean, mwMethodName, r.Method, r.Path)
			}

			mwHandler, ok := mwMethod.Interface().(func(*gin.Context))
			if !ok {
				log.Fatalf("route configuration error: middleware method '%s' must have signature func(*gin.Context)", mwMethodName)
			}

			handlers = append(handlers, mwHandler)
		}

		if r.Permission != "" {
			handlers = append(handlers, mws.RequirePermission(r.Permission))
		}
		handlers = append(handlers, fn)

		httpMethod := strings.ToUpper(strings.TrimSpace(r.Method))
		router.Handle(httpMethod, r.Path, handlers...)
	}

	return router
}
