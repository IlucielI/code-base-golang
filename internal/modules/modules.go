package modules

import (
	"code-base-golang/internal/config"
	"code-base-golang/internal/modules/docs"
	"code-base-golang/internal/modules/health"
	"code-base-golang/internal/modules/metrics"
	"code-base-golang/internal/pkg/controller"
)

// Modules is the central application registry aggregating all domain modules.
type Modules struct {
	Base    controller.Base
	Health  *health.Controller
	Docs    *docs.Controller
	Metrics *metrics.Controller
}

// New initializes and wires all application domain modules with infrastructure adapters.
func New(
	cfg config.Config,
	db health.Pinger,
	redis health.Pinger,
	storage health.Pinger,
	publisher health.Pinger,
	mailer health.Pinger,
) *Modules {
	base := controller.NewBase(cfg)
	healthUc := health.NewUsecase(db, redis, storage, publisher, mailer)

	return &Modules{
		Base:    base,
		Health:  health.NewController(base, healthUc),
		Docs:    docs.NewController(base),
		Metrics: metrics.NewController(base),
	}
}

// Handlers returns a list of all active module controllers for dynamic route dispatching.
func (m *Modules) Handlers() []any {
	if m == nil {
		return nil
	}
	return []any{
		m.Health,
		m.Docs,
		m.Metrics,
	}
}
