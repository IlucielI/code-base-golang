package health_test

import (
	"context"
	"errors"
	"testing"

	"code-base-golang/internal/constants"
	"code-base-golang/internal/modules/health"
)

type mockPinger struct {
	err error
}

func (m *mockPinger) Ping(ctx context.Context) error {
	return m.err
}

func TestUsecase_CheckHealth_EmptyDependencies(t *testing.T) {
	u := health.NewUsecase(nil, nil, nil, nil, nil)
	status := u.CheckHealth(context.Background())
	if status.Database != constants.IntegrationStatusDisconnected {
		t.Errorf("expected db disconnected, got %s", status.Database)
	}
	if status.Redis != constants.IntegrationStatusDisconnected {
		t.Errorf("expected redis disconnected, got %s", status.Redis)
	}
	if status.S3 != constants.IntegrationStatusDisconnected {
		t.Errorf("expected s3 disconnected, got %s", status.S3)
	}
	if status.RabbitMQ != constants.IntegrationStatusDisconnected {
		t.Errorf("expected rabbitmq disconnected, got %s", status.RabbitMQ)
	}
	if status.SMTP != constants.IntegrationStatusDisconnected {
		t.Errorf("expected smtp disconnected, got %s", status.SMTP)
	}
}

func TestUsecase_CheckHealth_Connected(t *testing.T) {
	db := &mockPinger{err: nil}
	rdb := &mockPinger{err: nil}
	s3 := &mockPinger{err: nil}
	rmq := &mockPinger{err: nil}
	smtp := &mockPinger{err: nil}

	u := health.NewUsecase(db, rdb, s3, rmq, smtp)
	status := u.CheckHealth(context.Background())

	if status.Database != constants.IntegrationStatusConnected {
		t.Errorf("expected db connected, got %s", status.Database)
	}
	if status.Redis != constants.IntegrationStatusConnected {
		t.Errorf("expected redis connected, got %s", status.Redis)
	}
	if status.S3 != constants.IntegrationStatusConnected {
		t.Errorf("expected s3 connected, got %s", status.S3)
	}
	if status.RabbitMQ != constants.IntegrationStatusConnected {
		t.Errorf("expected rabbitmq connected, got %s", status.RabbitMQ)
	}
	if status.SMTP != constants.IntegrationStatusConnected {
		t.Errorf("expected smtp connected, got %s", status.SMTP)
	}
}

func TestUsecase_CheckHealth_PingErrors(t *testing.T) {
	db := &mockPinger{err: errors.New("db ping failed")}
	rdb := &mockPinger{err: errors.New("redis ping failed")}
	s3 := &mockPinger{err: errors.New("s3 ping failed")}
	rmq := &mockPinger{err: errors.New("rmq ping failed")}
	smtp := &mockPinger{err: errors.New("smtp ping failed")}

	u := health.NewUsecase(db, rdb, s3, rmq, smtp)
	status := u.CheckHealth(context.Background())

	if status.Database != constants.IntegrationStatusDisconnected {
		t.Errorf("expected db disconnected on error, got %s", status.Database)
	}
	if status.Redis != constants.IntegrationStatusDisconnected {
		t.Errorf("expected redis disconnected on error, got %s", status.Redis)
	}
	if status.S3 != constants.IntegrationStatusDisconnected {
		t.Errorf("expected s3 disconnected on error, got %s", status.S3)
	}
	if status.RabbitMQ != constants.IntegrationStatusDisconnected {
		t.Errorf("expected rabbitmq disconnected on error, got %s", status.RabbitMQ)
	}
	if status.SMTP != constants.IntegrationStatusDisconnected {
		t.Errorf("expected smtp disconnected on error, got %s", status.SMTP)
	}
}
