package health

import (
	"code-base-golang/internal/adapters"
)

// Register initializes and wires the health module using the shared infrastructure adapters.
func Register(adp *adapters.Adapters) *Controller {
	return NewController(adp.Base, NewUsecase(adp))
}
