package models

import (
	"encoding/json"
	"github.com/jinzhu/gorm/dialects/postgres"
)

type Outbox struct {
	BaseModel
	EventType string          `json:"event_type" gorm:"type:varchar(100);index"`
	Payload   *postgres.Jsonb `json:"payload" gorm:"type:jsonb"`
	Process   bool            `json:"process" gorm:"type:boolean;not null;default:false;index"`
}

func (Outbox) TableName() string {
	return "outboxes"
}

type OutboxEvent struct {
	EventType string          `json:"event_type"`
	Payload   json.RawMessage `json:"payload"`
	Process   bool            `json:"process"`
}
