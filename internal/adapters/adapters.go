package adapters

import (
	"code-base-golang/internal/adapters/database"
	"code-base-golang/internal/adapters/rabbitmq"
	"code-base-golang/internal/adapters/redis"
	"code-base-golang/internal/adapters/s3"
	"code-base-golang/internal/adapters/smtp"
	"code-base-golang/internal/config"
	"code-base-golang/internal/pkg/controller"
)

// Adapters groups all shared infrastructure adapters and application base controller.
type Adapters struct {
	Config    config.Config
	DB        *database.Postgres
	Redis     *redis.Redis
	Storage   *s3.S3
	Publisher *rabbitmq.RabbitMQ
	Mailer    *smtp.SMTP
	Base      controller.Base
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
		DB:        db,
		Redis:     rdb,
		Storage:   storage,
		Publisher: publisher,
		Mailer:    mailer,
		Base:      controller.NewBase(cfg),
	}
}
