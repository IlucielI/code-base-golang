package metrics

import (
	"github.com/gin-gonic/gin"

	"code-base-golang/internal/pkg/handler"
	pkgmetrics "code-base-golang/internal/pkg/metrics"
)

// Controller handles Prometheus metrics scraping endpoint.
type Controller struct {
	handler.Base
}

// NewController creates a new metrics controller instance.
func NewController(base handler.Base) *Controller {
	return &Controller{Base: base}
}

// PrometheusMetrics serves Prometheus metrics for scraping.
func (c *Controller) PrometheusMetrics(ctx *gin.Context) {
	pkgmetrics.Handler().ServeHTTP(ctx.Writer, ctx.Request)
}
