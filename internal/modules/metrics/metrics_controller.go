package metrics

import (
	"github.com/gin-gonic/gin"

	pkgmetrics "code-base-golang/internal/pkg/metrics"
)

// Controller handles Prometheus metrics scraping endpoint.
type Controller struct{}

// NewController creates a new metrics controller instance.
func NewController() *Controller {
	return &Controller{}
}

// PrometheusMetrics serves Prometheus metrics for scraping.
func (c *Controller) PrometheusMetrics(ctx *gin.Context) {
	pkgmetrics.Handler().ServeHTTP(ctx.Writer, ctx.Request)
}
