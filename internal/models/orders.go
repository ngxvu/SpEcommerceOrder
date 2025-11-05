package models

import (
	"github.com/google/uuid"
	"order/pkg/http/utils"
)

type Order struct {
	BaseModel
	CustomerID        uuid.UUID        `json:"customer_id" gorm:"type:uuid;not null;index"`
	TotalAmount       float64          `json:"total_amount" gorm:"type:decimal(10,2);not null"`
	Status            string           `json:"status" gorm:"type:varchar(20);not null;index"`
	RewardGiven       bool             `json:"reward_given" gorm:"type:boolean;not null;default:false"`
	OrderItems        []OrderItem      `json:"order_items" gorm:"foreignKey:OrderID"`
	PromotionConfigID *uuid.UUID       `json:"promotion_config_id" gorm:"type:uuid;index"`
	PromotionConfig   *PromotionConfig `json:"promotion_config,omitempty" gorm:"foreignKey:PromotionConfigID;references:ID"`
}

func (Order) TableName() string {
	return "orders"
}

type CreateOrderRequest struct {
	CustomerID  uuid.UUID                `json:"customer_id" binding:"required,uuid"`
	TotalAmount float64                  `json:"total_amount" binding:"required,gt=0"`
	Status      string                   `json:"status" binding:"required,oneof=pending completed cancelled"`
	OrderItems  []CreateOrderItemRequest `json:"order_items" binding:"required"`
}

type CreateOrderItemRequest struct {
	ProductID   uuid.UUID `json:"product_id" binding:"required,uuid"`
	Quantity    int       `json:"quantity" binding:"required,gt=0"`
	UniquePrice float64   `json:"price" binding:"required,gt=0"`
}

type CreateOrderResponse struct {
	Meta *utils.MetaData         `json:"meta"`
	Data CreateOrderResponseData `json:"data"`
}

type CreateOrderResponseData struct {
	OrderID     uuid.UUID `json:"order_id"`
	CustomerID  uuid.UUID `json:"customer_id"`
	TotalAmount float64   `json:"total_amount"`
	Status      string    `json:"status"`
}
