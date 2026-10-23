package docs

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appdocs "code-base-golang/docs"
)

// Controller handles API documentation endpoints.
type Controller struct{}

// NewController creates a new docs controller instance.
func NewController() *Controller {
	return &Controller{}
}

// OpenAPISpec serves the raw OpenAPI 3.0 specification in YAML format.
func (c *Controller) OpenAPISpec(ctx *gin.Context) {
	ctx.Data(http.StatusOK, "application/x-yaml; charset=utf-8", appdocs.OpenAPISpec)
}

// APIDocs serves the modern Scalar API Reference interactive documentation page.
func (c *Controller) APIDocs(ctx *gin.Context) {
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", appdocs.DocsHTML)
}
