package daos

import (
	"time"

	"gorm.io/gorm"
)

// Subscriber represents a newsletter subscriber in the database
type Subscriber struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	Name      string         `json:"name" gorm:"size:100;not null"`
	Email     string         `json:"email" gorm:"uniqueIndex;size:255;not null"`
	IsActive  bool           `json:"is_active" gorm:"default:true;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Subscriptions []Subscription `json:"subscriptions,omitempty" gorm:"foreignKey:SubscriberID"`
	EmailLogs     []EmailLog     `json:"email_logs,omitempty" gorm:"foreignKey:SubscriberID"`
}

// TableName returns the table name for Subscriber
func (Subscriber) TableName() string {
	return "subscribers"
}
