package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"newsletter-service/internal/config"
	"newsletter-service/internal/providers/templates"
)

// SendGridProvider implements SendGrid API email provider
type SendGridProvider struct {
	config         *config.SendGridConfig
	emailsSentHour int64
	lastHourReset  time.Time
	isHealthy      bool
	lastError      error
}

// SendGridEmail represents the SendGrid API payload structure
type SendGridEmail struct {
	Personalizations []SendGridPersonalization `json:"personalizations"`
	From             SendGridContact           `json:"from"`
	Subject          string                    `json:"subject"`
	Content          []SendGridContent         `json:"content"`
}

type SendGridPersonalization struct {
	To []SendGridContact `json:"to"`
}

type SendGridContact struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type SendGridContent struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// NewSendGridProvider creates a new SendGrid provider
func NewSendGridProvider(config *config.SendGridConfig) EmailProviderInterface {
	return &SendGridProvider{
		config:         config,
		emailsSentHour: 0,
		lastHourReset:  time.Now(),
		isHealthy:      true,
	}
}

// SendEmail sends a single email via SendGrid API
func (p *SendGridProvider) SendEmail(ctx context.Context, notification *EmailNotification) error {
	// Generate HTML email using template
	htmlBody, err := templates.GenerateEmailHTML(notification.Subject, notification.Body)
	if err != nil {
		return fmt.Errorf("failed to generate email template: %w", err)
	}

	// Determine from address
	from := notification.From
	if from == "" {
		from = p.config.From
	}

	// Prepare SendGrid payload
	email := SendGridEmail{
		Personalizations: []SendGridPersonalization{
			{
				To: []SendGridContact{
					{Email: notification.To},
				},
			},
		},
		From:    SendGridContact{Email: from},
		Subject: notification.Subject,
		Content: []SendGridContent{
			{Type: "text/plain", Value: notification.Body},
			{Type: "text/html", Value: htmlBody},
		},
	}

	err = p.sendToSendGrid(ctx, email)

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

// SendBulkEmail sends bulk emails via SendGrid API
func (p *SendGridProvider) SendBulkEmail(ctx context.Context, notification *BulkEmailNotification) error {
	// Generate HTML email using template
	htmlBody, err := templates.GenerateEmailHTML(notification.Subject, notification.Body)
	if err != nil {
		return fmt.Errorf("failed to generate email template: %w", err)
	}

	// Determine from address
	from := notification.From
	if from == "" {
		from = p.config.From
	}

	// Convert recipients to SendGrid format
	recipients := make([]SendGridContact, len(notification.To))
	for i, email := range notification.To {
		recipients[i] = SendGridContact{Email: email}
	}

	// Prepare SendGrid bulk payload
	email := SendGridEmail{
		Personalizations: []SendGridPersonalization{
			{To: recipients},
		},
		From:    SendGridContact{Email: from},
		Subject: notification.Subject,
		Content: []SendGridContent{
			{Type: "text/plain", Value: notification.Body},
			{Type: "text/html", Value: htmlBody},
		},
	}

	err = p.sendToSendGrid(ctx, email)

	// Update statistics
	if err != nil {
		p.isHealthy = false
		p.lastError = err
	} else {
		p.isHealthy = true
		p.lastError = nil
		atomic.AddInt64(&p.emailsSentHour, int64(len(notification.To)))
	}

	return err
}

// sendToSendGrid handles the HTTP request to SendGrid API
func (p *SendGridProvider) sendToSendGrid(ctx context.Context, email SendGridEmail) error {
	jsonPayload, err := json.Marshal(email)
	if err != nil {
		return fmt.Errorf("failed to marshal SendGrid payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.sendgrid.com/v3/mail/send", bytes.NewReader(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create SendGrid request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send SendGrid request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("SendGrid API returned status %d", resp.StatusCode)
	}

	return nil
}

// SupportsBulk returns true as SendGrid supports bulk operations
func (p *SendGridProvider) SupportsBulk() bool {
	return p.config.BulkEnabled
}

// GetLimits returns SendGrid provider limitations
func (p *SendGridProvider) GetLimits() ProviderLimits {
	return ProviderLimits{
		MaxEmailsPerHour: p.config.MaxEmailsPerHour,
		MaxBatchSize:     p.config.MaxBatchSize,
		SupportsBulk:     p.config.BulkEnabled,
	}
}

// GetStats returns current SendGrid provider statistics
func (p *SendGridProvider) GetStats() ProviderStats {
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

// GetProviderName returns the provider name
func (p *SendGridProvider) GetProviderName() string {
	return "sendgrid"
}

// GetProviderType returns the provider type
func (p *SendGridProvider) GetProviderType() EmailProvider {
	return APIProviderType
}

// GetPriority returns provider priority
func (p *SendGridProvider) GetPriority() int {
	return p.config.Priority
}

// IsEnabled returns true if provider is enabled
func (p *SendGridProvider) IsEnabled() bool {
	return p.config.Enabled
}

// ValidateConfig validates the SendGrid configuration
func (p *SendGridProvider) ValidateConfig() error {
	if p.config.APIKey == "" {
		return fmt.Errorf("SendGrid API key is required")
	}
	if p.config.From == "" {
		return fmt.Errorf("SendGrid from address is required")
	}
	return nil
}
