package workers

import (
	"context"
	"encoding/json"
	"log"
)

// Server orchestrates registration of domain consumers to message broker events.
type Server struct {
	sub       EventSubscriber
	consumers []Consumer
}

// WorkerServer is an alias to Server for backwards compatibility.
type WorkerServer = Server

// New creates a new background worker Server instance.
func New(sub EventSubscriber, consumers ...Consumer) *Server {
	return &Server{
		sub:       sub,
		consumers: consumers,
	}
}

// RegisterWorker registers all configured domain consumers to the event subscriber.
func (s *Server) RegisterWorker() {
	if s.sub == nil {
		return
	}

	log.Println("Registering worker subscriptions to Domain Events...")
	for _, consumer := range s.consumers {
		if consumer != nil {
			consumer.RegisterWorker(s.sub)
		}
	}
	log.Println("Done RegisterWorker")
}

// HandleMessage is the shared handler shell for all worker subscriptions.
// It decodes the raw message into T, runs process with the broker-provided context,
// and logs the lifecycle (start/finish) and any parse/processing error uniformly.
// It returns true to ACK the message or false to NACK (routing to DLQ/retry),
// keeping individual handlers free of repetitive parse/log/ack boilerplate.
//
// Typical domain consumer handler:
//
//	func (c *Consumer) HandleUserRegistered(ctx context.Context, msg []byte) bool {
//	    return workers.HandleMessage(ctx, "user.registered", msg, c.uc.NotifyUserRegistered)
//	}
func HandleMessage[T any](ctx context.Context, name string, msg []byte, process func(context.Context, T) error) bool {
	var payload T
	if err := json.Unmarshal(msg, &payload); err != nil {
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
