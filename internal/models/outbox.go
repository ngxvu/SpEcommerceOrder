package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type OutboxStatus string

const (
	OutboxStatusPending OutboxStatus = "PENDING"
	OutboxStatusRetry   OutboxStatus = "RETRY"
	OutboxStatusDone    OutboxStatus = "DONE"
	OutboxStatusFailed  OutboxStatus = "FAILED"
)

type Outbox struct {
	BaseModel
	EventType     string       `gorm:"type:varchar(100);not null"`
	Payload       string       `gorm:"type:jsonb;not null"`
	AggregateType string       `gorm:"size:100;not null"`
	AggregateID   uuid.UUID    `gorm:"type:uuid;index"`
	Status        OutboxStatus `gorm:"size:20;not null;index"`
	Attempts      int          `gorm:"not null;default:0"`
	NextAttemptAt time.Time    `gorm:"index"`
	ProcessedAt   *time.Time
}

func (Outbox) TableName() string {
	return "outbox"
}

// BeforeCreate sets defaults
func (o *Outbox) BeforeCreate(tx *gorm.DB) (err error) {
	if o.NextAttemptAt.IsZero() {
		o.NextAttemptAt = time.Now()
	}
	return nil
}
