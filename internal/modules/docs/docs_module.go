package docs

import (
	"code-base-golang/internal/adapters"
)

// Register initializes the docs module using the shared infrastructure adapters.
func Register(adp *adapters.Adapters) *Controller {
	return NewController(adp.Base)
}
