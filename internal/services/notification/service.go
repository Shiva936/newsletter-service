package notification

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"

	"newsletter-service/internal/config"
	"newsletter-service/internal/constants"
	"newsletter-service/internal/providers"
	"newsletter-service/internal/services/content"
	"newsletter-service/internal/services/subscriber"
)

type notificationService struct {
	db                *gorm.DB
	contentService    content.Service
	subscriberService subscriber.Service
	providerFactory   *providers.ProviderFactory
	workerConfig      *config.WorkerConfig
}

func NewService(db *gorm.DB, contentService content.Service, subscriberService subscriber.Service) Service {
	return &notificationService{
		db:                db,
		contentService:    contentService,
		subscriberService: subscriberService,
	}
}

// NewServiceWithProviders creates a notification service with multi-provider support
func NewServiceWithProviders(db *gorm.DB, contentService content.Service, subscriberService subscriber.Service, cfg *config.Config) (Service, error) {
	// Initialize provider factory
	providerFactory, err := providers.NewProviderFactory(&cfg.Providers)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize provider factory: %w", err)
	}

	return &notificationService{
		db:                db,
		contentService:    contentService,
		subscriberService: subscriberService,
		providerFactory:   providerFactory,
		workerConfig:      &cfg.Worker,
	}, nil
}

// SendNotificationsByContentID sends notifications without provider (for backward compatibility)
func (s *notificationService) SendNotificationsByContentID(ctx context.Context, contentID uint) error {
	return fmt.Errorf("provider is required for sending notifications - use SendNotificationsByContentIDWithProvider")
}

// SendNotificationsByContentIDWithProvider sends notifications using the provided email provider with goroutine parallelization
func (s *notificationService) SendNotificationsByContentIDWithProvider(ctx context.Context, contentID uint, provider providers.EmailProviderInterface) error {
	if s.providerFactory != nil {
		// Use multi-provider approach if available
		return s.sendNotificationsMultiProvider(ctx, contentID)
	}

	// Fall back to single provider approach for backward compatibility
	return s.sendNotificationsSingleProvider(ctx, contentID, provider)
}

// sendNotificationsSingleProvider handles single provider backward compatibility
func (s *notificationService) sendNotificationsSingleProvider(ctx context.Context, contentID uint, provider providers.EmailProviderInterface) error {
	// Get content
	content, err := s.contentService.GetContentByID(ctx, contentID)
	if err != nil {
		return fmt.Errorf("failed to get content: %w", err)
	}

	// Get subscribers for the topic
	subscriptions, err := s.subscriberService.GetSubscriptionsByTopicID(ctx, content.TopicID)
	if err != nil {
		return fmt.Errorf("failed to get subscriptions: %w", err)
	}

	// Get active subscribers
	var activeSubscribers []struct {
		ID    uint
		Email string
	}

	for _, subscription := range subscriptions {
		subscriber, err := s.subscriberService.GetSubscriberByID(ctx, subscription.SubscriberID)
		if err != nil || !subscriber.IsActive {
			continue
		}

		activeSubscribers = append(activeSubscribers, struct {
			ID    uint
			Email string
		}{
			ID:    subscriber.ID,
			Email: subscriber.Email,
		})
	}

	if len(activeSubscribers) == 0 {
		fmt.Printf("No active subscribers found for content ID %d\n", contentID)
		return nil
	}

	// Send emails using the single provider
	sentCount := s.sendEmailsConcurrently(ctx, contentID, activeSubscribers, content, provider)
	totalCount := len(activeSubscribers)

	// Mark notifications as sent
	if totalCount > 0 {
		if markErr := s.contentService.MarkNotificationsSent(ctx, contentID); markErr != nil {
			fmt.Printf("Failed to mark notifications as sent for content %d: %v\n", contentID, markErr)
		}
	}

	fmt.Printf("Sent %d/%d notifications for content ID %d\n", sentCount, totalCount, contentID)
	return nil
}

