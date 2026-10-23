package docs

import (
	"code-base-golang/internal/adapters"
)

// Register initializes the docs module using the shared infrastructure adapters.
func Register(_ *adapters.Adapters) *Controller {
	return NewController()
}
