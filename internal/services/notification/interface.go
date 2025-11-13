package notification

import (
	"context"

	"newsletter-service/internal/providers"
)

// Service defines the interface for notification operations
type Service interface {
	SendNotificationsByContentID(ctx context.Context, contentID uint) error
	SendNotificationsByContentIDWithProvider(ctx context.Context, contentID uint, provider providers.EmailProviderInterface) error
	RetryFailedEmails(ctx context.Context) error
	RetryFailedEmailsWithProvider(ctx context.Context, provider providers.EmailProviderInterface) error
	GetEmailLogs(ctx context.Context) ([]*EmailLog, error)
	GetEmailLogsWithPagination(ctx context.Context, offset, limit int) ([]*EmailLog, int64, error)
	GetEmailLogByID(ctx context.Context, id uint) (*EmailLog, error)
	LogEmail(ctx context.Context, log *EmailLog) error
}