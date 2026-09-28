package workers

import (
	"log"

	"code-base-golang/internal/config"
	"code-base-golang/internal/services"
)

// WorkerServer manages and registers all background worker subscriptions.
type WorkerServer struct {
	cfg config.Config
	svc *services.Service
	sub EventSubscriber
}

// New creates a new WorkerServer instance.
func New(cfg config.Config, svc *services.Service, sub EventSubscriber) *WorkerServer {
	return &WorkerServer{
		cfg: cfg,
		svc: svc,
		sub: sub,
	}
}

// RegisterWorker registers all worker subscriptions to Domain Events.
func (w *WorkerServer) RegisterWorker() {
	if w == nil || w.sub == nil {
		return
	}

	log.Println("Registering worker subscriptions to Domain Events...")

	// Register domain event handlers here, for example:
	// w.sub.MustSubscribe(constants.TopicUserRegistered, w.HandleUserRegistered)

	log.Println("Done RegisterWorker")
}
