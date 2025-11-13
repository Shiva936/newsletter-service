package daos

import (
	"time"

	"gorm.io/gorm"
)

// EmailLog represents an email delivery log in the database
type EmailLog struct {
	ID           uint           `json:"id" gorm:"primarykey"`
	SubscriberID uint           `json:"subscriber_id" gorm:"not null;index"`
	ContentID    uint           `json:"content_id" gorm:"not null;index"`
	EmailAddress string         `json:"email_address" gorm:"size:255;not null"`
	Subject      string         `json:"subject" gorm:"size:255;not null"`
	Body         string         `json:"body" gorm:"type:text;not null"`
	Status       string         `json:"status" gorm:"size:20;not null;index"`
	SentAt       *time.Time     `json:"sent_at"`
	ErrorMessage *string        `json:"error_message" gorm:"type:text"`
	RetryCount   int            `json:"retry_count" gorm:"default:0"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Subscriber *Subscriber `json:"subscriber,omitempty" gorm:"foreignKey:SubscriberID"`
	Content    *Content    `json:"content,omitempty" gorm:"foreignKey:ContentID"`
}

// TableName returns the table name for EmailLog
func (EmailLog) TableName() string {
	return "email_logs"
}

// EmailNotification represents an email notification to be sent
type EmailNotification struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}
