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

// MailtrapProvider implements Mailtrap bulk API email provider
type MailtrapProvider struct {
	config         *config.MailtrapConfig
	emailsSentHour int64
	lastHourReset  time.Time
	isHealthy      bool
	lastError      error
}

// MailtrapEmail represents the Mailtrap API payload structure
type MailtrapEmail struct {
	From     MailtrapContact   `json:"from"`
	To       []MailtrapContact `json:"to"`
	Subject  string            `json:"subject"`
	Text     string            `json:"text,omitempty"`
	HTML     string            `json:"html,omitempty"`
	Category string            `json:"category,omitempty"`
}

type MailtrapContact struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

// NewMailtrapProvider creates a new Mailtrap provider
func NewMailtrapProvider(config *config.MailtrapConfig) EmailProviderInterface {
	return &MailtrapProvider{
		config:         config,
		emailsSentHour: 0,
		lastHourReset:  time.Now(),
		isHealthy:      true,
	}
}

// SendEmail sends a single email via Mailtrap API
func (p *MailtrapProvider) SendEmail(ctx context.Context, notification *EmailNotification) error {
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

	// Prepare Mailtrap payload
	email := MailtrapEmail{
		From: MailtrapContact{
			Email: from,
			Name:  "Newsletter Service",
		},
		To: []MailtrapContact{
			{Email: notification.To},
		},
		Subject:  notification.Subject,
		Text:     notification.Body,
		HTML:     htmlBody,
		Category: "Newsletter",
	}

	err = p.sendToMailtrap(ctx, email)

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

// SendBulkEmail sends bulk emails via Mailtrap API
func (p *MailtrapProvider) SendBulkEmail(ctx context.Context, notification *BulkEmailNotification) error {
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

	// Convert recipients to Mailtrap format
	recipients := make([]MailtrapContact, len(notification.To))
	for i, email := range notification.To {
		recipients[i] = MailtrapContact{Email: email}
	}

	// Prepare Mailtrap bulk payload
	email := MailtrapEmail{
		From: MailtrapContact{
			Email: from,
			Name:  "Newsletter Service",
		},
		To:       recipients,
		Subject:  notification.Subject,
		Text:     notification.Body,
		HTML:     htmlBody,
		Category: "Newsletter",
	}

	err = p.sendToMailtrap(ctx, email)

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

// sendToMailtrap handles the HTTP request to Mailtrap API
func (p *MailtrapProvider) sendToMailtrap(ctx context.Context, email MailtrapEmail) error {
	jsonPayload, err := json.Marshal(email)
	if err != nil {
		return fmt.Errorf("failed to marshal Mailtrap payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.config.BulkEndpoint, bytes.NewReader(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create Mailtrap request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.config.APIToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Mailtrap request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Mailtrap API returned status %d", resp.StatusCode)
	}

	return nil
}

// SupportsBulk returns true as Mailtrap supports bulk operations
func (p *MailtrapProvider) SupportsBulk() bool {
	return p.config.BulkEnabled
}

// GetLimits returns Mailtrap provider limitations
func (p *MailtrapProvider) GetLimits() ProviderLimits {
	return ProviderLimits{
		MaxEmailsPerHour: p.config.MaxEmailsPerHour,
		MaxBatchSize:     p.config.MaxBatchSize,
		SupportsBulk:     p.config.BulkEnabled,
	}
}

// GetStats returns current Mailtrap provider statistics
func (p *MailtrapProvider) GetStats() ProviderStats {
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
func (p *MailtrapProvider) GetProviderName() string {
	return "mailtrap"
}

// GetProviderType returns the provider type
func (p *MailtrapProvider) GetProviderType() EmailProvider {
	return APIProviderType
}

// GetPriority returns provider priority
func (p *MailtrapProvider) GetPriority() int {
	return p.config.Priority
}

// IsEnabled returns true if provider is enabled
func (p *MailtrapProvider) IsEnabled() bool {
	return p.config.Enabled
}

// ValidateConfig validates the Mailtrap configuration
func (p *MailtrapProvider) ValidateConfig() error {
	if p.config.APIToken == "" {
		return fmt.Errorf("Mailtrap API token is required")
	}
	if p.config.BulkEndpoint == "" {
		return fmt.Errorf("Mailtrap bulk endpoint is required")
	}
	if p.config.From == "" {
		return fmt.Errorf("Mailtrap from address is required")
	}
	return nil
}
