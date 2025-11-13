package providers

import (
	"context"
	"fmt"
	"net/smtp"

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

// NewSMTPConfigFromAppConfig creates SMTP config from app config
func NewSMTPConfigFromAppConfig(appConfig config.SMTPConfig) *SMTPConfig {
	return &SMTPConfig{
		Host:     appConfig.Host,
		Port:     appConfig.Port,
		Username: appConfig.Username,
		Password: appConfig.Password,
		From:     appConfig.From,
	}
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

// SMTPEmailProvider implements the SMTP email provider
type SMTPEmailProvider struct {
	config *SMTPConfig
}

// NewSMTPProvider creates a new SMTP provider
func NewSMTPProvider(config *SMTPConfig) EmailProviderInterface {
	return &SMTPEmailProvider{
		config: config,
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
	return smtp.SendMail(addr, auth, from, to, msg)
}

// ValidateConfig validates the provider configuration
func (p *SMTPEmailProvider) ValidateConfig() error {
	return p.config.Validate()
}

// GetProviderName returns the provider name
func (p *SMTPEmailProvider) GetProviderName() EmailProvider {
	return SMTPProvider
}
