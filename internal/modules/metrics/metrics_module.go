package metrics

import (
	"code-base-golang/internal/adapters"
	"code-base-golang/internal/pkg/controller"
)

// Register initializes the metrics module using the shared infrastructure adapters.
func Register(adp *adapters.Adapters) *Controller {
	if adp == nil {
		return NewController(controller.Base{})
	}
	return NewController(adp.Base)
}
