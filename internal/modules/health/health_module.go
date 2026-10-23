package health

import (
	"code-base-golang/internal/pkg/app"
	"code-base-golang/internal/pkg/controller"
)

// Register initializes and wires the health module using the shared application context.
func Register(appCtx *app.Context) *Controller {
	pingers := make(map[string]Pinger)
	base := controller.Base{}

	if appCtx != nil {
		base = appCtx.Base
		if appCtx.DB != nil {
			pingers["database"] = appCtx.DB
		}
		if appCtx.Redis != nil {
			pingers["redis"] = appCtx.Redis
		}
		if appCtx.Storage != nil {
			pingers["s3"] = appCtx.Storage
		}
		if appCtx.Publisher != nil {
			pingers["rabbitmq"] = appCtx.Publisher
		}
		if appCtx.Mailer != nil {
			pingers["smtp"] = appCtx.Mailer
		}
	}

	return NewController(base, NewUsecase(pingers))
}
