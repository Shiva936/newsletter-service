package providers

import (
	"fmt"
)

// Factory manages creation of email providers
type Factory struct {
	providers map[EmailProvider]func(ProviderConfig) (EmailProviderInterface, error)
}

// NewFactory creates a new provider factory
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
