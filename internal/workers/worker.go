package workers

import (
	"context"
	"log"

	"code-base-golang/internal/config"
	"code-base-golang/internal/modules"
	"code-base-golang/internal/queue"
)

// WorkerServer manages and registers all background worker subscriptions.
type WorkerServer struct {
	cfg  config.Config
	mods *modules.Modules
	sub  EventSubscriber
}

// New creates a new WorkerServer instance.
func New(cfg config.Config, mods *modules.Modules, sub EventSubscriber) *WorkerServer {
	return &WorkerServer{
		cfg:  cfg,
		mods: mods,
		sub:  sub,
	}
}

// RegisterWorker registers all worker subscriptions to Domain Events.
func (w *WorkerServer) RegisterWorker() {
	if w.sub == nil {
		return
	}

	log.Println("Registering worker subscriptions to Domain Events...")

	// Register domain event handlers here, for example:
	// w.sub.MustSubscribe(constants.TopicUserRegistered, w.HandleUserRegistered)

	log.Println("Done RegisterWorker")
}

// handleMessage is the shared handler shell for all worker subscriptions. It
// decodes the raw message into T, runs process with the broker-provided context,
// and logs the lifecycle (start/finish) and any parse/processing error uniformly.
// It returns true to ACK the message or false to NACK (routing to DLQ/retry),
// keeping individual handlers free of repetitive parse/log/ack boilerplate.
//
// Typical handler:
//
//	func (w *WorkerServer) HandleUserRegistered(ctx context.Context, msg []byte) bool {
//	    return handleMessage(ctx, "user.registered", msg, w.uc.NotifyUserRegistered)
//	}
func handleMessage[T any](ctx context.Context, name string, msg []byte, process func(context.Context, T) error) bool {
	var payload T
	if err := queue.Parse(msg, &payload); err != nil {
		log.Printf("[WORKER ERROR] %s: failed to parse payload: %v", name, err)
		return false
	}

	log.Printf("[WORKER] %s: started", name)
	if err := process(ctx, payload); err != nil {
		log.Printf("[WORKER ERROR] %s: %v", name, err)
		return false
	}
	log.Printf("[WORKER] %s: finished", name)
	return true
}
