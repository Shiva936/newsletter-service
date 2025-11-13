package content

import "context"

type Repository interface {
	Create(ctx context.Context, content *Content) error
	GetByID(ctx context.Context, id uint) (*Content, error)
	GetAll(ctx context.Context) ([]*Content, error)
	GetAllWithPagination(ctx context.Context, offset, limit int) ([]*Content, int64, error)
	Update(ctx context.Context, id uint, updates map[string]interface{}) error
	Delete(ctx context.Context, id uint) error
	Publish(ctx context.Context, id uint) error
	GetPendingNotifications(ctx context.Context) ([]uint, error)
	MarkNotificationsSent(ctx context.Context, id uint) error
}

type Service interface {
	CreateContent(ctx context.Context, content *Content) error
	GetContentByID(ctx context.Context, id uint) (*Content, error)
	GetAllContent(ctx context.Context) ([]*Content, error)
	GetAllContentWithPagination(ctx context.Context, offset, limit int) ([]*Content, int64, error)
	UpdateContent(ctx context.Context, id uint, updates map[string]interface{}) error
	DeleteContent(ctx context.Context, id uint) error
	PublishContent(ctx context.Context, id uint) error
	GetPendingNotifications(ctx context.Context) ([]uint, error)
	MarkNotificationsSent(ctx context.Context, id uint) error
}
