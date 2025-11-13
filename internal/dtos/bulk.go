package dtos

import "time"

// BulkCreateSubscribersRequest represents a request to create multiple subscribers
type BulkCreateSubscribersRequest struct {
	Subscribers []CreateSubscriberRequest `json:"subscribers" validate:"required,min=1,max=100,dive"`
}

// BulkCreateSubscribersResponse represents the response from bulk create operation
type BulkCreateSubscribersResponse struct {
	Success []SubscriberResponse `json:"success"`
	Errors  []BulkError          `json:"errors"`
	Summary BulkOperationSummary `json:"summary"`
}

// BulkUpdateSubscribersRequest represents a request to update multiple subscribers
type BulkUpdateSubscribersRequest struct {
	Updates []BulkUpdateSubscriber `json:"updates" validate:"required,min=1,max=100,dive"`
}

type BulkUpdateSubscriber struct {
	ID               uint     `json:"id" validate:"required"`
	Email            string   `json:"email" validate:"omitempty,email,max=255"`
	Name             string   `json:"name" validate:"omitempty,max=100"`
	IsActive         *bool    `json:"is_active" validate:"omitempty"`
	SubscribedTopics []string `json:"subscribed_topics" validate:"omitempty,dive,min=1"`
}

// BulkDeleteSubscribersRequest represents a request to delete multiple subscribers
type BulkDeleteSubscribersRequest struct {
	IDs []uint `json:"ids" validate:"required,min=1,max=100,dive,required"`
}

// BulkOperationSummary provides summary of bulk operation results
type BulkOperationSummary struct {
	Total       int       `json:"total"`
	Success     int       `json:"success"`
	Errors      int       `json:"errors"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
	Duration    string    `json:"duration"`
}

// BulkError represents an error that occurred during bulk operation
type BulkError struct {
	Index   int    `json:"index"`
	ID      uint   `json:"id,omitempty"`
	Email   string `json:"email,omitempty"`
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// BulkResponse represents a generic bulk operation response
type BulkResponse struct {
	Success interface{}          `json:"success"`
	Errors  []BulkError          `json:"errors"`
	Summary BulkOperationSummary `json:"summary"`
}
