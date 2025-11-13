package handlers

import (
	"newsletter-service/internal/services/content"
	"newsletter-service/internal/services/notification"
	"newsletter-service/internal/services/subscriber"
	"newsletter-service/internal/services/topic"
)

// Handler aggregates all individual handlers
type Handler struct {
	Topic        *TopicHandler
	Subscriber   *SubscriberHandler
	Content      *ContentHandler
	Notification *NotificationHandler
	Health       *HealthHandler
	Unsubscribe  *UnsubscribeHandler
}

// NewHandler creates a new handler with all service handlers
func NewHandler(
	topicService topic.Service,
	subscriberService subscriber.Service,
	contentService content.Service,
	notificationService notification.Service,
) *Handler {
	return &Handler{
		Topic:        NewTopicHandler(topicService),
		Subscriber:   NewSubscriberHandler(subscriberService),
		Content:      NewContentHandler(contentService),
		Notification: NewNotificationHandler(notificationService),
		Health:       NewHealthHandler(),
		Unsubscribe:  NewUnsubscribeHandler(subscriberService),
	}
}
