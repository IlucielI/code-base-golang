package services

import (
	"context"

	"code-base-golang/internal/config"
	"code-base-golang/internal/constants"
	"code-base-golang/internal/repositories"
)

// Service is the unified application service container.
type Service struct {
	cfg       config.Config
	storage   FileStorage
	repo      *repositories.Repositories
	publisher EventPublisher
	mailer    EmailSender
}

// New creates a new unified service container.
func New(cfg config.Config, repo *repositories.Repositories, storage FileStorage, publisher ...EventPublisher) *Service {
	var pub EventPublisher
	if len(publisher) > 0 {
		pub = publisher[0]
	}
	return &Service{
		cfg:       cfg,
		storage:   storage,
		repo:      repo,
		publisher: pub,
	}
}

// SetStorage allows injecting or overriding storage adapter (e.g. for testing or switching provider).
func (s *Service) SetStorage(storage FileStorage) {
	s.storage = storage
}

// Storage returns the underlying file storage provider.
func (s *Service) Storage() FileStorage {
	if s == nil {
		return nil
	}
	return s.storage
}

// SetRepositories allows injecting or overriding repositories (e.g. for testing).
func (s *Service) SetRepositories(repo *repositories.Repositories) {
	s.repo = repo
}

// Repositories returns the underlying repositories container.
func (s *Service) Repositories() *repositories.Repositories {
	if s == nil {
		return nil
	}
	return s.repo
}

// SetPublisher allows injecting or overriding event publisher adapter (e.g. for testing or switching broker).
func (s *Service) SetPublisher(pub EventPublisher) {
	s.publisher = pub
}

// Publisher returns the underlying event publisher adapter.
func (s *Service) Publisher() EventPublisher {
	if s == nil {
		return nil
	}
	return s.publisher
}

// SetMailer allows injecting or overriding email sender adapter (e.g. for testing or switching mail provider).
func (s *Service) SetMailer(mailer EmailSender) {
	s.mailer = mailer
}

// WithMailer fluently sets the email sender adapter.
func (s *Service) WithMailer(mailer EmailSender) *Service {
	s.mailer = mailer
	return s
}

// Mailer returns the underlying email sender adapter.
func (s *Service) Mailer() EmailSender {
	if s == nil {
		return nil
	}
	return s.mailer
}

// wrapError wraps unknown or system errors into structured AppError.
func (s *Service) wrapError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	return constants.ErrInternalServerError.Wrap(err)
}
