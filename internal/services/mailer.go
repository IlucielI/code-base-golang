package services

import "context"

// EmailMessage represents the content and recipients of an outgoing email.
type EmailMessage struct {
	To       []string
	Cc       []string
	Bcc      []string
	Subject  string
	TextBody string
	HTMLBody string
	ReplyTo  string
}

// EmailSender defines the abstract contract for transactional email dispatchers.
// Concrete implementations (SMTP, AWS SES, SendGrid, Resend) can satisfy this interface.
type EmailSender interface {
	// Ping checks if the mail delivery service / SMTP host is reachable.
	Ping(ctx context.Context) error

	// Send dispatches an email message to the specified recipients.
	Send(ctx context.Context, msg EmailMessage) error
}
