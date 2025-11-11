package services

import (
	"context"
	"encoding/json"
	paymentclient "order/internal/clients/payment"
	model "order/internal/models"
	"order/internal/repositories"
	pgGorm "order/internal/repositories/pg-gorm"
	"order/pkg/core/logger"
	"order/pkg/http/utils"
	"order/pkg/http/utils/app_errors"
	"time"
)

type OrderService struct {
	repo      repo.OrderRepoInterface
	newPgRepo pgGorm.PGInterface
	payment   paymentclient.PaymentClient
	outbox    *repo.OutboxRepository
}

type OrderServiceInterface interface {
	CreateOrder(ctx context.Context, orderRequest model.CreateOrderRequest) (*model.CreateOrderResponse, error)
}

func NewOrderService(
	repo repo.OrderRepoInterface,
	newRepo pgGorm.PGInterface,
	payment paymentclient.PaymentClient,
	outbox *repo.OutboxRepository,
) *OrderService {
	return &OrderService{
		repo:      repo,
		newPgRepo: newRepo,
		payment:   payment,
		outbox:    outbox,
	}
}

func (oS *OrderService) CreateOrder(ctx context.Context, orderRequest model.CreateOrderRequest) (*model.CreateOrderResponse, error) {
	log := logger.WithTag("OrderService|CreateOrder")

	tx := oS.newPgRepo.GetRepo().Begin()
	defer tx.Rollback()

	// Create order in DB
	createOrderResp, err := oS.repo.CreateOrder(ctx, tx, &orderRequest)
	if err != nil {
		logger.LogError(log, err, "failed to create order")
		return nil, app_errors.AppError(app_errors.StatusInternalServerError, "create order failed")
	}

	b, err := json.Marshal(createOrderResp)
	if err != nil {
		logger.LogError(log, err, "failed to marshal create order response")
		return nil, app_errors.AppError(app_errors.StatusInternalServerError, app_errors.StatusInternalServerError)
	}

	outbox := &model.OutboxEvent{
		AggregateType: "order",
		AggregateID:   createOrderResp.Data.OrderID,
		EventType:     "payment_required",
		Payload:       json.RawMessage(b),
		Status:        model.OutboxStatusPending,
		Attempts:      0,
		NextAttemptAt: time.Now(),
	}

	err = oS.outbox.CreateOutBox(ctx, tx, outbox)
	if err != nil {
		logger.LogError(log, err, "failed to create outbox")
		return nil, app_errors.AppError(app_errors.StatusInternalServerError, app_errors.StatusInternalServerError)
	}

	if err := tx.Commit().Error; err != nil {
		logger.LogError(log, err, "failed to commit tx")
		return nil, app_errors.AppError(app_errors.StatusInternalServerError, app_errors.StatusInternalServerError)
	}

	return createOrderResp, nil
}
