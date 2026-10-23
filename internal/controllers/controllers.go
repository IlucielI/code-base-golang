package controllers

import (
	"code-base-golang/internal/config"
	"code-base-golang/internal/modules/health"
	"code-base-golang/internal/pkg/controller"
	"code-base-golang/internal/usecases"
)

// Controllers is the central container aggregating all module handlers and global endpoints.
type Controllers struct {
	controller.Base
	*health.Controller
	uc *usecases.Usecases
}

// New initializes the central controllers container with application dependencies.
func New(cfg config.Config, uc *usecases.Usecases) *Controllers {
	base := controller.NewBase(cfg)
	var healthCtrl *health.Controller
	if uc != nil && uc.Health != nil {
		healthCtrl = health.NewController(base, uc.Health)
	} else {
		healthCtrl = health.NewController(base, nil)
	}

	return &Controllers{
		Base:       base,
		Controller: healthCtrl,
		uc:         uc,
	}
}

// SetUsecases allows overriding or injecting custom / mock Usecases for tests.
func (c *Controllers) SetUsecases(u *usecases.Usecases) {
	c.uc = u
	if u != nil && u.Health != nil {
		c.Controller = health.NewController(c.Base, u.Health)
	}
}

// Usecases returns the underlying usecases container.
func (c *Controllers) Usecases() *usecases.Usecases {
	return c.uc
}
