package services

import (
	"context"
	paymentclient "order/internal/clients/payment"
	model "order/internal/models"
	"order/internal/repositories"
	pgGorm "order/internal/repositories/pg-gorm"
	"order/pkg/core/logger"
	"order/pkg/http/utils/app_errors"
	pbPayment "order/pkg/proto/paymentpb"
)

type OrderService struct {
	repo      repo.OrderRepoInterface
	newPgRepo pgGorm.PGInterface
	payment   paymentclient.PaymentClient
}

type OrderServiceInterface interface {
	CreateOrder(ctx context.Context, orderRequest model.CreateOrderRequest) (*model.CreateOrderResponse, error)
}

func NewOrderService(
	repo repo.OrderRepoInterface,
	newRepo pgGorm.PGInterface,
	payment paymentclient.PaymentClient,
) *OrderService {
	return &OrderService{
		repo:      repo,
		newPgRepo: newRepo,
		payment:   payment,
	}
}

func (oS *OrderService) CreateOrder(ctx context.Context, orderRequest model.CreateOrderRequest) (*model.CreateOrderResponse, error) {
	log := logger.WithTag("OrderService|CreateOrder")

	tx := oS.newPgRepo.GetRepo().Begin()
	defer tx.Rollback()

	createOrderResponse, err := oS.repo.CreateOrder(ctx, tx, &orderRequest)
	if err != nil {
		logger.LogError(log, err, "failed to create order")
		err = app_errors.AppError(app_errors.StatusInternalServerError, app_errors.StatusInternalServerError)
		return nil, err
	}

	tx.Commit()

	payReq := &pbPayment.PayRequest{
		OrderId:    createOrderResponse.Data.OrderID.String(),
		CustomerId: createOrderResponse.Data.CustomerID.String(),
		Amount:     createOrderResponse.Data.TotalAmount,
		Status:     createOrderResponse.Data.Status,
	}
	_, err = oS.payment.Pay(ctx, payReq)
	if err != nil {
		logger.LogError(log, err, "failed to process payment")
		err = app_errors.AppError(app_errors.StatusInternalServerError, "payment processing failed")
	}

	return createOrderResponse, nil
}
