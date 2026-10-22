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

// NewRouter constructs the central Gin engine and wires controllers and middleware.
func NewRouter(cfg config.Config, ctrls *controllers.Controllers) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Central middlewares container
	mws := middlewares.New(cfg)
	mwsVal := reflect.ValueOf(mws)

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

		// Dynamically dispatch route-level middlewares defined in routes.yaml via Middlewares container
		for _, mw := range r.Middlewares {
			mwClean := strings.TrimSpace(mw)
			if mwClean == "" {
				continue
			}
			mwMethodName := toPascalCase(mwClean)
			mwMethod := mwsVal.MethodByName(mwMethodName)

			if !mwMethod.IsValid() {
				log.Fatalf("route configuration error: middleware '%s' (method '%s') not found on Middlewares for route %s %s", mwClean, mwMethodName, r.Method, r.Path)
			}

			var mwHandler gin.HandlerFunc
			isZeroArg := mwMethod.Type().NumIn() == 0
			isVariadicZeroArg := mwMethod.Type().IsVariadic() && mwMethod.Type().NumIn() == 1
			if (isZeroArg || isVariadicZeroArg) && mwMethod.Type().NumOut() == 1 {
				res := mwMethod.Call(nil)[0].Interface()
				var isHandler bool
				mwHandler, isHandler = res.(gin.HandlerFunc)
				if !isHandler {
					log.Fatalf("route configuration error: middleware method '%s' must return gin.HandlerFunc", mwMethodName)
				}
			} else if mwMethod.Type().NumIn() == 1 && mwMethod.Type().NumOut() == 0 {
				var isHandler bool
				mwHandler, isHandler = mwMethod.Interface().(func(*gin.Context))
				if !isHandler {
					log.Fatalf("route configuration error: middleware method '%s' must have signature func(*gin.Context)", mwMethodName)
				}
			} else {
				log.Fatalf("route configuration error: middleware method '%s' must have signature func() gin.HandlerFunc or func(*gin.Context)", mwMethodName)
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

// toPascalCase converts snake_case or lowercase strings to PascalCase.
// e.g. "auth" -> "Auth", "optional_auth" -> "OptionalAuth", "client_meta" -> "ClientMeta".
func toPascalCase(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if !strings.Contains(s, "_") {
		return strings.ToUpper(s[:1]) + s[1:]
	}
	parts := strings.Split(s, "_")
	var b strings.Builder
	for _, p := range parts {
		if len(p) > 0 {
			b.WriteString(strings.ToUpper(p[:1]))
			b.WriteString(strings.ToLower(p[1:]))
		}
	}
	return b.String()
}
