package health

import (
	"context"
	"errors"
	"testing"

	"code-base-golang/internal/constants"
)

type mockChecker struct {
	dbErr, redisErr, s3Err, rmqErr, smtpErr error
}

func (m *mockChecker) PingDatabase(ctx context.Context) error  { return m.dbErr }
func (m *mockChecker) PingRedis(ctx context.Context) error     { return m.redisErr }
func (m *mockChecker) PingStorage(ctx context.Context) error   { return m.s3Err }
func (m *mockChecker) PingPublisher(ctx context.Context) error { return m.rmqErr }
func (m *mockChecker) PingMailer(ctx context.Context) error    { return m.smtpErr }

func TestUsecase_CheckHealth_NilChecker(t *testing.T) {
	u := NewUsecase(nil)
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
	checker := &mockChecker{}
	u := NewUsecase(checker)
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
	checker := &mockChecker{
		dbErr:    errors.New("db ping failed"),
		redisErr: errors.New("redis ping failed"),
		s3Err:    errors.New("s3 ping failed"),
		rmqErr:   errors.New("rmq ping failed"),
		smtpErr:  errors.New("smtp ping failed"),
	}

	u := NewUsecase(checker)
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
