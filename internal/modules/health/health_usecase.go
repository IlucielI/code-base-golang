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
	check := func(pingFn func(context.Context) error) string {
		if u.checker == nil || pingFn == nil || ctxutil.RunWithTimeout(ctx, 0, pingFn) != nil {
			return constants.IntegrationStatusDisconnected
		}
		return constants.IntegrationStatusConnected
	}

	if u.checker == nil {
		return Services{
			Database: constants.IntegrationStatusDisconnected,
			Redis:    constants.IntegrationStatusDisconnected,
			S3:       constants.IntegrationStatusDisconnected,
			RabbitMQ: constants.IntegrationStatusDisconnected,
			SMTP:     constants.IntegrationStatusDisconnected,
		}
	}

	return Services{
		Database: check(u.checker.PingDatabase),
		Redis:    check(u.checker.PingRedis),
		S3:       check(u.checker.PingStorage),
		RabbitMQ: check(u.checker.PingPublisher),
		SMTP:     check(u.checker.PingMailer),
	}
}
