package dtos

import "time"

type CreateTopicRequest struct {
	Name        string `json:"name" validate:"required,max=100"`
	Description string `json:"description"`
}

type UpdateTopicRequest struct {
	Name        string `json:"name" validate:"omitempty,max=100"`
	Description string `json:"description" validate:"omitempty"`
}

type TopicResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
