package health

import (
	"context"

	"code-base-golang/internal/constants"
	"code-base-golang/internal/pkg/ctxutil"
)

// HealthChecker defines the diagnostic ping interface required by the health usecase.
type HealthChecker interface {
	PingDatabase(ctx context.Context) error
	PingRedis(ctx context.Context) error
	PingStorage(ctx context.Context) error
	PingPublisher(ctx context.Context) error
	PingMailer(ctx context.Context) error
}

// Usecase encapsulates business logic for health checks across system integrations.
type Usecase struct {
	checker HealthChecker
}

// NewUsecase constructs a new health Usecase instance.
func NewUsecase(checker HealthChecker) *Usecase {
	return &Usecase{
		checker: checker,
	}
}

// CheckHealth inspects the connectivity of sub-systems (database, redis, s3, rabbitmq, smtp).
func (u *Usecase) CheckHealth(ctx context.Context) Services {
	return Services{
		Database: pingStatus(ctx, u.checker.PingDatabase),
		Redis:    pingStatus(ctx, u.checker.PingRedis),
		S3:       pingStatus(ctx, u.checker.PingStorage),
		RabbitMQ: pingStatus(ctx, u.checker.PingPublisher),
		SMTP:     pingStatus(ctx, u.checker.PingMailer),
	}
}

// pingStatus checks whether a ping operation succeeds within the default timeout.
func pingStatus(ctx context.Context, pingFn func(context.Context) error) string {
	if ctxutil.RunWithTimeout(ctx, 0, pingFn) != nil {
		return constants.IntegrationStatusDisconnected
	}
	return constants.IntegrationStatusConnected
}
