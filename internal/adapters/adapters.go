package adapters

import (
	"context"
	"errors"
	"time"

	"code-base-golang/internal/adapters/database"
	"code-base-golang/internal/adapters/rabbitmq"
	"code-base-golang/internal/adapters/redis"
	"code-base-golang/internal/adapters/s3"
	"code-base-golang/internal/adapters/smtp"
	"code-base-golang/internal/config"
)

// Adapters groups all shared infrastructure adapters and runtime metadata.
type Adapters struct {
	Config    config.Config
	StartedAt time.Time
	DB        *database.Postgres
	Redis     *redis.Redis
	Storage   *s3.S3
	Publisher *rabbitmq.RabbitMQ
	Mailer    *smtp.SMTP
}

// New initializes and returns a container of all configured infrastructure adapters.
func New(
	cfg config.Config,
	db *database.Postgres,
	rdb *redis.Redis,
	storage *s3.S3,
	publisher *rabbitmq.RabbitMQ,
	mailer *smtp.SMTP,
) *Adapters {
	return &Adapters{
		Config:    cfg,
		StartedAt: time.Now(),
		DB:        db,
		Redis:     rdb,
		Storage:   storage,
		Publisher: publisher,
		Mailer:    mailer,
	}
}

// PingDatabase checks the health connectivity of the PostgreSQL database adapter.
func (a *Adapters) PingDatabase(ctx context.Context) error {
	if a.DB == nil {
		return errors.New("database adapter is not initialized")
	}
	return a.DB.Ping(ctx)
}

// PingRedis checks the health connectivity of the Redis adapter.
func (a *Adapters) PingRedis(ctx context.Context) error {
	if a.Redis == nil {
		return errors.New("redis adapter is not initialized")
	}
	return a.Redis.Ping(ctx)
}

// PingStorage checks the health connectivity of the S3 storage adapter.
func (a *Adapters) PingStorage(ctx context.Context) error {
	if a.Storage == nil {
		return errors.New("storage adapter is not initialized")
	}
	return a.Storage.Ping(ctx)
}

// PingPublisher checks the health connectivity of the RabbitMQ message publisher adapter.
func (a *Adapters) PingPublisher(ctx context.Context) error {
	if a.Publisher == nil {
		return errors.New("rabbitmq adapter is not initialized")
	}
	return a.Publisher.Ping(ctx)
}

// PingMailer checks the health connectivity of the SMTP mailer adapter.
func (a *Adapters) PingMailer(ctx context.Context) error {
	if a.Mailer == nil {
		return errors.New("smtp adapter is not initialized")
	}
	return a.Mailer.Ping(ctx)
}
