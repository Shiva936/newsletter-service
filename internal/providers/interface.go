package providers

import (
	"context"
)

// EmailProvider represents different email service provider types
type EmailProvider string

const (
	SMTPProviderType EmailProvider = "smtp"
	APIProviderType  EmailProvider = "api"

	// Legacy constants for backward compatibility
	SMTPProvider EmailProvider = "smtp"
)

// EmailNotification represents an email to be sent
type EmailNotification struct {
	To      string
	Subject string
	Body    string
	From    string // Optional, will use default if empty
}

// BulkEmailNotification represents a bulk email to be sent
type BulkEmailNotification struct {
	To      []string
	Subject string
	Body    string
	From    string // Optional, will use default if empty
}

// ProviderLimits represents provider limitations and capabilities
type ProviderLimits struct {
	MaxEmailsPerHour int
	MaxBatchSize     int
	SupportsBulk     bool
}

// ProviderStats represents real-time provider statistics
type ProviderStats struct {
	EmailsSentLastHour int
	CurrentLoad        int // Percentage 0-100
	IsHealthy          bool
	LastError          error
}

// EmailProviderInterface defines the contract for all email providers
type EmailProviderInterface interface {
	SendEmail(ctx context.Context, notification *EmailNotification) error
	SendBulkEmail(ctx context.Context, notification *BulkEmailNotification) error
	SupportsBulk() bool
	GetLimits() ProviderLimits
	GetStats() ProviderStats
	GetProviderType() EmailProvider
	GetProviderName() string
	GetPriority() int
	IsEnabled() bool
	ValidateConfig() error
}

// BatchManager handles batching logic for providers that don't support true bulk
type BatchManager interface {
	AddToBatch(email *EmailNotification) error
	ProcessBatch(ctx context.Context) error
	GetBatchSize() int
	IsBatchReady() bool
}

// ProviderConfig represents the configuration for any email provider
type ProviderConfig interface {
	GetProviderType() EmailProvider
	Validate() error
}
