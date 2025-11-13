package daos

import (
	"time"

	"gorm.io/gorm"
)

// Content represents newsletter content in the database
type Content struct {
	ID                  uint           `json:"id" gorm:"primarykey"`
	TopicID             uint           `json:"topic_id" gorm:"not null;index"`
	Title               string         `json:"title" gorm:"size:255;not null"`
	Body                string         `json:"body" gorm:"type:text;not null"`
	IsPublished         bool           `json:"is_published" gorm:"default:false;index"`
	PublishedAt         *time.Time     `json:"published_at"`
	NotificationsSent   bool           `json:"notifications_sent" gorm:"default:false;index"`
	NotificationsSentAt *time.Time     `json:"notifications_sent_at"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Topic     *Topic     `json:"topic,omitempty" gorm:"foreignKey:TopicID"`
	EmailLogs []EmailLog `json:"email_logs,omitempty" gorm:"foreignKey:ContentID"`
}

// TableName returns the table name for Content
func (Content) TableName() string {
	return "contents"
}
