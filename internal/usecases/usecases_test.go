package usecases

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"code-base-golang/internal/config"
	"code-base-golang/internal/constants"
)

// mockStorage implements FileStorage for unit testing without S3 or MinIO.
type mockStorage struct {
	pingErr error
}

func (m *mockStorage) Ping(ctx context.Context) error {
	return m.pingErr
}

func (m *mockStorage) Upload(ctx context.Context, bucketName, objectName string, reader io.Reader, size int64, contentType string) error {
	return nil
}

func (m *mockStorage) Download(ctx context.Context, bucketName, objectName string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("mock content")), nil
}

func (m *mockStorage) Delete(ctx context.Context, bucketName, objectName string) error {
	return nil
}

func (m *mockStorage) PresignGetObject(ctx context.Context, bucketName, objectName string, expiry time.Duration) (string, error) {
	return "https://mock.storage/get", nil
}

func (m *mockStorage) PresignPutObject(ctx context.Context, bucketName, objectName string, expiry time.Duration) (string, error) {
	return "https://mock.storage/put", nil
}

func TestUsecases_FileStorage_Mock(t *testing.T) {
	mock := &mockStorage{}
	uc := New(config.Config{}, nil, mock)

	if uc.Storage() == nil {
		t.Fatal("expected storage to be non-nil")
	}

	// Test SetStorage override
	mock2 := &mockStorage{pingErr: errors.New("storage unreachable")}
	uc.SetStorage(mock2)

	health := uc.CheckHealth(context.Background())
	if health.S3 != constants.IntegrationStatusDisconnected {
		t.Errorf("expected storage disconnected on error, got %s", health.S3)
	}

	// Test healthy storage
	uc.SetStorage(&mockStorage{pingErr: nil})
	health = uc.CheckHealth(context.Background())
	if health.S3 != constants.IntegrationStatusConnected {
		t.Errorf("expected storage connected, got %s", health.S3)
	}
}

// mockPublisher implements EventPublisher for unit testing without RabbitMQ or NATS.
type mockPublisher struct {
	pingErr    error
	publishErr error
	published  []string
}

func (m *mockPublisher) Ping(ctx context.Context) error {
	return m.pingErr
}

func (m *mockPublisher) Publish(ctx context.Context, topic string, payload any) error {
	if m.publishErr != nil {
		return m.publishErr
	}
	m.published = append(m.published, topic)
	return nil
}

func (m *mockPublisher) MustPublish(ctx context.Context, topic string, payload any) {
	if err := m.Publish(ctx, topic, payload); err != nil {
		panic(err)
	}
}

func TestUsecases_EventPublisher_Mock(t *testing.T) {
	pub := &mockPublisher{}
	uc := New(config.Config{}, nil, nil, pub)

	if uc.Publisher() == nil {
		t.Fatal("expected publisher to be non-nil")
	}

	// Test publishing
	err := uc.Publisher().Publish(context.Background(), "user.registered", map[string]string{"id": "123"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(pub.published) != 1 || pub.published[0] != "user.registered" {
		t.Fatalf("unexpected published topics: %v", pub.published)
	}

	// Test health check with broker error
	uc.SetPublisher(&mockPublisher{pingErr: errors.New("broker down")})
	health := uc.CheckHealth(context.Background())
	if health.RabbitMQ != constants.IntegrationStatusDisconnected {
		t.Errorf("expected rabbitmq disconnected on ping error, got %s", health.RabbitMQ)
	}
}

// mockMailer implements EmailSender for unit testing without SMTP server.
type mockMailer struct {
	pingErr error
	sendErr error
	sent    []EmailMessage
}

func (m *mockMailer) Ping(ctx context.Context) error {
	return m.pingErr
}

func (m *mockMailer) Send(ctx context.Context, msg EmailMessage) error {
	if m.sendErr != nil {
		return m.sendErr
	}
	m.sent = append(m.sent, msg)
	return nil
}

func TestUsecases_EmailSender_Mock(t *testing.T) {
	mailer := &mockMailer{}
	uc := New(config.Config{}, nil, nil).WithMailer(mailer)

	if uc.Mailer() == nil {
		t.Fatal("expected mailer to be non-nil")
	}

	// Test sending
	err := uc.Mailer().Send(context.Background(), EmailMessage{
		To:      []string{"test@example.com"},
		Subject: "Test Subject",
	})
	if err != nil {
		t.Fatalf("expected nil error sending email, got %v", err)
	}
	if len(mailer.sent) != 1 || mailer.sent[0].Subject != "Test Subject" {
		t.Fatalf("unexpected sent email: %v", mailer.sent)
	}

	// Test health check
	health := uc.CheckHealth(context.Background())
	if health.SMTP != constants.IntegrationStatusConnected {
		t.Errorf("expected smtp connected, got %s", health.SMTP)
	}

	// Test health check with SMTP error
	uc.SetMailer(&mockMailer{pingErr: errors.New("smtp down")})
	health = uc.CheckHealth(context.Background())
	if health.SMTP != constants.IntegrationStatusDisconnected {
		t.Errorf("expected smtp disconnected on error, got %s", health.SMTP)
	}
}
