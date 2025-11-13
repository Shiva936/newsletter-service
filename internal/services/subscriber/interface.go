package subscriber

import "context"

// BulkSubscriberUpdate represents an update operation for bulk processing
type BulkSubscriberUpdate struct {
	ID         uint                   `json:"id"`
	Updates    map[string]interface{} `json:"updates"`
	TopicNames []string               `json:"topic_names"`
}

type Repository interface {
	Create(ctx context.Context, subscriber *Subscriber) error
	CreateWithTopics(ctx context.Context, subscriber *Subscriber, topicIDs []uint) error
	GetByID(ctx context.Context, id uint) (*Subscriber, error)
	GetByIDWithTopics(ctx context.Context, id uint) (*Subscriber, []string, error)
	GetAll(ctx context.Context) ([]*Subscriber, error)
	GetAllWithPagination(ctx context.Context, offset, limit int) ([]*Subscriber, int64, error)
	Update(ctx context.Context, id uint, updates map[string]interface{}) error
	UpdateSubscribedTopics(ctx context.Context, subscriberID uint, topicIDs []uint) error
	Delete(ctx context.Context, id uint) error
	Subscribe(ctx context.Context, subscriberID, topicID uint) error
	Unsubscribe(ctx context.Context, subscriptionID uint) error
	GetAllSubscriptions(ctx context.Context) ([]*Subscription, error)
	GetAllSubscriptionsWithPagination(ctx context.Context, offset, limit int) ([]*Subscription, int64, error)
	GetSubscriptionsBySubscriberID(ctx context.Context, subscriberID uint) ([]*Subscription, error)
	GetSubscriptionsByTopicID(ctx context.Context, topicID uint) ([]*Subscription, error)
	GetSubscribedTopicNames(ctx context.Context, subscriberID uint) ([]string, error)
}

type Service interface {
	CreateSubscriber(ctx context.Context, subscriber *Subscriber) error
	CreateSubscriberWithTopics(ctx context.Context, subscriber *Subscriber, topicNames []string) error
	BulkCreateSubscribers(ctx context.Context, subscribers []*Subscriber, topicNamesList [][]string) ([]uint, []error)
	GetSubscriberByID(ctx context.Context, id uint) (*Subscriber, error)
	GetSubscriberByIDWithTopics(ctx context.Context, id uint) (*Subscriber, []string, error)
	GetAllSubscribers(ctx context.Context) ([]*Subscriber, error)
	GetAllSubscribersWithPagination(ctx context.Context, offset, limit int) ([]*Subscriber, int64, error)
	UpdateSubscriber(ctx context.Context, id uint, updates map[string]interface{}) error
	UpdateSubscriberWithTopics(ctx context.Context, id uint, updates map[string]interface{}, topicNames []string) error
	BulkUpdateSubscribers(ctx context.Context, updates []BulkSubscriberUpdate) []error
	DeleteSubscriber(ctx context.Context, id uint) error
	BulkDeleteSubscribers(ctx context.Context, ids []uint) []error
	Subscribe(ctx context.Context, subscriberID, topicID uint) error
	Unsubscribe(ctx context.Context, subscriptionID uint) error
	GetAllSubscriptions(ctx context.Context) ([]*Subscription, error)
	GetAllSubscriptionsWithPagination(ctx context.Context, offset, limit int) ([]*Subscription, int64, error)
	GetSubscriptionsBySubscriberID(ctx context.Context, subscriberID uint) ([]*Subscription, error)
	GetSubscriptionsByTopicID(ctx context.Context, topicID uint) ([]*Subscription, error)
}
