package subscriber

import (
	"context"
	"fmt"

	"newsletter-service/internal/services/topic"
)

type service struct {
	repo         Repository
	topicService topic.Service
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func NewServiceWithTopic(repo Repository, topicService topic.Service) Service {
	return &service{
		repo:         repo,
		topicService: topicService,
	}
}

func (s *service) CreateSubscriber(ctx context.Context, subscriber *Subscriber) error {
	return s.repo.Create(ctx, subscriber)
}

func (s *service) GetSubscriberByID(ctx context.Context, id uint) (*Subscriber, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) GetAllSubscribers(ctx context.Context) ([]*Subscriber, error) {
	return s.repo.GetAll(ctx)
}

func (s *service) GetAllSubscribersWithPagination(ctx context.Context, offset, limit int) ([]*Subscriber, int64, error) {
	return s.repo.GetAllWithPagination(ctx, offset, limit)
}

func (s *service) UpdateSubscriber(ctx context.Context, id uint, updates map[string]interface{}) error {
	return s.repo.Update(ctx, id, updates)
}

func (s *service) DeleteSubscriber(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *service) Subscribe(ctx context.Context, subscriberID, topicID uint) error {
	return s.repo.Subscribe(ctx, subscriberID, topicID)
}

func (s *service) Unsubscribe(ctx context.Context, subscriptionID uint) error {
	return s.repo.Unsubscribe(ctx, subscriptionID)
}

func (s *service) GetAllSubscriptions(ctx context.Context) ([]*Subscription, error) {
	return s.repo.GetAllSubscriptions(ctx)
}

func (s *service) GetAllSubscriptionsWithPagination(ctx context.Context, offset, limit int) ([]*Subscription, int64, error) {
	return s.repo.GetAllSubscriptionsWithPagination(ctx, offset, limit)
}

func (s *service) GetSubscriptionsBySubscriberID(ctx context.Context, subscriberID uint) ([]*Subscription, error) {
	return s.repo.GetSubscriptionsBySubscriberID(ctx, subscriberID)
}

func (s *service) GetSubscriptionsByTopicID(ctx context.Context, topicID uint) ([]*Subscription, error) {
	return s.repo.GetSubscriptionsByTopicID(ctx, topicID)
}

func (s *service) CreateSubscriberWithTopics(ctx context.Context, subscriber *Subscriber, topicNames []string) error {
	if s.topicService == nil {
		return fmt.Errorf("topic service not available - use NewServiceWithTopic")
	}

	// Get topics by names
	topics, err := s.topicService.GetTopicsByNames(ctx, topicNames)
	if err != nil {
		return fmt.Errorf("failed to get topics: %w", err)
	}

	// Check if all topics were found
	if len(topics) != len(topicNames) {
		return fmt.Errorf("some topics not found")
	}

	// Extract topic IDs
	topicIDs := make([]uint, len(topics))
	for i, topic := range topics {
		topicIDs[i] = topic.ID
	}

	return s.repo.CreateWithTopics(ctx, subscriber, topicIDs)
}

func (s *service) GetSubscriberByIDWithTopics(ctx context.Context, id uint) (*Subscriber, []string, error) {
	return s.repo.GetByIDWithTopics(ctx, id)
}

func (s *service) UpdateSubscriberWithTopics(ctx context.Context, id uint, updates map[string]interface{}, topicNames []string) error {
	if s.topicService == nil {
		return fmt.Errorf("topic service not available - use NewServiceWithTopic")
	}

	// Update subscriber fields first
	if len(updates) > 0 {
		if err := s.repo.Update(ctx, id, updates); err != nil {
			return err
		}
	}

	// Update topics if provided
	if topicNames != nil {
		topics, err := s.topicService.GetTopicsByNames(ctx, topicNames)
		if err != nil {
			return fmt.Errorf("failed to get topics: %w", err)
		}

		// Check if all topics were found
		if len(topics) != len(topicNames) {
			return fmt.Errorf("some topics not found")
		}

		// Extract topic IDs
		topicIDs := make([]uint, len(topics))
		for i, topic := range topics {
			topicIDs[i] = topic.ID
		}

		if err := s.repo.UpdateSubscribedTopics(ctx, id, topicIDs); err != nil {
			return fmt.Errorf("failed to update subscriptions: %w", err)
		}
	}

	return nil
}

func (s *service) BulkCreateSubscribers(ctx context.Context, subscribers []*Subscriber, topicNamesList [][]string) ([]uint, []error) {
	var successIDs []uint
	var errors []error

	for i, subscriber := range subscribers {
		var topicNames []string
		if i < len(topicNamesList) {
			topicNames = topicNamesList[i]
		}

		if len(topicNames) > 0 {
			err := s.CreateSubscriberWithTopics(ctx, subscriber, topicNames)
			if err != nil {
				errors = append(errors, fmt.Errorf("subscriber %d: %w", i, err))
			} else {
				successIDs = append(successIDs, subscriber.ID)
			}
		} else {
			err := s.CreateSubscriber(ctx, subscriber)
			if err != nil {
				errors = append(errors, fmt.Errorf("subscriber %d: %w", i, err))
			} else {
				successIDs = append(successIDs, subscriber.ID)
			}
		}
	}

	return successIDs, errors
}

func (s *service) BulkUpdateSubscribers(ctx context.Context, updates []BulkSubscriberUpdate) []error {
	var errors []error

	for i, update := range updates {
		err := s.UpdateSubscriberWithTopics(ctx, update.ID, update.Updates, update.TopicNames)
		if err != nil {
			errors = append(errors, fmt.Errorf("subscriber %d (ID: %d): %w", i, update.ID, err))
		}
	}

	return errors
}

func (s *service) BulkDeleteSubscribers(ctx context.Context, ids []uint) []error {
	var errors []error

	for i, id := range ids {
		err := s.DeleteSubscriber(ctx, id)
		if err != nil {
			errors = append(errors, fmt.Errorf("subscriber %d (ID: %d): %w", i, id, err))
		}
	}

	return errors
}
