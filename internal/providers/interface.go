package providers

import (
	"context"
)

// EmailProvider represents different email service providers
type EmailProvider string

const (
	SMTPProvider EmailProvider = "smtp"
	// Future providers can be added here
	// SendGridProvider EmailProvider = "sendgrid"
	// AWSProvider      EmailProvider = "aws"
)

// EmailNotification represents an email to be sent
type EmailNotification struct {
	To      string
	Subject string
	Body    string
	From    string // Optional, will use default if empty
}

// EmailProviderInterface defines the contract for all email providers
type EmailProviderInterface interface {
	SendEmail(ctx context.Context, notification *EmailNotification) error
	ValidateConfig() error
	GetProviderName() EmailProvider
}

// ProviderConfig represents the configuration for any email provider
type ProviderConfig interface {
	GetProviderType() EmailProvider
	Validate() error
}
