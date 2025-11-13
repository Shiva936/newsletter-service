package schedulers

import (
	"context"
	"log"

	"newsletter-service/internal/providers"
	"newsletter-service/internal/services/content"
	"newsletter-service/internal/services/notification"
)

type NotificationScheduler struct {
	contentService      content.Service
	notificationService notification.Service
	emailProvider       providers.EmailProviderInterface
}

func NewNotificationScheduler(contentService content.Service, notificationService notification.Service) *NotificationScheduler {
	return &NotificationScheduler{
		contentService:      contentService,
		notificationService: notificationService,
	}
}

func NewNotificationSchedulerWithProvider(contentService content.Service, notificationService notification.Service, emailProvider providers.EmailProviderInterface) *NotificationScheduler {
	return &NotificationScheduler{
		contentService:      contentService,
		notificationService: notificationService,
		emailProvider:       emailProvider,
	}
}

func (s *NotificationScheduler) ProcessPendingNotifications(ctx context.Context) error {
	// Get pending content IDs
	pendingContentIDs, err := s.contentService.GetPendingNotifications(ctx)
	if err != nil {
		return err
	}

	log.Printf("Found %d pending notifications", len(pendingContentIDs))

	// Process each content ID
	for _, contentID := range pendingContentIDs {
		log.Printf("Processing notification for content ID: %d", contentID)

		// Use provider-aware method if provider is available, otherwise use standard method
		var err error
		if s.emailProvider != nil {
			err = s.notificationService.SendNotificationsByContentIDWithProvider(ctx, contentID, s.emailProvider)
		} else {
			err = s.notificationService.SendNotificationsByContentID(ctx, contentID)
		}

		if err != nil {
			log.Printf("Failed to send notification for content %d: %v", contentID, err)
			continue
		}

		log.Printf("Successfully sent notification for content ID: %d", contentID)
	}

	return nil
}

// RetryFailedNotifications retries sending failed email notifications
func (s *NotificationScheduler) RetryFailedNotifications(ctx context.Context) error {
	log.Printf("Starting failed notifications retry process")

	// Use provider-aware method if provider is available, otherwise use standard method
	var err error
	if s.emailProvider != nil {
		err = s.notificationService.RetryFailedEmailsWithProvider(ctx, s.emailProvider)
	} else {
		err = s.notificationService.RetryFailedEmails(ctx)
	}

	if err != nil {
		log.Printf("Failed to retry failed notifications: %v", err)
		return err
	}

	log.Printf("Successfully initiated retry for failed notifications")
	return nil
}
