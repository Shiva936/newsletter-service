package providers

import (
	"fmt"
	"sync"

	"newsletter-service/internal/config"
)

// ProviderFactory manages multiple email providers
type ProviderFactory struct {
	providers    []EmailProviderInterface
	loadBalancer LoadBalancer
	mutex        sync.RWMutex
}

// LoadBalancer interface for different load balancing strategies
type LoadBalancer interface {
	SelectProvider(providers []EmailProviderInterface, emailCount int) EmailProviderInterface
	DistributeLoad(providers []EmailProviderInterface, emails []EmailNotification) map[EmailProviderInterface][]EmailNotification
}

// NewProviderFactory creates a new provider factory from dynamic configuration
func NewProviderFactory(cfg *config.ProvidersConfig) (*ProviderFactory, error) {
	factory := &ProviderFactory{
		providers: make([]EmailProviderInterface, 0),
	}

	// Initialize only enabled providers
	for _, providerName := range cfg.Enabled {
		// Check SMTP providers
		if smtpConfig, exists := cfg.SMTP[providerName]; exists {
			provider := NewDynamicSMTPProvider(providerName, &smtpConfig)

			// Wrap with batch manager if needed (SMTP doesn't support bulk)
			batchedProvider := NewBatchedEmailProvider(provider, 50, false) // 50 batch size, no bulk
			factory.providers = append(factory.providers, batchedProvider)
			continue
		}

		// Check API providers
		if apiConfig, exists := cfg.API[providerName]; exists {
			provider := NewDynamicAPIProvider(providerName, &apiConfig)

			// Wrap with batch manager based on bulk_enabled setting
			batchedProvider := NewBatchedEmailProvider(provider, apiConfig.MaxBatchSize, apiConfig.BulkEnabled)
			factory.providers = append(factory.providers, batchedProvider)
			continue
		}

		fmt.Printf("Warning: Enabled provider '%s' not found in configuration\n", providerName)
	}

	// Initialize load balancer
	switch cfg.LoadBalancing {
	case "weighted":
		factory.loadBalancer = NewWeightedLoadBalancer()
	case "least_load":
		factory.loadBalancer = NewLeastLoadBalancer()
	default:
		factory.loadBalancer = NewRoundRobinLoadBalancer()
	}

	if len(factory.providers) == 0 {
		return nil, fmt.Errorf("no enabled email providers configured")
	}

	fmt.Printf("Initialized %d email providers: %v\n", len(factory.providers), cfg.Enabled)
	return factory, nil
}

// GetProvider returns a provider based on load balancing strategy
func (f *ProviderFactory) GetProvider(emailCount int) EmailProviderInterface {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	return f.loadBalancer.SelectProvider(f.providers, emailCount)
}

// GetProviders returns all enabled providers
func (f *ProviderFactory) GetProviders() []EmailProviderInterface {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	// Filter enabled providers
	enabled := make([]EmailProviderInterface, 0)
	for _, provider := range f.providers {
		if provider.IsEnabled() {
			enabled = append(enabled, provider)
		}
	}
	return enabled
}

// DistributeEmails distributes emails across providers
func (f *ProviderFactory) DistributeEmails(emails []EmailNotification) map[EmailProviderInterface][]EmailNotification {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	return f.loadBalancer.DistributeLoad(f.providers, emails)
}

// GetHealthyProviders returns only healthy providers
func (f *ProviderFactory) GetHealthyProviders() []EmailProviderInterface {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	healthy := make([]EmailProviderInterface, 0)
	for _, provider := range f.providers {
		if provider.IsEnabled() && provider.GetStats().IsHealthy {
			healthy = append(healthy, provider)
		}
	}
	return healthy
}

// GetBulkCapableProviders returns providers that support bulk operations
func (f *ProviderFactory) GetBulkCapableProviders() []EmailProviderInterface {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	bulk := make([]EmailProviderInterface, 0)
	for _, provider := range f.providers {
		if provider.IsEnabled() && provider.SupportsBulk() {
			bulk = append(bulk, provider)
		}
	}
	return bulk
}

// Legacy Factory for backward compatibility
type Factory struct {
	providers map[EmailProvider]func(ProviderConfig) (EmailProviderInterface, error)
}

// NewFactory creates a new provider factory (legacy)
func NewFactory() *Factory {
	factory := &Factory{
		providers: make(map[EmailProvider]func(ProviderConfig) (EmailProviderInterface, error)),
	}

	// Register default providers
	factory.RegisterProvider(SMTPProvider, createSMTPProvider)

	return factory
}

// RegisterProvider registers a new provider creator function
func (f *Factory) RegisterProvider(providerType EmailProvider, creator func(ProviderConfig) (EmailProviderInterface, error)) {
	f.providers[providerType] = creator
}

// CreateProvider creates an email provider based on configuration
func (f *Factory) CreateProvider(config ProviderConfig) (EmailProviderInterface, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid provider config: %w", err)
	}

	creator, exists := f.providers[config.GetProviderType()]
	if !exists {
		return nil, fmt.Errorf("unsupported email provider: %s", config.GetProviderType())
	}

	return creator(config)
}

// GetSupportedProviders returns list of supported providers
func (f *Factory) GetSupportedProviders() []EmailProvider {
	providers := make([]EmailProvider, 0, len(f.providers))
	for provider := range f.providers {
		providers = append(providers, provider)
	}
	return providers
}

// createSMTPProvider is the factory function for SMTP provider
func createSMTPProvider(config ProviderConfig) (EmailProviderInterface, error) {
	smtpConfig, ok := config.(*SMTPConfig)
	if !ok {
		return nil, fmt.Errorf("invalid SMTP config type")
	}
	return NewSMTPProvider(smtpConfig), nil
}
