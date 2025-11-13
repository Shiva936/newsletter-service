package subscriber

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

func (r *repository) Create(ctx context.Context, subscriber *Subscriber) error {
	return r.db.WithContext(ctx).Create(subscriber).Error
}

func (r *repository) GetByID(ctx context.Context, id uint) (*Subscriber, error) {
	var subscriber Subscriber
	err := r.db.WithContext(ctx).First(&subscriber, id).Error
	if err != nil {
		return nil, err
	}
	return &subscriber, nil
}

func (r *repository) GetAll(ctx context.Context) ([]*Subscriber, error) {
	var subscribers []*Subscriber
	err := r.db.WithContext(ctx).Order("created_at desc").Find(&subscribers).Error
	return subscribers, err
}

func (r *repository) GetAllWithPagination(ctx context.Context, offset, limit int) ([]*Subscriber, int64, error) {
	var subscribers []*Subscriber
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&Subscriber{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.WithContext(ctx).Order("created_at desc").Offset(offset).Limit(limit).Find(&subscribers).Error
	return subscribers, total, err
}

func (r *repository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&Subscriber{}).Where("id = ?", id).Updates(updates).Error
}

func (r *repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Subscriber{}, id).Error
}

func (r *repository) Subscribe(ctx context.Context, subscriberID, topicID uint) error {
	subscription := &Subscription{
		SubscriberID: subscriberID,
		TopicID:      topicID,
	}
	return r.db.WithContext(ctx).Create(subscription).Error
}

func (r *repository) Unsubscribe(ctx context.Context, subscriptionID uint) error {
	return r.db.WithContext(ctx).Delete(&Subscription{}, subscriptionID).Error
}

func (r *repository) GetAllSubscriptions(ctx context.Context) ([]*Subscription, error) {
	var subscriptions []*Subscription
	err := r.db.WithContext(ctx).Order("created_at desc").Find(&subscriptions).Error
	return subscriptions, err
}

func (r *repository) GetAllSubscriptionsWithPagination(ctx context.Context, offset, limit int) ([]*Subscription, int64, error) {
	var subscriptions []*Subscription
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&Subscription{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.WithContext(ctx).Order("created_at desc").Offset(offset).Limit(limit).Find(&subscriptions).Error
	return subscriptions, total, err
}

func (r *repository) GetSubscriptionsBySubscriberID(ctx context.Context, subscriberID uint) ([]*Subscription, error) {
	var subscriptions []*Subscription
	err := r.db.WithContext(ctx).Where("subscriber_id = ?", subscriberID).Order("created_at desc").Find(&subscriptions).Error
	return subscriptions, err
}

func (r *repository) GetSubscriptionsByTopicID(ctx context.Context, topicID uint) ([]*Subscription, error) {
	var subscriptions []*Subscription
	err := r.db.WithContext(ctx).Where("topic_id = ?", topicID).Order("created_at desc").Find(&subscriptions).Error
	return subscriptions, err
}

func (r *repository) CreateWithTopics(ctx context.Context, subscriber *Subscriber, topicIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create subscriber
		if err := tx.Create(subscriber).Error; err != nil {
			return err
		}

		// Create subscriptions
		for _, topicID := range topicIDs {
			subscription := &Subscription{
				SubscriberID: subscriber.ID,
				TopicID:      topicID,
			}
			if err := tx.Create(subscription).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *repository) GetByIDWithTopics(ctx context.Context, id uint) (*Subscriber, []string, error) {
	var subscriber Subscriber
	err := r.db.WithContext(ctx).First(&subscriber, id).Error
	if err != nil {
		return nil, nil, err
	}

	topicNames, err := r.GetSubscribedTopicNames(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	return &subscriber, topicNames, nil
}

func (r *repository) UpdateSubscribedTopics(ctx context.Context, subscriberID uint, topicIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete existing subscriptions
		if err := tx.Where("subscriber_id = ?", subscriberID).Delete(&Subscription{}).Error; err != nil {
			return err
		}

		// Create new subscriptions
		for _, topicID := range topicIDs {
			subscription := &Subscription{
				SubscriberID: subscriberID,
				TopicID:      topicID,
			}
			if err := tx.Create(subscription).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *repository) GetSubscribedTopicNames(ctx context.Context, subscriberID uint) ([]string, error) {
	var topicNames []string
	err := r.db.WithContext(ctx).
		Table("subscriptions").
		Select("topics.name").
		Joins("JOIN topics ON topics.id = subscriptions.topic_id").
		Where("subscriptions.subscriber_id = ? AND subscriptions.deleted_at IS NULL AND topics.deleted_at IS NULL", subscriberID).
		Pluck("topics.name", &topicNames).Error
	return topicNames, err
}
