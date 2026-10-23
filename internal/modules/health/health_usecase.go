package health

import (
	"context"

	"code-base-golang/internal/constants"
	"code-base-golang/internal/pkg/ctxutil"
)

// Pinger defines the interface for sub-systems that can report their health status via Ping.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Usecase encapsulates business logic for health checks across system integrations.
type Usecase struct {
	pingers map[string]Pinger
}

// NewUsecase constructs a new health Usecase instance.
func NewUsecase(pingers map[string]Pinger) *Usecase {
	if pingers == nil {
		pingers = make(map[string]Pinger)
	}
	return &Usecase{
		pingers: pingers,
	}
}

// CheckHealth inspects the connectivity of sub-systems (database, redis, s3, rabbitmq, smtp).
func (u *Usecase) CheckHealth(ctx context.Context) Services {
	check := func(name string) string {
		p := u.pingers[name]
		if p == nil || ctxutil.RunWithTimeout(ctx, 0, p.Ping) != nil {
			return constants.IntegrationStatusDisconnected
		}
		return constants.IntegrationStatusConnected
	}

	return Services{
		Database: check("database"),
		Redis:    check("redis"),
		S3:       check("s3"),
		RabbitMQ: check("rabbitmq"),
		SMTP:     check("smtp"),
	}
}
