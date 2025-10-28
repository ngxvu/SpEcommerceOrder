package repo

import (
	"context"
	"gorm.io/gorm"
	model "order/internal/models"
	pgGorm "order/internal/repositories/pg-gorm"
	"order/pkg/http/utils"
)

type OrderRepository struct {
	db pgGorm.PGInterface
}

func NewOrderRepository(newPgRepo pgGorm.PGInterface) *OrderRepository {
	return &OrderRepository{db: newPgRepo}
}

type OrderRepoInterface interface {
	CreateOrder(ctx context.Context, tx *gorm.DB, orderRequest *model.CreateOrderRequest) (*model.CreateOrderResponse, error)
}

func (a *OrderRepository) CreateOrder(ctx context.Context, tx *gorm.DB, orderRequest *model.CreateOrderRequest) (*model.CreateOrderResponse, error) {

	var cancel context.CancelFunc
	if tx == nil {
		tx, cancel = a.db.DBWithTimeout(ctx)
		defer cancel()
	}

	orderRecord := &model.Order{
		CustomerID:  orderRequest.CustomerID,
		TotalAmount: orderRequest.TotalAmount,
		Status:      orderRequest.Status,
	}

	if err := tx.Create(orderRecord).Error; err != nil {
		return nil, err
	}

	for _, item := range orderRequest.OrderItems {
		orderItem := &model.OrderItem{
			OrderID:   orderRecord.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: item.UniquePrice,
		}
		if err := tx.Create(orderItem).Error; err != nil {
			return nil, err
		}
	}

	response := &model.CreateOrderResponse{
		Meta: utils.NewMetaData(ctx),
		Data: "Order created successfully",
	}

	return response, nil
}
