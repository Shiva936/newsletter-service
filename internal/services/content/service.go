package content

import "context"

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateContent(ctx context.Context, content *Content) error {
	return s.repo.Create(ctx, content)
}

func (s *service) GetContentByID(ctx context.Context, id uint) (*Content, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) GetAllContent(ctx context.Context) ([]*Content, error) {
	return s.repo.GetAll(ctx)
}

func (s *service) GetAllContentWithPagination(ctx context.Context, offset, limit int) ([]*Content, int64, error) {
	return s.repo.GetAllWithPagination(ctx, offset, limit)
}

func (s *service) UpdateContent(ctx context.Context, id uint, updates map[string]interface{}) error {
	return s.repo.Update(ctx, id, updates)
}

func (s *service) DeleteContent(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *service) PublishContent(ctx context.Context, id uint) error {
	return s.repo.Publish(ctx, id)
}

func (s *service) GetPendingNotifications(ctx context.Context) ([]uint, error) {
	return s.repo.GetPendingNotifications(ctx)
}

func (s *service) MarkNotificationsSent(ctx context.Context, id uint) error {
	return s.repo.MarkNotificationsSent(ctx, id)
}
