package health

import (
	"time"

	"github.com/gin-gonic/gin"

	"code-base-golang/internal/constants"
	"code-base-golang/internal/pkg/controller"
)

// Controller handles HTTP endpoints for the health check domain.
// It embeds controller.Base to inherit standard response envelopes and error handlers.
type Controller struct {
	controller.Base
	uc *Usecase
}

// NewController creates a new health domain controller instance.
func NewController(base controller.Base, uc *Usecase) *Controller {
	return &Controller{
		Base: base,
		uc:   uc,
	}
}

// HealthCheck handles health status inquiries.
func (c *Controller) HealthCheck(ctx *gin.Context) {
	if c.uc == nil {
		c.RespondOK(ctx, constants.ResponseMessageSuccess, Data{
			Version: c.Cfg.Version,
			GitHash: c.Cfg.GitHash,
			Uptime:  time.Since(c.StartedAt).String(),
			Services: Services{
				Database: constants.IntegrationStatusDisconnected,
				Redis:    constants.IntegrationStatusDisconnected,
				S3:       constants.IntegrationStatusDisconnected,
				RabbitMQ: constants.IntegrationStatusDisconnected,
				SMTP:     constants.IntegrationStatusDisconnected,
			},
		})
		return
	}

	servicesStatus := c.uc.CheckHealth(ctx.Request.Context())

	c.RespondOK(ctx, constants.ResponseMessageSuccess, Data{
		Version:  c.Cfg.Version,
		GitHash:  c.Cfg.GitHash,
		Uptime:   time.Since(c.StartedAt).String(),
		Services: servicesStatus,
	})
}
