package topic

import "context"

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateTopic(ctx context.Context, topic *Topic) error {
	return s.repo.Create(ctx, topic)
}

func (s *service) GetTopicByID(ctx context.Context, id uint) (*Topic, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) GetAllTopics(ctx context.Context) ([]*Topic, error) {
	return s.repo.GetAll(ctx)
}

func (s *service) GetAllTopicsWithPagination(ctx context.Context, offset, limit int) ([]*Topic, int64, error) {
	return s.repo.GetAllWithPagination(ctx, offset, limit)
}

func (s *service) UpdateTopic(ctx context.Context, id uint, updates map[string]interface{}) error {
	return s.repo.Update(ctx, id, updates)
}

func (s *service) DeleteTopic(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *service) GetTopicByName(ctx context.Context, name string) (*Topic, error) {
	return s.repo.GetByName(ctx, name)
}

func (s *service) GetTopicsByNames(ctx context.Context, names []string) ([]*Topic, error) {
	return s.repo.GetByNames(ctx, names)
}
