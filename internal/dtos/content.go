package dtos

import "time"

type CreateContentRequest struct {
	TopicID uint   `json:"topic_id" validate:"required"`
	Title   string `json:"title" validate:"required,max=255"`
	Body    string `json:"body" validate:"required"`
}

type UpdateContentRequest struct {
	TopicID uint   `json:"topic_id" validate:"omitempty"`
	Title   string `json:"title" validate:"omitempty,max=255"`
	Body    string `json:"body" validate:"omitempty"`
}

type ContentResponse struct {
	ID          uint       `json:"id"`
	TopicID     uint       `json:"topic_id"`
	Title       string     `json:"title"`
	Body        string     `json:"body"`
	IsPublished bool       `json:"is_published"`
	PublishedAt *time.Time `json:"published_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
