package services

import (
	"context"
)

// EventPublisher defines the abstract contract for publishing domain events.
// Any message broker adapter (RabbitMQ, NATS, Kafka, Redis Streams) that implements
// these methods can be seamlessly injected into the application services.
type EventPublisher interface {
	// Ping checks if the message broker is reachable.
	Ping(ctx context.Context) error

	// Publish serializes and publishes an event payload to the given topic/subject.
	Publish(ctx context.Context, topic string, payload any) error

	// MustPublish publishes an event payload and panics on failure.
	MustPublish(ctx context.Context, topic string, payload any)
}
