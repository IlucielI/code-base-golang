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
	u := health.NewUsecase(nil)
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
	pingers := map[string]health.Pinger{
		"database": &mockPinger{err: nil},
		"redis":    &mockPinger{err: nil},
		"s3":       &mockPinger{err: nil},
		"rabbitmq": &mockPinger{err: nil},
		"smtp":     &mockPinger{err: nil},
	}

	u := health.NewUsecase(pingers)
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
	pingers := map[string]health.Pinger{
		"database": &mockPinger{err: errors.New("db ping failed")},
		"redis":    &mockPinger{err: errors.New("redis ping failed")},
		"s3":       &mockPinger{err: errors.New("s3 ping failed")},
		"rabbitmq": &mockPinger{err: errors.New("rmq ping failed")},
		"smtp":     &mockPinger{err: errors.New("smtp ping failed")},
	}

	u := health.NewUsecase(pingers)
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
