package modules

import (
	"code-base-golang/internal/adapters"
	"code-base-golang/internal/modules/docs"
	"code-base-golang/internal/modules/health"
	"code-base-golang/internal/modules/metrics"
	"code-base-golang/internal/pkg/controller"
	"code-base-golang/internal/workers"
)

// Modules is the central application registry aggregating active domain modules.
type Modules struct {
	Health  *health.Controller
	Docs    *docs.Controller
	Metrics *metrics.Controller
}

// New initializes all active domain modules using the shared infrastructure adapters.
func New(adp *adapters.Adapters) *Modules {
	return &Modules{
		Health:  health.Register(adp),
		Docs:    docs.Register(adp),
		Metrics: metrics.Register(adp),
	}
}

// Handlers returns a list of all active module controllers for dynamic route dispatching.
func (m *Modules) Handlers() []controller.Controller {
	return []controller.Controller{
		m.Health,
		m.Docs,
		m.Metrics,
	}
}

// Consumers returns a list of all active domain event consumers for background workers.
func (m *Modules) Consumers() []workers.Consumer {
	return []workers.Consumer{
		// Domain consumers are added here as new modules implement workers.Consumer
	}
}
