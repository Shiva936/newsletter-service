package notification

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"newsletter-service/internal/constants"
	"newsletter-service/internal/providers"
	"newsletter-service/internal/services/content"
	"newsletter-service/internal/services/subscriber"
)

type notificationService struct {
	db                *gorm.DB
	contentService    content.Service
	subscriberService subscriber.Service
}

func NewService(db *gorm.DB, contentService content.Service, subscriberService subscriber.Service) Service {
	return &notificationService{
		db:                db,
		contentService:    contentService,
		subscriberService: subscriberService,
	}
}

// SendNotificationsByContentID sends notifications without provider (for backward compatibility)
func (s *notificationService) SendNotificationsByContentID(ctx context.Context, contentID uint) error {
	return fmt.Errorf("provider is required for sending notifications - use SendNotificationsByContentIDWithProvider")
}

// SendNotificationsByContentIDWithProvider sends notifications using the provided email provider
func (s *notificationService) SendNotificationsByContentIDWithProvider(ctx context.Context, contentID uint, provider providers.EmailProviderInterface) error {
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

	// Track if any emails were sent successfully
	sentCount := 0
	totalCount := 0

	// Send emails to all subscribed users
	for _, subscription := range subscriptions {
		totalCount++

		// Get subscriber details
		subscriber, err := s.subscriberService.GetSubscriberByID(ctx, subscription.SubscriberID)
		if err != nil {
			continue // Skip this subscriber if not found
		}

		if !subscriber.IsActive {
			continue // Skip inactive subscribers
		}

		notification := &providers.EmailNotification{
			To:      subscriber.Email,
			Subject: content.Title,
			Body:    content.Body,
		}

		emailLog := &EmailLog{
			SubscriberID: subscriber.ID,
			ContentID:    contentID,
			EmailAddress: subscriber.Email,
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
		} else {
			emailLog.SentAt = &now
			sentCount++
		}

		// Log the email
		if logErr := s.LogEmail(ctx, emailLog); logErr != nil {
			fmt.Printf("Failed to log email: %v\n", logErr)
		}
	}

	// Mark notifications as sent if we processed all subscribers (even if some failed)
	if totalCount > 0 {
		if markErr := s.contentService.MarkNotificationsSent(ctx, contentID); markErr != nil {
			fmt.Printf("Failed to mark notifications as sent for content %d: %v\n", contentID, markErr)
		}
	}

	fmt.Printf("Sent %d/%d notifications for content ID %d\n", sentCount, totalCount, contentID)
	return nil
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
