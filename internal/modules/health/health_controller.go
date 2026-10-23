package health

import (
	"time"

	"github.com/gin-gonic/gin"

	"code-base-golang/internal/config"
	"code-base-golang/internal/constants"
	"code-base-golang/internal/pkg/response"
)

// Controller handles HTTP endpoints for the health check domain.
type Controller struct {
	cfg       config.Config
	startedAt time.Time
	uc        *Usecase
}

// NewController creates a new health domain controller instance.
func NewController(cfg config.Config, startedAt time.Time, uc *Usecase) *Controller {
	return &Controller{
		cfg:       cfg,
		startedAt: startedAt,
		uc:        uc,
	}
}

// HealthCheck handles health status inquiries.
func (c *Controller) HealthCheck(ctx *gin.Context) {
	response.OK(ctx, constants.ResponseMessageSuccess, Data{
		Version:  c.cfg.Version,
		GitHash:  c.cfg.GitHash,
		Uptime:   time.Since(c.startedAt).String(),
		Services: c.uc.CheckHealth(ctx.Request.Context()),
	})
}
