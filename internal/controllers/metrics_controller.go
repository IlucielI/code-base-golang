package controllers

import (
	"github.com/gin-gonic/gin"

	"code-base-golang/internal/pkg/metrics"
)

// PrometheusMetrics serves Prometheus metrics for scraping.
func (c *Controllers) PrometheusMetrics(ctx *gin.Context) {
	metrics.Handler().ServeHTTP(ctx.Writer, ctx.Request)
}
