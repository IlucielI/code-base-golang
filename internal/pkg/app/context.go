package app

import (
	"code-base-golang/internal/adapters/database"
	"code-base-golang/internal/adapters/rabbitmq"
	"code-base-golang/internal/adapters/redis"
	"code-base-golang/internal/adapters/s3"
	"code-base-golang/internal/adapters/smtp"
	"code-base-golang/internal/config"
	"code-base-golang/internal/pkg/controller"
)

// Context encapsulates application configuration and shared infrastructure adapters.
type Context struct {
	Config    config.Config
	DB        *database.Postgres
	Redis     *redis.Redis
	Storage   *s3.S3
	Publisher *rabbitmq.RabbitMQ
	Mailer    *smtp.SMTP
	Base      controller.Base
}

// NewContext creates a new application infrastructure Context.
func NewContext(
	cfg config.Config,
	db *database.Postgres,
	rdb *redis.Redis,
	storage *s3.S3,
	publisher *rabbitmq.RabbitMQ,
	mailer *smtp.SMTP,
) *Context {
	return &Context{
		Config:    cfg,
		DB:        db,
		Redis:     rdb,
		Storage:   storage,
		Publisher: publisher,
		Mailer:    mailer,
		Base:      controller.NewBase(cfg),
	}
}