// sendNotificationsMultiProvider uses the multi-provider architecture
func (s *notificationService) sendNotificationsMultiProvider(ctx context.Context, contentID uint) error {
	// Get content
	content, err := s.contentService.GetContentByID(ctx, contentID)
	if err != nil {
		return fmt.Errorf("failed to get content: %w", err)
	}

	// Get subscribers for the topic
	subscriptions, err := s.subscriberService.GetSubscriptionsByTopicID(ctx, content.TopicID)
	if err != nil {
		return fmt.Errorf("failed to get subscriptions: %w", err)
	}

	// Collect active subscriber emails
	var activeEmails []providers.EmailNotification
	var activeSubscribers []struct {
		ID    uint
		Email string
	}

	for _, subscription := range subscriptions {
		subscriber, err := s.subscriberService.GetSubscriberByID(ctx, subscription.SubscriberID)
		if err != nil || !subscriber.IsActive {
			continue
		}

		activeEmails = append(activeEmails, providers.EmailNotification{
			To:      subscriber.Email,
			Subject: content.Title,
			Body:    content.Body,
		})

		activeSubscribers = append(activeSubscribers, struct {
			ID    uint
			Email string
		}{
			ID:    subscriber.ID,
			Email: subscriber.Email,
		})
	}

	if len(activeEmails) == 0 {
		fmt.Printf("No active subscribers found for content ID %d\n", contentID)
		return nil
	}

	// Check if we should use bulk providers
	bulkProviders := s.providerFactory.GetBulkCapableProviders()
	if len(activeEmails) > 10 && len(bulkProviders) > 0 {
		// Use bulk sending for large lists
		return s.sendBulkEmails(ctx, contentID, activeEmails, activeSubscribers, content)
	}

	// Use distributed individual sending
	return s.sendDistributedEmails(ctx, contentID, activeEmails, activeSubscribers, content)
}

// sendBulkEmails uses bulk-capable providers for large email lists
func (s *notificationService) sendBulkEmails(ctx context.Context, contentID uint, emails []providers.EmailNotification, subscribers []struct {
	ID    uint
	Email string
}, content *content.Content) error {

	bulkProviders := s.providerFactory.GetBulkCapableProviders()
	if len(bulkProviders) == 0 {
		return fmt.Errorf("no bulk capable providers available")
	}

	// Use the best bulk provider (highest priority, healthy)
	bestProvider := bulkProviders[0]
	for _, provider := range bulkProviders {
		if provider.GetStats().IsHealthy && provider.GetPriority() < bestProvider.GetPriority() {
			bestProvider = provider
		}
	}

	// Prepare bulk notification
	recipientEmails := make([]string, len(emails))
	for i, email := range emails {
		recipientEmails[i] = email.To
	}

	bulkNotification := &providers.BulkEmailNotification{
		To:      recipientEmails,
		Subject: content.Title,
		Body:    content.Body,
	}

	// Send bulk email
	if err := bestProvider.SendBulkEmail(ctx, bulkNotification); err != nil {
		fmt.Printf("Bulk email failed (%v), falling back to distributed sending\n", err)
		return s.sendDistributedEmails(ctx, contentID, emails, subscribers, content)
	}

	// Log success for all subscribers
	return s.logBulkEmailSuccess(ctx, contentID, subscribers, content)
}

