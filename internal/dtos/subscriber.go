package dtos

import "time"

type CreateSubscriberRequest struct {
	Name             string   `json:"name" validate:"required,max=100"`
	Email            string   `json:"email" validate:"required,email,max=255"`
	SubscribedTopics []string `json:"subscribed_topics" validate:"omitempty,dive,min=1"`
}

type UpdateSubscriberRequest struct {
	Email            string   `json:"email" validate:"omitempty,email,max=255"`
	Name             string   `json:"name" validate:"omitempty,max=100"`
	IsActive         *bool    `json:"is_active" validate:"omitempty"`
	SubscribedTopics []string `json:"subscribed_topics" validate:"omitempty,dive,min=1"`
}

type SubscriberResponse struct {
	ID               uint      `json:"id"`
	Email            string    `json:"email"`
	Name             string    `json:"name"`
	IsActive         bool      `json:"is_active"`
	SubscribedTopics []string  `json:"subscribed_topics"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type CreateSubscriptionRequest struct {
	SubscriberID uint `json:"subscriber_id" validate:"required"`
	TopicID      uint `json:"topic_id" validate:"required"`
}

type SubscriptionResponse struct {
	ID           uint      `json:"id"`
	SubscriberID uint      `json:"subscriber_id"`
	TopicID      uint      `json:"topic_id"`
	CreatedAt    time.Time `json:"created_at"`
}
