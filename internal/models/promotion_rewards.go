package models

import (
	"github.com/google/uuid"
	"time"
)

type PromotionReward struct {
	BaseModel
	PromotionConfigID uuid.UUID        `json:"promotion_config_id" gorm:"type:uuid;not null;index"`
	PromotionConfig   *PromotionConfig `json:"promotion_config" gorm:"foreignKey:PromotionConfigID;references:ID"`
	OrderID           uuid.UUID        `json:"order_id" gorm:"type:uuid;not null;index"`
	CustomerID        uuid.UUID        `json:"customer_id" gorm:"type:uuid;not null;index"`
	ReceivedAt        time.Time        `json:"received_at" gorm:"type:timestamp;not null"`
}

func (PromotionReward) TableName() string {
	return "promotion_rewards"
}
