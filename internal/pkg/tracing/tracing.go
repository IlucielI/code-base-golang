// Package tracing wires OpenTelemetry distributed tracing for the application.
//
// It is intentionally opt-in: when tracing is disabled (the default) Init returns
// a no-op shutdown function and installs no global providers, so there is zero
// runtime overhead and no network connection attempts. When enabled, it builds a
// TracerProvider with a configurable exporter (OTLP/HTTP or stdout) and registers
// it globally along with the W3C trace-context + baggage propagators.
package tracing

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// Config holds the tracing settings resolved from application configuration.
// It deliberately mirrors only the fields tracing needs, so this package does
// not depend on the central config package.
type Config struct {
	Enabled      bool
	ServiceName  string
	Version      string
	Environment  string
	Exporter     string            // "otlp", "stdout", or "none"
	OTLPEndpoint string            // host:port for OTLP/HTTP, e.g. "localhost:4318"
	OTLPInsecure bool              // use plain HTTP instead of TLS
	OTLPHeaders  map[string]string // extra OTLP headers (e.g. auth tokens for managed backends)
	SampleRatio  float64           // 0.0..1.0; parent-based ratio sampler
}

// ShutdownFunc flushes buffered spans and releases exporter resources.
type ShutdownFunc func(context.Context) error

// noopShutdown is returned whenever tracing is not active.
func noopShutdown(context.Context) error { return nil }

// Init configures global OpenTelemetry tracing according to cfg.
//
// When cfg.Enabled is false (or the exporter is "none"), it installs nothing and
// returns a no-op shutdown. Otherwise it registers a global TracerProvider and
// propagator, and returns a shutdown function the caller MUST defer to flush
// spans on exit.
func Init(ctx context.Context, cfg Config) (ShutdownFunc, error) {
	exporterKind := strings.ToLower(strings.TrimSpace(cfg.Exporter))
	if !cfg.Enabled || exporterKind == "none" || exporterKind == "" {
		return noopShutdown, nil
	}

	exporter, err := newExporter(ctx, exporterKind, cfg)
	if err != nil {
		return noopShutdown, err
	}

	// resource.New may return a usable partial resource alongside a non-fatal
	// detector error; keep it, and only fall back to the default if nothing came back.
	res, err := newResource(ctx, cfg)
	if err != nil && res == nil {
		res = resource.Default()
	}

	ratio := cfg.SampleRatio
	if ratio <= 0 {
		ratio = 1.0
	}
	if ratio > 1 {
		ratio = 1.0
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(ratio))),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Telemetry pipeline errors (e.g. the collector being unreachable) are
	// operational, not application failures. Route them through a dedicated,
	// clearly-prefixed handler so they are greppable and never mistaken for app
	// errors, instead of the SDK default that writes to stderr.
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		if err == nil {
			return
		}
		log.Printf("[OTEL ERROR] telemetry pipeline error: %v", err)
	}))

	return func(shutdownCtx context.Context) error {
		return tp.Shutdown(shutdownCtx)
	}, nil
}

// newExporter builds a span exporter for the requested kind.
func newExporter(ctx context.Context, kind string, cfg Config) (sdktrace.SpanExporter, error) {
	switch kind {
	case "stdout":
		return stdouttrace.New(stdouttrace.WithPrettyPrint())
	case "otlp":
		opts := []otlptracehttp.Option{}
		if endpoint := strings.TrimSpace(cfg.OTLPEndpoint); endpoint != "" {
			opts = append(opts, otlptracehttp.WithEndpoint(endpoint))
		}
		if cfg.OTLPInsecure {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		if len(cfg.OTLPHeaders) > 0 {
			opts = append(opts, otlptracehttp.WithHeaders(cfg.OTLPHeaders))
		}
		dialCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		return otlptracehttp.New(dialCtx, opts...)
	default:
		return nil, fmt.Errorf("tracing: unsupported exporter %q (want otlp, stdout, or none)", kind)
	}
}

// newResource describes the service emitting the telemetry. It enriches the
// service identity (name/version/environment) with auto-detected host, process
// and telemetry-SDK attributes and any OTEL_RESOURCE_ATTRIBUTES from the
// environment, so that in multi-replica deployments each instance is
// distinguishable in the tracing backend.
func newResource(ctx context.Context, cfg Config) (*resource.Resource, error) {
	serviceName := strings.TrimSpace(cfg.ServiceName)
	if serviceName == "" {
		serviceName = "code-base-golang"
	}

	// resource.New already merges the provided detectors onto the SDK default
	// resource (telemetry.sdk.*), so no manual merge with resource.Default() is
	// needed. On a non-fatal detector error it still returns a usable resource.
	return resource.New(ctx,
		resource.WithFromEnv(),      // OTEL_RESOURCE_ATTRIBUTES / OTEL_SERVICE_NAME
		resource.WithTelemetrySDK(), // telemetry.sdk.{name,language,version}
		resource.WithHost(),         // host.name
		resource.WithProcess(),      // process.pid, executable, runtime, etc.
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(cfg.Version),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
	)
}

// Tracer returns a named tracer from the global provider. Safe to call even when
// tracing is disabled (it returns a no-op tracer).
func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}