// sendDistributedEmails distributes emails across multiple providers
func (s *notificationService) sendDistributedEmails(ctx context.Context, contentID uint, emails []providers.EmailNotification, subscribers []struct {
	ID    uint
	Email string
}, content *content.Content) error {

	// Distribute emails across healthy providers
	distribution := s.providerFactory.DistributeEmails(emails)

	var wg sync.WaitGroup
	concurrencyLimit := s.getConcurrencyLimit()
	semaphore := make(chan struct{}, concurrencyLimit)
	successCount := make(chan int, len(emails))

	// Send emails for each provider distribution
	for provider, providerEmails := range distribution {
		if !provider.GetStats().IsHealthy {
			continue
		}

		for _, email := range providerEmails {
			wg.Add(1)
			go func(p providers.EmailProviderInterface, e providers.EmailNotification) {
				defer wg.Done()
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				// Find subscriber for this email
				var subscriberID uint
				for _, sub := range subscribers {
					if sub.Email == e.To {
						subscriberID = sub.ID
						break
					}
				}

				// Send email and log result
				if err := p.SendEmail(ctx, &e); err != nil {
					s.logEmailFailure(ctx, contentID, subscriberID, e, err)
					successCount <- 0
				} else {
					s.logEmailSuccess(ctx, contentID, subscriberID, e)
					successCount <- 1
				}
			}(provider, email)
		}
	}

	// Wait for completion and count results
	go func() {
		wg.Wait()
		close(successCount)
	}()

	sentCount := 0
	for count := range successCount {
		sentCount += count
	}

	// Mark notifications as sent
	if sentCount > 0 {
		if markErr := s.contentService.MarkNotificationsSent(ctx, contentID); markErr != nil {
			fmt.Printf("Failed to mark notifications as sent for content %d: %v\n", contentID, markErr)
		}
	}

	fmt.Printf("Sent %d/%d notifications for content ID %d using multi-provider distribution\n", sentCount, len(emails), contentID)
	return nil
}

// getConcurrencyLimit returns the appropriate concurrency limit based on configuration
func (s *notificationService) getConcurrencyLimit() int {
	if s.workerConfig != nil {
		return s.workerConfig.MaxAsyncProcess
	}
	return 10 // Default
}

// Helper methods for logging
func (s *notificationService) logEmailSuccess(ctx context.Context, contentID uint, subscriberID uint, email providers.EmailNotification) {
	now := time.Now()
	emailLog := &EmailLog{
		SubscriberID: subscriberID,
		ContentID:    contentID,
		EmailAddress: email.To,
		Subject:      email.Subject,
		Body:         email.Body,
		Status:       constants.StatusSent,
		SentAt:       &now,
		RetryCount:   0,
	}

	if err := s.LogEmail(ctx, emailLog); err != nil {
		fmt.Printf("Failed to log email success for %s: %v\n", email.To, err)
	}
}

func (s *notificationService) logEmailFailure(ctx context.Context, contentID uint, subscriberID uint, email providers.EmailNotification, sendErr error) {
	emailLog := &EmailLog{
		SubscriberID: subscriberID,
		ContentID:    contentID,
		EmailAddress: email.To,
		Subject:      email.Subject,
		Body:         email.Body,
		Status:       constants.StatusFailed,
		RetryCount:   0,
	}

	if sendErr != nil {
		errorMsg := sendErr.Error()
		emailLog.ErrorMessage = &errorMsg
	}

	if err := s.LogEmail(ctx, emailLog); err != nil {
		fmt.Printf("Failed to log email failure for %s: %v\n", email.To, err)
	}
}

func (s *notificationService) logBulkEmailSuccess(ctx context.Context, contentID uint, subscribers []struct {
	ID    uint
	Email string
}, content *content.Content) error {
	now := time.Now()
	var wg sync.WaitGroup

	for _, subscriber := range subscribers {
		wg.Add(1)
		go func(subID uint, email string) {
			defer wg.Done()
			emailLog := &EmailLog{
				SubscriberID: subID,
				ContentID:    contentID,
				EmailAddress: email,
				Subject:      content.Title,
				Body:         content.Body,
				Status:       constants.StatusSent,
				SentAt:       &now,
				RetryCount:   0,
			}

			if err := s.LogEmail(ctx, emailLog); err != nil {
				fmt.Printf("Failed to log bulk email success for %s: %v\n", email, err)
			}
		}(subscriber.ID, subscriber.Email)
	}

	wg.Wait()
	return nil
}

