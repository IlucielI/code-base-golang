package metrics

import (
	"code-base-golang/internal/adapters"
)

// Register initializes the metrics module using the shared infrastructure adapters.
func Register(_ *adapters.Adapters) *Controller {
	return NewController()
}
