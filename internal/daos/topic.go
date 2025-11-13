package daos

import (
	"time"

	"gorm.io/gorm"
)

// Topic represents a newsletter topic in the database
type Topic struct {
	ID          uint           `json:"id" gorm:"primarykey"`
	Name        string         `json:"name" gorm:"uniqueIndex;size:100;not null"`
	Description string         `json:"description" gorm:"type:text"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Contents      []Content      `json:"contents,omitempty" gorm:"foreignKey:TopicID"`
	Subscriptions []Subscription `json:"subscriptions,omitempty" gorm:"foreignKey:TopicID"`
}

// TableName returns the table name for Topic
func (Topic) TableName() string {
	return "topics"
}
