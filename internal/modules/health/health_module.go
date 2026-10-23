package health

import (
	"code-base-golang/internal/adapters"
	"code-base-golang/internal/pkg/controller"
)

// Register initializes and wires the health module using the shared infrastructure adapters.
func Register(adp *adapters.Adapters) *Controller {
	pingers := make(map[string]Pinger)
	base := controller.Base{}

	if adp != nil {
		base = adp.Base
		if adp.DB != nil {
			pingers["database"] = adp.DB
		}
		if adp.Redis != nil {
			pingers["redis"] = adp.Redis
		}
		if adp.Storage != nil {
			pingers["s3"] = adp.Storage
		}
		if adp.Publisher != nil {
			pingers["rabbitmq"] = adp.Publisher
		}
		if adp.Mailer != nil {
			pingers["smtp"] = adp.Mailer
		}
	}

	return NewController(base, NewUsecase(pingers))
}
