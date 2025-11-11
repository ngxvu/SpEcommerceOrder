package models

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm/dialects/postgres"
	"gorm.io/gorm"
	"time"
)

type OutboxStatus string

const (
	OutboxStatusPending OutboxStatus = "pending"
	OutboxStatusRetry   OutboxStatus = "retry"
	OutboxStatusDone    OutboxStatus = "done"
	OutboxStatusFailed  OutboxStatus = "failed"
)

type Outbox struct {
	BaseModel
	EventType     string          `json:"event_type" gorm:"type:varchar(100);not null"`
	Payload       *postgres.Jsonb `json:"payload" gorm:"type:jsonb"`
	AggregateType string          `gorm:"size:100;not null"`
	AggregateID   uuid.UUID       `gorm:"type:uuid;index"`
	Status        OutboxStatus    `gorm:"size:20;not null;index"`
	Attempts      int             `gorm:"not null;default:0"`
	NextAttemptAt time.Time       `gorm:"index"`
	ProcessedAt   *time.Time
	UpdatedAt     time.Time
}

func (Outbox) TableName() string {
	return "outboxes"
}

type OutboxEvent struct {
	EventType     string          `json:"event_type"`
	Payload       json.RawMessage `json:"payload"`
	Process       bool            `json:"process"`
	AggregateType string          `json:"aggregate_type"`
	AggregateID   uuid.UUID       `json:"aggregate_id"`
	Status        OutboxStatus    `json:"status"`
	Attempts      int             `json:"attempts"`
	NextAttemptAt time.Time       `json:"next_attempt_at"`
	ProcessedAt   *time.Time      `json:"processed_at"`
}

// BeforeCreate sets defaults
func (o *Outbox) BeforeCreate(tx *gorm.DB) (err error) {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	if o.NextAttemptAt.IsZero() {
		o.NextAttemptAt = time.Now()
	}
	return nil
}
