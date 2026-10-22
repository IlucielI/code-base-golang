package logger

import (
	"context"
	"io"
	"log/slog"
	"os"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"

	"code-base-golang/internal/config"
	"code-base-golang/internal/pkg/ctxmeta"
)

// ContextHandler wraps an existing slog.Handler to inject contextual fields
// such as request_id, trace_id, span_id, and user_id.
type ContextHandler struct {
	handler slog.Handler
}

// NewContextHandler creates a ContextHandler that extracts request_id, trace_id,
// and user_id from context.Context.
func NewContextHandler(h slog.Handler) *ContextHandler {
	return &ContextHandler{handler: h}
}

// Enabled reports whether the handler handles records at the given level.
func (h *ContextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

// Handle injects context metadata before delegating to the wrapped handler.
func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if ctx != nil {
		if reqID := ctxmeta.GetRequestID(ctx); reqID != "" {
			r.AddAttrs(slog.String("request_id", reqID))
		}

		span := trace.SpanFromContext(ctx)
		if span != nil && span.SpanContext().IsValid() {
			sc := span.SpanContext()
			r.AddAttrs(
				slog.String("trace_id", sc.TraceID().String()),
				slog.String("span_id", sc.SpanID().String()),
			)
		}

		if authUser, ok := ctxmeta.GetAuthUser(ctx); ok && authUser.UserID != uuid.Nil {
			r.AddAttrs(slog.String("user_id", authUser.UserID.String()))
		}
	}

	return h.handler.Handle(ctx, r)
}

// WithAttrs returns a new ContextHandler whose attributes consist of
// both the receiver's attributes and the arguments.
func (h *ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ContextHandler{handler: h.handler.WithAttrs(attrs)}
}

// WithGroup returns a new ContextHandler with the given group name.
func (h *ContextHandler) WithGroup(name string) slog.Handler {
	return &ContextHandler{handler: h.handler.WithGroup(name)}
}

// Option configures logger creation.
type Option func(*options)

type options struct {
	writer io.Writer
	level  slog.Level
}

// WithWriter overrides the output destination.
func WithWriter(w io.Writer) Option {
	return func(o *options) {
		if w != nil {
			o.writer = w
		}
	}
}

// WithLevel sets the minimum log level.
func WithLevel(lvl slog.Level) Option {
	return func(o *options) {
		o.level = lvl
	}
}

// New initializes a slog.Logger with context enrichment and environment-aware formatting.
func New(cfg config.Config, opts ...Option) *slog.Logger {
	o := options{
		writer: os.Stdout,
		level:  slog.LevelInfo,
	}
	for _, opt := range opts {
		opt(&o)
	}

	handlerOpts := &slog.HandlerOptions{
		Level: o.level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.String("timestamp", a.Value.Time().UTC().Format("2006-01-02T15:04:05Z07:00"))
			}
			if a.Key == slog.LevelKey {
				if lvl, ok := a.Value.Any().(slog.Level); ok {
					switch lvl {
					case slog.LevelDebug:
						return slog.String(slog.LevelKey, "DEBUG")
					case slog.LevelInfo:
						return slog.String(slog.LevelKey, "INFO")
					case slog.LevelWarn:
						return slog.String(slog.LevelKey, "WARN")
					case slog.LevelError:
						return slog.String(slog.LevelKey, "ERROR")
					}
				}
			}
			return a
		},
	}

	var baseHandler slog.Handler
	if cfg.AppEnv == "development" {
		baseHandler = slog.NewTextHandler(o.writer, handlerOpts)
	} else {
		baseHandler = slog.NewJSONHandler(o.writer, handlerOpts)
	}

	return slog.New(NewContextHandler(baseHandler))
}

// Init configures and sets the global default slog.Logger.
func Init(cfg config.Config, opts ...Option) *slog.Logger {
	l := New(cfg, opts...)
	slog.SetDefault(l)
	return l
}
