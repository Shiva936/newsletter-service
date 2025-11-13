package content

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, content *Content) error {
	return r.db.WithContext(ctx).Create(content).Error
}

func (r *repository) GetByID(ctx context.Context, id uint) (*Content, error) {
	var content Content
	err := r.db.WithContext(ctx).First(&content, id).Error
	if err != nil {
		return nil, err
	}
	return &content, nil
}

func (r *repository) GetAll(ctx context.Context) ([]*Content, error) {
	var contents []*Content
	err := r.db.WithContext(ctx).Order("created_at desc").Find(&contents).Error
	return contents, err
}

func (r *repository) GetAllWithPagination(ctx context.Context, offset, limit int) ([]*Content, int64, error) {
	var contents []*Content
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&Content{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.WithContext(ctx).Order("created_at desc").Offset(offset).Limit(limit).Find(&contents).Error
	return contents, total, err
}

func (r *repository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&Content{}).Where("id = ?", id).Updates(updates).Error
}

func (r *repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Content{}, id).Error
}

func (r *repository) Publish(ctx context.Context, id uint) error {
	now := time.Now()
	updates := map[string]interface{}{
		"is_published": true,
		"published_at": now,
	}
	return r.db.WithContext(ctx).Model(&Content{}).Where("id = ?", id).Updates(updates).Error
}

func (r *repository) GetPendingNotifications(ctx context.Context) ([]uint, error) {
	var contentIDs []uint
	err := r.db.WithContext(ctx).
		Model(&Content{}).
		Select("id").
		Where("is_published = ? AND notifications_sent = ?", true, false).
		Pluck("id", &contentIDs).Error
	return contentIDs, err
}

func (r *repository) MarkNotificationsSent(ctx context.Context, id uint) error {
	now := time.Now()
	updates := map[string]interface{}{
		"notifications_sent":    true,
		"notifications_sent_at": now,
	}
	return r.db.WithContext(ctx).Model(&Content{}).Where("id = ?", id).Updates(updates).Error
}
