package usecases

import (
	"context"

	"code-base-golang/internal/config"
	"code-base-golang/internal/constants"
	"code-base-golang/internal/repositories"
)

// Usecases is the unified application use cases container.
type Usecases struct {
	cfg       config.Config
	storage   FileStorage
	repo      *repositories.Repositories
	publisher EventPublisher
	mailer    EmailSender
}

// New creates a new unified use cases container.
func New(cfg config.Config, repo *repositories.Repositories, storage FileStorage, publisher ...EventPublisher) *Usecases {
	var pub EventPublisher
	if len(publisher) > 0 {
		pub = publisher[0]
	}
	return &Usecases{
		cfg:       cfg,
		storage:   storage,
		repo:      repo,
		publisher: pub,
	}
}

// SetStorage allows injecting or overriding storage adapter (e.g. for testing or switching provider).
func (u *Usecases) SetStorage(storage FileStorage) {
	u.storage = storage
}

// Storage returns the underlying file storage provider.
func (u *Usecases) Storage() FileStorage {
	return u.storage
}

// SetRepositories allows injecting or overriding repositories (e.g. for testing).
func (u *Usecases) SetRepositories(repo *repositories.Repositories) {
	u.repo = repo
}

// Repositories returns the underlying repositories container.
func (u *Usecases) Repositories() *repositories.Repositories {
	return u.repo
}

// SetPublisher allows injecting or overriding event publisher adapter (e.g. for testing or switching broker).
func (u *Usecases) SetPublisher(pub EventPublisher) {
	u.publisher = pub
}

// Publisher returns the underlying event publisher adapter.
func (u *Usecases) Publisher() EventPublisher {
	return u.publisher
}

// SetMailer allows injecting or overriding email sender adapter (e.g. for testing or switching mail provider).
func (u *Usecases) SetMailer(mailer EmailSender) {
	u.mailer = mailer
}

// WithMailer fluently sets the email sender adapter.
func (u *Usecases) WithMailer(mailer EmailSender) *Usecases {
	u.mailer = mailer
	return u
}

// Mailer returns the underlying email sender adapter.
func (u *Usecases) Mailer() EmailSender {
	return u.mailer
}

// WrapError wraps unknown or system errors into structured AppError.
func (u *Usecases) WrapError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	return constants.ErrInternalServerError.Wrap(err)
}
