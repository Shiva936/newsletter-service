package topic

import (
	"context"

	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, topic *Topic) error {
	return r.db.WithContext(ctx).Create(topic).Error
}

func (r *repository) GetByID(ctx context.Context, id uint) (*Topic, error) {
	var topic Topic
	err := r.db.WithContext(ctx).First(&topic, id).Error
	if err != nil {
		return nil, err
	}
	return &topic, nil
}

func (r *repository) GetAll(ctx context.Context) ([]*Topic, error) {
	var topics []*Topic
	err := r.db.WithContext(ctx).Order("created_at desc").Find(&topics).Error
	return topics, err
}

func (r *repository) GetAllWithPagination(ctx context.Context, offset, limit int) ([]*Topic, int64, error) {
	var topics []*Topic
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&Topic{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.WithContext(ctx).Order("created_at desc").Offset(offset).Limit(limit).Find(&topics).Error
	return topics, total, err
}

func (r *repository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&Topic{}).Where("id = ?", id).Updates(updates).Error
}

func (r *repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Topic{}, id).Error
}

func (r *repository) GetByName(ctx context.Context, name string) (*Topic, error) {
	var topic Topic
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&topic).Error
	if err != nil {
		return nil, err
	}
	return &topic, nil
}

func (r *repository) GetByNames(ctx context.Context, names []string) ([]*Topic, error) {
	var topics []*Topic
	err := r.db.WithContext(ctx).Where("name IN ?", names).Find(&topics).Error
	return topics, err
}
