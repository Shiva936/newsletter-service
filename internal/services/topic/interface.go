package topic

import (
	"context"
)

type Repository interface {
	Create(ctx context.Context, topic *Topic) error
	GetByID(ctx context.Context, id uint) (*Topic, error)
	GetByName(ctx context.Context, name string) (*Topic, error)
	GetByNames(ctx context.Context, names []string) ([]*Topic, error)
	GetAll(ctx context.Context) ([]*Topic, error)
	GetAllWithPagination(ctx context.Context, offset, limit int) ([]*Topic, int64, error)
	Update(ctx context.Context, id uint, updates map[string]interface{}) error
	Delete(ctx context.Context, id uint) error
}

type Service interface {
	CreateTopic(ctx context.Context, topic *Topic) error
	GetTopicByID(ctx context.Context, id uint) (*Topic, error)
	GetTopicByName(ctx context.Context, name string) (*Topic, error)
	GetTopicsByNames(ctx context.Context, names []string) ([]*Topic, error)
	GetAllTopics(ctx context.Context) ([]*Topic, error)
	GetAllTopicsWithPagination(ctx context.Context, offset, limit int) ([]*Topic, int64, error)
	UpdateTopic(ctx context.Context, id uint, updates map[string]interface{}) error
	DeleteTopic(ctx context.Context, id uint) error
}
