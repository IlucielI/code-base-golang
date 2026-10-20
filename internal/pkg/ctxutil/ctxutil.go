// Package ctxutil provides small, reusable helpers for working with
// context.Context, such as running a bounded operation against a derived
// timeout context.
//
// It is intentionally free of any domain vocabulary so it can be used from any
// layer without introducing coupling.
package ctxutil

import (
	"context"
	"time"
)

// DefaultProbeTimeout bounds a single connectivity probe. It is deliberately
// short so a caller (such as a health check) never blocks on an unresponsive
// dependency.
const DefaultProbeTimeout = 2 * time.Second

// Operation is any function that honours context cancellation and reports an error.
type Operation func(context.Context) error

// RunWithTimeout runs op against a child context bounded by timeout. When
// timeout <= 0, DefaultProbeTimeout is used. The child context is always
// cancelled before returning, so callers need no additional cleanup.
func RunWithTimeout(ctx context.Context, timeout time.Duration, op Operation) error {
	if timeout <= 0 {
		timeout = DefaultProbeTimeout
	}
	opCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return op(opCtx)
}
