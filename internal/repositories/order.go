package repo

import (
	"context"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"order/internal/events"
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
	UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status events.PaymentStatus) error
	GetByID(ctx context.Context, orderID uuid.UUID) (*model.Order, error)
}

func (a *OrderRepository) GetByID(ctx context.Context, orderID uuid.UUID) (*model.Order, error) {
	var cancel context.CancelFunc
	tx, cancel := a.db.DBWithTimeout(ctx)
	defer cancel()

	var order model.Order
	if err := tx.Preload("OrderItems").Where("id = ?", orderID).First(&order).Error; err != nil {
		return nil, err
	}

	return &order, nil
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
		Data: model.CreateOrderResponseData{
			OrderID:     orderRecord.ID,
			CustomerID:  orderRecord.CustomerID,
			TotalAmount: orderRecord.TotalAmount,
			Status:      orderRecord.Status,
		},
	}

	return response, nil
}

func (a *OrderRepository) UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status events.PaymentStatus) error {

	var cancel context.CancelFunc
	tx, cancel := a.db.DBWithTimeout(ctx)
	defer cancel()

	if err := tx.Model(&model.Order{}).Where("id = ?", orderID).Update("status", string(status)).Error; err != nil {
		return err
	}

	return nil

}
