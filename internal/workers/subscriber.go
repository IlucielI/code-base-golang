package workers

import "context"

// WorkerHandler is the function signature for handling domain event messages.
// Return true to ACK the message, or false to NACK and route to DLQ/retry.
type WorkerHandler = func(ctx context.Context, msg []byte) (ack bool)

// EventSubscriber defines the abstract contract for message broker subscriptions.
// Any message broker adapter (RabbitMQ, NATS, Kafka) that provides these methods
// can be used to drive background workers without changing worker logic.
type EventSubscriber interface {
	// Subscribe registers a worker handler for the given topic/subject.
	Subscribe(topic string, handler WorkerHandler) error

	// MustSubscribe registers a handler and panics on configuration or binding failure.
	MustSubscribe(topic string, handler WorkerHandler)
}
