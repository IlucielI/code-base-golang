package usecases

import (
	"context"

	"code-base-golang/internal/constants"
	"code-base-golang/internal/dtos"
	"code-base-golang/internal/pkg/ctxutil"
)

// CheckHealth inspects the connectivity of sub-systems (database, redis, s3 storage, rabbitmq, smtp).
// The usecases container is guaranteed non-nil by construction: cmd/api/main.go always injects a
// live container, and routes.NewRouter rejects a nil controller. Individual adapters may
// still be nil (e.g. the broker fails to connect outside production), so each is guarded
// and reported as disconnected rather than panicking.
func (u *Usecases) CheckHealth(ctx context.Context) dtos.HealthServices {
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
