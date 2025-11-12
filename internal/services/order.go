package services

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	paymentclient "order/internal/clients/payment"
	model "order/internal/models"
	"order/internal/repositories"
	pgGorm "order/internal/repositories/pg-gorm"
	"order/pkg/core/logger"
	"order/pkg/http/utils/app_errors"
	"order/pkg/proto/paymentpb"
	"time"
)

type OrderService struct {
	repo       repo.OrderRepoInterface
	newPgRepo  pgGorm.PGInterface
	payment    paymentclient.PaymentClient
	outboxRepo *repo.OutboxRepository
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
		repo:       repo,
		newPgRepo:  newRepo,
		payment:    payment,
		outboxRepo: outbox,
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

	payReq := &paymentpb.PayRequest{
		OrderId:    createOrderResp.Data.OrderID.String(),
		CustomerId: createOrderResp.Data.CustomerID.String(),
		Amount:     createOrderResp.Data.TotalAmount,
		Status:     createOrderResp.Data.Status,
	}

	bs, _ := json.Marshal(payReq)

	outbox := &model.Outbox{
		EventID:       uuid.New(),
		EventType:     "payment_required",
		AggregateType: "order",
		AggregateID:   createOrderResp.Data.OrderID,
		Payload:       string(bs),
		Status:        model.OutboxStatusPending,
		Attempts:      0,
		NextAttemptAt: time.Now(),
	}

	// write outboxRepo in same tx so it's transactional with order creation
	if err := oS.outboxRepo.CreateOutbox(ctx, tx, outbox); err != nil {
		logger.LogError(log, err, "failed to create outboxRepo")
		return nil, app_errors.AppError(app_errors.StatusInternalServerError, "create outboxRepo failed")
	}

	// commit tx and return (no synchronous payment call)
	if err := tx.Commit().Error; err != nil {
		logger.LogError(log, err, "failed to commit tx")
		return nil, app_errors.AppError(app_errors.StatusInternalServerError, "commit failed")
	}

	return createOrderResp, nil
}
