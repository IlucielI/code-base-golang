package health

import (
	"code-base-golang/internal/adapters"
	"code-base-golang/internal/pkg/controller"
)

// Register initializes and wires the health module using the shared infrastructure adapters.
func Register(adp *adapters.Adapters) *Controller {
	if adp == nil {
		return NewController(controller.Base{}, NewUsecase(nil))
	}
	return NewController(adp.Base, NewUsecase(adp))
}
