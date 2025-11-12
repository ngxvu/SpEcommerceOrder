package services

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
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

	tracer := otel.Tracer("order/service")
	ctx, span := tracer.Start(ctx, "OrderService.CreateOrder",
		trace.WithAttributes(attribute.String("customer_id", orderRequest.CustomerID.String())))
	defer span.End()

	span.AddEvent("begin tx")
	tx := oS.newPgRepo.GetRepo().Begin()
	defer tx.Rollback()

	// Create order in DB
	createOrderResp, err := oS.repo.CreateOrder(ctx, tx, &orderRequest)
	if err != nil {
		// tracer
		span.RecordError(err)
		span.SetStatus(codes.Error, "create order failed")
		return nil, err

		err = app_errors.AppError(app_errors.StatusInternalServerError, app_errors.StatusInternalServerError)
		// logger
		logger.LogError(log, err, "failed to create order")
		return nil, err
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

	// prepare outbox payload...
	span.AddEvent("create outbox", trace.WithAttributes(attribute.String("aggregate_id", createOrderResp.Data.OrderID.String())))
	if err = oS.outboxRepo.CreateOutbox(ctx, tx, outbox); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "create outbox failed")

		err = app_errors.AppError(app_errors.StatusInternalServerError, "create outboxRepo failed")
		logger.LogError(log, err, "failed to create outboxRepo")
		return nil, err
	}

	if err = tx.Commit().Error; err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "commit failed")

		err = app_errors.AppError(app_errors.StatusInternalServerError, app_errors.StatusInternalServerError)
		logger.LogError(log, err, "failed to commit tx")
		return nil, err
	}

	span.SetStatus(codes.Ok, "created")
	return createOrderResp, nil
}
