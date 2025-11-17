package providers

import (
	"context"
	"fmt"
	"net/smtp"
	"sync/atomic"
	"time"

	"newsletter-service/internal/config"
	"newsletter-service/internal/providers/templates"
)

// SMTPConfig holds SMTP-specific configuration
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// GetProviderType returns the provider type
func (c *SMTPConfig) GetProviderType() EmailProvider {
	return SMTPProvider
}

// Validate validates the SMTP configuration
func (c *SMTPConfig) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("SMTP host is required")
	}
	if c.Port <= 0 {
		return fmt.Errorf("SMTP port must be positive")
	}
	if c.Username == "" {
		return fmt.Errorf("SMTP username is required")
	}
	if c.Password == "" {
		return fmt.Errorf("SMTP password is required")
	}
	return nil
}

// SMTPEmailProvider implements the enhanced SMTP email provider
type SMTPEmailProvider struct {
	name           string
	config         *SMTPConfig
	priority       int
	maxEmailsHour  int
	emailsSentHour int64
	lastHourReset  time.Time
	isHealthy      bool
	lastError      error
}

// NewSMTPProvider creates a new SMTP provider (legacy)
func NewSMTPProvider(config *SMTPConfig) EmailProviderInterface {
	return &SMTPEmailProvider{
		config:         config,
		emailsSentHour: 0,
		lastHourReset:  time.Now(),
		isHealthy:      true,
	}
}

// NewDynamicSMTPProvider creates a new SMTP provider from dynamic config
func NewDynamicSMTPProvider(name string, config *config.SMTPProviderConfig) EmailProviderInterface {
	return &SMTPEmailProvider{
		name:           name,
		config:         convertToSMTPConfig(config),
		priority:       config.Priority,
		maxEmailsHour:  config.MaxEmailsPerHour,
		emailsSentHour: 0,
		lastHourReset:  time.Now(),
		isHealthy:      true,
	}
}

// convertToSMTPConfig converts dynamic config to internal config
func convertToSMTPConfig(config *config.SMTPProviderConfig) *SMTPConfig {
	return &SMTPConfig{
		Host:     config.Host,
		Port:     config.Port,
		Username: config.Username,
		Password: config.Password,
		From:     config.From,
	}
}

// SendEmail sends an email using SMTP
func (p *SMTPEmailProvider) SendEmail(ctx context.Context, notification *EmailNotification) error {
	// Generate HTML email using template
	htmlBody, err := templates.GenerateEmailHTML(notification.Subject, notification.Body)
	if err != nil {
		return fmt.Errorf("failed to generate email template: %w", err)
	}

	auth := smtp.PlainAuth("", p.config.Username, p.config.Password, p.config.Host)

	to := []string{notification.To}

	// Determine from address
	from := notification.From
	if from == "" {
		from = p.config.From
	}
	if from == "" {
		from = p.config.Username
	}

	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s\r\n",
		from,
		notification.To,
		notification.Subject,
		htmlBody,
	))

	addr := fmt.Sprintf("%s:%d", p.config.Host, p.config.Port)
	err = smtp.SendMail(addr, auth, from, to, msg)

	// Update statistics
	if err != nil {
		p.isHealthy = false
		p.lastError = err
	} else {
		p.isHealthy = true
		p.lastError = nil
		atomic.AddInt64(&p.emailsSentHour, 1)
	}

	return err
}

// SendBulkEmail sends bulk emails (SMTP doesn't support true bulk, so send individually)
func (p *SMTPEmailProvider) SendBulkEmail(ctx context.Context, notification *BulkEmailNotification) error {
	var lastError error
	successCount := 0

	for _, recipient := range notification.To {
		singleNotification := &EmailNotification{
			To:      recipient,
			Subject: notification.Subject,
			Body:    notification.Body,
			From:    notification.From,
		}

		if err := p.SendEmail(ctx, singleNotification); err != nil {
			lastError = err
		} else {
			successCount++
		}
	}

	// Consider successful if at least 50% succeeded
	if successCount < len(notification.To)/2 {
		return fmt.Errorf("bulk email failed: %d/%d succeeded, last error: %v", successCount, len(notification.To), lastError)
	}

	return nil
}

// SupportsBulk returns false as SMTP doesn't support true bulk operations
func (p *SMTPEmailProvider) SupportsBulk() bool {
	return false
}

// GetLimits returns provider limitations
func (p *SMTPEmailProvider) GetLimits() ProviderLimits {
	maxEmails := 1000 // Default
	if p.maxEmailsHour > 0 {
		maxEmails = p.maxEmailsHour
	}

	return ProviderLimits{
		MaxEmailsPerHour: maxEmails,
		MaxBatchSize:     1, // SMTP sends individually
		SupportsBulk:     false,
	}
}

// GetStats returns current provider statistics
func (p *SMTPEmailProvider) GetStats() ProviderStats {
	// Reset counter if more than an hour has passed
	if time.Since(p.lastHourReset) > time.Hour {
		atomic.StoreInt64(&p.emailsSentHour, 0)
		p.lastHourReset = time.Now()
	}

	emailsSent := int(atomic.LoadInt64(&p.emailsSentHour))
	currentLoad := (emailsSent * 100) / p.GetLimits().MaxEmailsPerHour

	return ProviderStats{
		EmailsSentLastHour: emailsSent,
		CurrentLoad:        currentLoad,
		IsHealthy:          p.isHealthy,
		LastError:          p.lastError,
	}
}

// GetPriority returns provider priority (lower = higher priority)
func (p *SMTPEmailProvider) GetPriority() int {
	return 1 // Default priority for SMTP
}

// IsEnabled returns true if provider is enabled
func (p *SMTPEmailProvider) IsEnabled() bool {
	return true // SMTP is always enabled if configured
}

// ValidateConfig validates the provider configuration
func (p *SMTPEmailProvider) ValidateConfig() error {
	return p.config.Validate()
}

// GetProviderName returns the provider name
func (p *SMTPEmailProvider) GetProviderName() string {
	if p.name != "" {
		return p.name
	}
	return "smtp"
}

// GetProviderType returns the provider type
func (p *SMTPEmailProvider) GetProviderType() EmailProvider {
	return SMTPProviderType
}
