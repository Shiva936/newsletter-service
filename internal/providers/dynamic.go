package providers

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"newsletter-service/internal/config"
)

// NewDynamicAPIProvider creates an API provider from dynamic configuration
func NewDynamicAPIProvider(name string, cfg *config.APIProviderConfig) EmailProviderInterface {
	// Return a generic API provider for now
	// In the future, we could add specific implementations based on provider type
	return &GenericAPIProvider{
		apiKey:         cfg.Token,
		endpoint:       cfg.Endpoint,
		name:           name,
		priority:       cfg.Priority,
		maxEmailsHour:  cfg.MaxEmailsPerHour,
		bulkEnabled:    cfg.BulkEnabled,
		maxBatchSize:   cfg.MaxBatchSize,
		isHealthy:      true,
		emailsSentHour: 0,
	}
}

// GenericAPIProvider is a flexible API-based email provider
type GenericAPIProvider struct {
	apiKey         string
	endpoint       string
	name           string
	priority       int
	maxEmailsHour  int
	bulkEnabled    bool
	maxBatchSize   int
	isHealthy      bool
	emailsSentHour int64
	lastError      error
	lastHourReset  time.Time
}

// Implement EmailProviderInterface methods for GenericAPIProvider
func (p *GenericAPIProvider) SendEmail(ctx context.Context, notification *EmailNotification) error {
	// Generic API implementation would go here
	// For now, just return a placeholder
	atomic.AddInt64(&p.emailsSentHour, 1)
	return nil
}

func (p *GenericAPIProvider) SendBulkEmail(ctx context.Context, notification *BulkEmailNotification) error {
	// Generic bulk API implementation
	atomic.AddInt64(&p.emailsSentHour, int64(len(notification.To)))
	return nil
}

func (p *GenericAPIProvider) SupportsBulk() bool {
	return p.bulkEnabled
}

func (p *GenericAPIProvider) GetLimits() ProviderLimits {
	return ProviderLimits{
		MaxEmailsPerHour: p.maxEmailsHour,
		MaxBatchSize:     p.maxBatchSize,
		SupportsBulk:     p.bulkEnabled,
	}
}

func (p *GenericAPIProvider) GetStats() ProviderStats {
	// Reset counter if more than an hour has passed
	if time.Since(p.lastHourReset) > time.Hour {
		atomic.StoreInt64(&p.emailsSentHour, 0)
		p.lastHourReset = time.Now()
	}

	emailsSent := int(atomic.LoadInt64(&p.emailsSentHour))
	currentLoad := 0
	if p.maxEmailsHour > 0 {
		currentLoad = (emailsSent * 100) / p.maxEmailsHour
	}

	return ProviderStats{
		EmailsSentLastHour: emailsSent,
		CurrentLoad:        currentLoad,
		IsHealthy:          p.isHealthy,
		LastError:          p.lastError,
	}
}

func (p *GenericAPIProvider) GetProviderName() string {
	return p.name
}

func (p *GenericAPIProvider) GetProviderType() EmailProvider {
	return APIProviderType
}

func (p *GenericAPIProvider) GetPriority() int {
	return p.priority
}

func (p *GenericAPIProvider) IsEnabled() bool {
	return true
}

func (p *GenericAPIProvider) ValidateConfig() error {
	if p.apiKey == "" {
		return fmt.Errorf("API key is required")
	}
	return nil
}
