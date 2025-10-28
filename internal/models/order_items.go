package models

import "github.com/google/uuid"

type OrderItem struct {
	BaseModel
	OrderID   uuid.UUID `json:"order_id" gorm:"type:uuid;not null;index"`
	ProductID uuid.UUID `json:"product_id" gorm:"type:uuid;not null;index"`
	Quantity  int       `json:"quantity" gorm:"type:int;not null"`
	UnitPrice float64   `json:"unit_price" gorm:"type:decimal(10,2);not null"`
	Order     Order     `json:"order" gorm:"foreignKey:OrderID;references:ID"`
}

func (OrderItem) TableName() string {
	return "order_items"
}
