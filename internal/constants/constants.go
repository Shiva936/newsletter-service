package constants

// Status constants
const (
	StatusPending = "pending"
	StatusSent    = "sent"
	StatusFailed  = "failed"
)

// Notification constants
const (
	ProviderSMTP = "smtp"
	
	NotificationTypeEmail = "email"
)

// Subscriber status constants
const (
	SubscriberStatusActive   = true
	SubscriberStatusInactive = false
)

// Content status constants
const (
	ContentStatusDraft     = false
	ContentStatusPublished = true
)

// Pagination defaults
const (
	DefaultPageSize = 20
	MaxPageSize     = 100
	MinPageSize     = 1
	DefaultPage     = 1
)

// Retry limits
const (
	MaxEmailRetryCount = 3
	MaxRetryAttempts   = 5
)

// Rate limiting
const (
	DefaultRateLimit         = 100 // requests per minute
	DefaultRateLimitDuration = 60  // seconds
)

// Authentication
const (
	BasicAuthRealm     = "Newsletter Service API"
	SchedulerAuthRealm = "Newsletter Scheduler API"
)

// Database table names
const (
	TableNameTopics        = "topics"
	TableNameSubscribers   = "subscribers"
	TableNameSubscriptions = "subscriptions"
	TableNameContents      = "contents"
	TableNameEmailLogs     = "email_logs"
)

// API response messages
const (
	MsgTopicCreatedSuccessfully          = "Topic created successfully"
	MsgTopicUpdatedSuccessfully          = "Topic updated successfully"
	MsgTopicDeletedSuccessfully          = "Topic deleted successfully"
	MsgSubscriberCreatedSuccessfully     = "Subscriber created successfully"
	MsgSubscriberUpdatedSuccessfully     = "Subscriber updated successfully"
	MsgSubscriberDeletedSuccessfully     = "Subscriber deleted successfully"
	MsgSubscriptionCreatedSuccessfully   = "Subscription created successfully"
	MsgSubscriptionDeletedSuccessfully   = "Subscription deleted successfully"
	MsgContentCreatedSuccessfully        = "Content created successfully"
	MsgContentUpdatedSuccessfully        = "Content updated successfully"
	MsgContentDeletedSuccessfully        = "Content deleted successfully"
	MsgContentPublishedSuccessfully      = "Content published successfully"
	MsgNotificationsSentSuccessfully     = "Notifications sent successfully"
	MsgFailedNotificationsRetryInitiated = "Failed notifications retry initiated"
)

// Error messages
const (
	ErrInvalidRequestBody      = "Invalid request body"
	ErrInvalidPaginationParams = "Invalid pagination parameters"
	ErrInvalidTopicID          = "Invalid topic ID"
	ErrInvalidSubscriberID     = "Invalid subscriber ID"
	ErrInvalidSubscriptionID   = "Invalid subscription ID"
	ErrInvalidContentID        = "Invalid content ID"
	ErrInvalidEmailLogID       = "Invalid email log ID"
	ErrInvalidSendTimeFormat   = "Invalid send_time format"
	ErrTopicNotFound           = "Topic not found"
	ErrSubscriberNotFound      = "Subscriber not found"
	ErrSubscriptionNotFound    = "Subscription not found"
	ErrContentNotFound         = "Content not found"
	ErrEmailLogNotFound        = "Email log not found"
	ErrUnauthorized            = "Unauthorized"
	ErrForbidden               = "Forbidden"
	ErrTooManyRequests         = "Too many requests"
	ErrInternalServerError     = "Internal server error"
)

// Health check responses
const (
	HealthStatusHealthy  = "healthy"
	ServiceNameMain      = "newsletter-service"
	ServiceNameScheduler = "newsletter-scheduler"
)
