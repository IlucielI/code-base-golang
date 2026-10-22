package workers_test

import (
	"testing"

	"code-base-golang/internal/config"
	"code-base-golang/internal/queue"
	"code-base-golang/internal/workers"
)

type mockSubscriber struct {
	subscribedTopics []string
}

func (m *mockSubscriber) Subscribe(topic string, handler workers.WorkerHandler) error {
	m.subscribedTopics = append(m.subscribedTopics, topic)
	return nil
}

func (m *mockSubscriber) MustSubscribe(topic string, handler workers.WorkerHandler) {
	_ = m.Subscribe(topic, handler)
}

func TestWorkerServer_NilSubscriber(t *testing.T) {
	server := workers.New(config.Config{}, nil, nil)
	server.RegisterWorker() // should safely skip when subscriber is nil

	mock := &mockSubscriber{}
	serverWithMock := workers.New(config.Config{}, nil, mock)
	serverWithMock.RegisterWorker() // should execute register cleanly
}

func TestQueue_MustParse(t *testing.T) {
	data := []byte(`{"id":"entity-123"}`)
	var p queue.Id
	queue.MustParse(data, &p)
	if p.Id != "entity-123" {
		t.Errorf("expected entity-123, got %s", p.Id)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected MustParse to panic on invalid json")
		}
	}()
	queue.MustParse([]byte("invalid json"), &p)
}

func TestQueue_Parse(t *testing.T) {
	data := []byte(`{"id":"order-999"}`)
	var p queue.Id
	if err := queue.Parse(data, &p); err != nil {
		t.Errorf("expected nil error parsing payload, got %v", err)
	}
	if p.Id != "order-999" {
		t.Errorf("expected order-999, got %s", p.Id)
	}

	if err := queue.Parse([]byte("bad json"), &p); err == nil {
		t.Error("expected error parsing invalid json, got nil")
	}
}
