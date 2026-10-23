package modules

import (
	"code-base-golang/internal/modules/docs"
	"code-base-golang/internal/modules/health"
	"code-base-golang/internal/modules/metrics"
	"code-base-golang/internal/pkg/app"
)

// Modules is the central application registry aggregating active domain modules.
type Modules struct {
	Health  *health.Controller
	Docs    *docs.Controller
	Metrics *metrics.Controller
}

// New initializes all active domain modules using the shared application context.
func New(appCtx *app.Context) *Modules {
	return &Modules{
		Health:  health.Register(appCtx),
		Docs:    docs.Register(appCtx),
		Metrics: metrics.Register(appCtx),
	}
}

// Handlers returns a list of all active module controllers for dynamic route dispatching.
func (m *Modules) Handlers() []any {
	return []any{
		m.Health,
		m.Docs,
		m.Metrics,
	}
}
