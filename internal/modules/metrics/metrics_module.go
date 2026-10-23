package metrics

import (
	"code-base-golang/internal/adapters"
)

// Register initializes the metrics module using the shared infrastructure adapters.
func Register(adp *adapters.Adapters) *Controller {
	return NewController(adp.Base)
}
