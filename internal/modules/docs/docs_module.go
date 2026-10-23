package docs

import (
	"code-base-golang/internal/pkg/app"
	"code-base-golang/internal/pkg/controller"
)

// Register initializes the docs module using the shared application context.
func Register(appCtx *app.Context) *Controller {
	if appCtx == nil {
		return NewController(controller.Base{})
	}
	return NewController(appCtx.Base)
}
