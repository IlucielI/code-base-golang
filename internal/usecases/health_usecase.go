package usecases

import (
	"context"

	"code-base-golang/internal/constants"
	"code-base-golang/internal/dtos"
)

// CheckHealth inspects the connectivity of sub-systems (database, redis, s3 storage, rabbitmq, smtp).
// The usecases container is guaranteed non-nil by construction: cmd/api/main.go always injects a
// live container, and routes.NewRouter rejects a nil controller. Individual adapters may
// still be nil (e.g. the broker fails to connect outside production), so each is guarded
// and reported as disconnected rather than panicking.
func (u *Usecases) CheckHealth(ctx context.Context) dtos.HealthServices {
	if u.Health == nil {
		return dtos.HealthServices{
			Database: constants.IntegrationStatusDisconnected,
			Redis:    constants.IntegrationStatusDisconnected,
			S3:       constants.IntegrationStatusDisconnected,
			RabbitMQ: constants.IntegrationStatusDisconnected,
			SMTP:     constants.IntegrationStatusDisconnected,
		}
	}
	return u.Health.CheckHealth(ctx)
}
