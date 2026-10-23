package workers

import (
	"testing"
)

type mockSubscriber struct {
	subscribedTopics []string
}

func (m *mockSubscriber) Subscribe(topic string, handler WorkerHandler) error {
	m.subscribedTopics = append(m.subscribedTopics, topic)
	return nil
}

func (m *mockSubscriber) MustSubscribe(topic string, handler WorkerHandler) {
	_ = m.Subscribe(topic, handler)
}

type mockConsumer struct {
	registered bool
}

func (m *mockConsumer) RegisterWorker(sub EventSubscriber) {
	m.registered = true
}

func TestServer_NilSubscriber(t *testing.T) {
	server := New(nil)
	server.RegisterWorker() // should safely skip when subscriber is nil
}

func TestServer_RegisterWorker_WithConsumers(t *testing.T) {
	mockSub := &mockSubscriber{}
	consumer := &mockConsumer{}

	server := New(mockSub, consumer, nil)
	server.RegisterWorker()

	if !consumer.registered {
		t.Error("expected consumer to be registered")
	}
}
