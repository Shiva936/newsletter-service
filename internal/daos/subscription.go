package daos

import (
	"time"

	"gorm.io/gorm"
)

// Subscription represents a subscription relationship between subscriber and topic
type Subscription struct {
	ID           uint           `json:"id" gorm:"primarykey"`
	SubscriberID uint           `json:"subscriber_id" gorm:"not null;index"`
	TopicID      uint           `json:"topic_id" gorm:"not null;index"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Subscriber *Subscriber `json:"subscriber,omitempty" gorm:"foreignKey:SubscriberID"`
	Topic      *Topic      `json:"topic,omitempty" gorm:"foreignKey:TopicID"`
}

// TableName returns the table name for Subscription
func (Subscription) TableName() string {
	return "subscriptions"
}
