package services

import (
	"context"

	"code-base-golang/internal/constants"
	"code-base-golang/internal/dtos"
	"code-base-golang/internal/pkg/ctxutil"
)

// CheckHealth inspects the connectivity of sub-systems (database, redis, s3 storage).
// The service is guaranteed non-nil by construction: cmd/api/main.go always injects a
// live service, and routes.NewRouter rejects a nil controller. Individual adapters may
// still be nil (e.g. the broker fails to connect outside production), so each is guarded
// and reported as disconnected rather than panicking.
func (s *Service) CheckHealth(ctx context.Context) dtos.HealthServices {
	dbStatus := constants.IntegrationStatusConnected
	if s.repo == nil || ctxutil.RunWithTimeout(ctx, 0, s.repo.PingDB) != nil {
		dbStatus = constants.IntegrationStatusDisconnected
	}

	redisStatus := constants.IntegrationStatusConnected
	if s.repo == nil || ctxutil.RunWithTimeout(ctx, 0, s.repo.PingRedis) != nil {
		redisStatus = constants.IntegrationStatusDisconnected
	}

	s3Status := constants.IntegrationStatusConnected
	if s.storage == nil || ctxutil.RunWithTimeout(ctx, 0, s.storage.Ping) != nil {
		s3Status = constants.IntegrationStatusDisconnected
	}

	rmqStatus := constants.IntegrationStatusConnected
	if s.publisher == nil || ctxutil.RunWithTimeout(ctx, 0, s.publisher.Ping) != nil {
		rmqStatus = constants.IntegrationStatusDisconnected
	}

	smtpStatus := constants.IntegrationStatusConnected
	if s.mailer == nil || ctxutil.RunWithTimeout(ctx, 0, s.mailer.Ping) != nil {
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
