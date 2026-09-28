package workers_test

import (
	"testing"

	"code-base-golang/internal/config"
	"code-base-golang/internal/payload"
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

func TestWorkerServer_Nil(t *testing.T) {
	var s *workers.WorkerServer
	s.RegisterWorker() // should not panic

	server := workers.New(config.Config{}, nil, nil)
	server.RegisterWorker() // should safely skip when subscriber is nil

	mock := &mockSubscriber{}
	serverWithMock := workers.New(config.Config{}, nil, mock)
	serverWithMock.RegisterWorker() // should execute register cleanly
}

func TestPayload_MustParse(t *testing.T) {
	data := []byte(`{"id":"entity-123"}`)
	var p payload.Id
	payload.MustParse(data, &p)
	if p.Id != "entity-123" {
		t.Errorf("expected entity-123, got %s", p.Id)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected MustParse to panic on invalid json")
		}
	}()
	payload.MustParse([]byte("invalid json"), &p)
}

func TestPayload_Parse(t *testing.T) {
	data := []byte(`{"id":"order-999"}`)
	var p payload.Id
	if err := payload.Parse(data, &p); err != nil {
		t.Errorf("expected nil error parsing payload, got %v", err)
	}
	if p.Id != "order-999" {
		t.Errorf("expected order-999, got %s", p.Id)
	}

	if err := payload.Parse([]byte("bad json"), &p); err == nil {
		t.Error("expected error parsing invalid json, got nil")
	}
}
