package health

import (
	"context"

	"code-base-golang/internal/constants"
	"code-base-golang/internal/dtos"
	"code-base-golang/internal/pkg/ctxutil"
	"code-base-golang/internal/repositories"
)

// Pinger defines the interface for sub-systems that can report their health status via Ping.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Usecase encapsulates business logic for health checks across system integrations.
type Usecase struct {
	repo      *repositories.Repositories
	storage   Pinger
	publisher Pinger
	mailer    Pinger
}

// NewUsecase constructs a new health Usecase instance.
func NewUsecase(repo *repositories.Repositories, storage Pinger, publisher Pinger, mailer Pinger) *Usecase {
	return &Usecase{
		repo:      repo,
		storage:   storage,
		publisher: publisher,
		mailer:    mailer,
	}
}

// CheckHealth inspects the connectivity of sub-systems (database, redis, s3, rabbitmq, smtp).
func (u *Usecase) CheckHealth(ctx context.Context) dtos.HealthServices {
	dbStatus := constants.IntegrationStatusConnected
	if u.repo == nil || ctxutil.RunWithTimeout(ctx, 0, u.repo.PingDB) != nil {
		dbStatus = constants.IntegrationStatusDisconnected
	}

	redisStatus := constants.IntegrationStatusConnected
	if u.repo == nil || ctxutil.RunWithTimeout(ctx, 0, u.repo.PingRedis) != nil {
		redisStatus = constants.IntegrationStatusDisconnected
	}

	s3Status := constants.IntegrationStatusConnected
	if u.storage == nil || ctxutil.RunWithTimeout(ctx, 0, u.storage.Ping) != nil {
		s3Status = constants.IntegrationStatusDisconnected
	}

	rmqStatus := constants.IntegrationStatusConnected
	if u.publisher == nil || ctxutil.RunWithTimeout(ctx, 0, u.publisher.Ping) != nil {
		rmqStatus = constants.IntegrationStatusDisconnected
	}

	smtpStatus := constants.IntegrationStatusConnected
	if u.mailer == nil || ctxutil.RunWithTimeout(ctx, 0, u.mailer.Ping) != nil {
		smtpStatus = constants.IntegrationStatusDisconnected
	}

	return dtos.HealthServices{
		Database: dbStatus,
		Redis:    redisStatus,
		S3:       s3Status,
		RabbitMQ: rmqStatus,
		SMTP:     smtpStatus,
	}
}