// sendEmailsConcurrently sends emails using goroutines with worker pool pattern for optimal performance
func (s *notificationService) sendEmailsConcurrently(ctx context.Context, contentID uint, subscribers []struct {
	ID    uint
	Email string
}, content *content.Content, provider providers.EmailProviderInterface) int {
	var wg sync.WaitGroup

	// Use max_async_process from worker config or default to 10
	concurrencyLimit := 10 // This should ideally come from WorkerConfig.MaxAsyncProcess
	semaphore := make(chan struct{}, concurrencyLimit)
	successCount := make(chan int, len(subscribers))

	for _, subscriber := range subscribers {
		wg.Add(1)
		go func(subID uint, email string) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			notification := &providers.EmailNotification{
				To:      email,
				Subject: content.Title,
				Body:    content.Body,
			}

			emailLog := &EmailLog{
				SubscriberID: subID,
				ContentID:    contentID,
				EmailAddress: email,
				Subject:      content.Title,
				Body:         content.Body,
				Status:       constants.StatusSent,
				RetryCount:   0,
			}

			// Send email
			now := time.Now()
			if err := provider.SendEmail(ctx, notification); err != nil {
				emailLog.Status = constants.StatusFailed
				errorMsg := err.Error()
				emailLog.ErrorMessage = &errorMsg
				successCount <- 0
			} else {
				emailLog.SentAt = &now
				successCount <- 1
			}

			// Log the email
			if logErr := s.LogEmail(ctx, emailLog); logErr != nil {
				fmt.Printf("Failed to log email for %s: %v\n", email, logErr)
			}
		}(subscriber.ID, subscriber.Email)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(successCount)
	}()

	// Count successful sends
	sentCount := 0
	for count := range successCount {
		sentCount += count
	}

	return sentCount
}

// RetryFailedEmails retries without provider (for backward compatibility)
func (s *notificationService) RetryFailedEmails(ctx context.Context) error {
	return fmt.Errorf("provider is required for retrying emails - use RetryFailedEmailsWithProvider")
}

// RetryFailedEmailsWithProvider retries failed emails using the provided email provider
func (s *notificationService) RetryFailedEmailsWithProvider(ctx context.Context, provider providers.EmailProviderInterface) error {
	var failedEmails []*EmailLog

	// Get failed emails that haven't exceeded retry limit
	err := s.db.WithContext(ctx).Where("status = ? AND retry_count < ?", constants.StatusFailed, constants.MaxEmailRetryCount).Find(&failedEmails).Error
	if err != nil {
		return fmt.Errorf("failed to get failed emails: %w", err)
	}

	for _, emailLog := range failedEmails {
		// Get subscriber
		subscriber, err := s.subscriberService.GetSubscriberByID(ctx, emailLog.SubscriberID)
		if err != nil {
			continue
		}

		if !subscriber.IsActive {
			continue
		}

		notification := &providers.EmailNotification{
			To:      emailLog.EmailAddress,
			Subject: emailLog.Subject,
			Body:    emailLog.Body,
		}

		// Retry sending
		if err := provider.SendEmail(ctx, notification); err != nil {
			// Update retry count
			emailLog.RetryCount++
			errorMsg := err.Error()
			emailLog.ErrorMessage = &errorMsg
		} else {
			// Mark as sent
			emailLog.Status = constants.StatusSent
			now := time.Now()
			emailLog.SentAt = &now
			emailLog.ErrorMessage = nil
		}

		// Update the log
		s.db.WithContext(ctx).Save(emailLog)
	}

	return nil
}

func (s *notificationService) GetEmailLogs(ctx context.Context) ([]*EmailLog, error) {
	var logs []*EmailLog
	err := s.db.WithContext(ctx).Find(&logs).Error
	return logs, err
}

func (s *notificationService) GetEmailLogsWithPagination(ctx context.Context, offset, limit int) ([]*EmailLog, int64, error) {
	var logs []*EmailLog
	var total int64

	// Get total count
	if err := s.db.WithContext(ctx).Model(&EmailLog{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := s.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&logs).Error
	return logs, total, err
}

func (s *notificationService) GetEmailLogByID(ctx context.Context, id uint) (*EmailLog, error) {
	var log EmailLog
	err := s.db.WithContext(ctx).First(&log, id).Error
	return &log, err
}

func (s *notificationService) LogEmail(ctx context.Context, log *EmailLog) error {
	return s.db.WithContext(ctx).Create(log).Error
}
