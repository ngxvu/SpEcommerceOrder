package models

import "time"

type PromotionConfig struct {
	BaseModel
	Name            string            `json:"name" gorm:"type:varchar(100);not null;unique"`
	CustomerLimit   int               `json:"customer_limit" gorm:"type:int;not null;default:1"`
	RewardLimit     int               `json:"reward_limit" gorm:"type:int;not null;default:1"`
	MinOrderValue   float64           `json:"min_order_value" gorm:"type:decimal(10,2);not null;default:0.00"`
	IsActive        bool              `json:"is_active" gorm:"type:boolean;not null;default:true"`
	StartTime       time.Time         `json:"start_time" gorm:"type:timestamp;not null"`
	EndTime         time.Time         `json:"end_time" gorm:"type:timestamp;not null"`
	PromotionReward []PromotionReward `json:"promotion_rewards" gorm:"foreignKey:PromotionConfigID"`
	Order           []Order           `json:"orders" gorm:"foreignKey:PromotionConfigID;references:ID"`
}

func (PromotionConfig) TableName() string {
	return "promotion_configs"
}
