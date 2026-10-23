package health

import (
	"time"

	"github.com/gin-gonic/gin"

	"code-base-golang/internal/constants"
	"code-base-golang/internal/pkg/handler"
)

// Controller handles HTTP endpoints for the health check domain.
// It embeds handler.Base to inherit standard response envelopes and error handlers.
type Controller struct {
	handler.Base
	uc *Usecase
}

// NewController creates a new health domain controller instance.
func NewController(base handler.Base, uc *Usecase) *Controller {
	return &Controller{
		Base: base,
		uc:   uc,
	}
}

// HealthCheck handles health status inquiries.
func (c *Controller) HealthCheck(ctx *gin.Context) {
	c.RespondOK(ctx, constants.ResponseMessageSuccess, Data{
		Version:  c.Cfg.Version,
		GitHash:  c.Cfg.GitHash,
		Uptime:   time.Since(c.StartedAt).String(),
		Services: c.uc.CheckHealth(ctx.Request.Context()),
	})
}
