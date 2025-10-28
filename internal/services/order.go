package services

import (
	"context"
	model "order/internal/models"
	"order/internal/repositories"
	pgGorm "order/internal/repositories/pg-gorm"
	"order/pkg/core/logger"
	"order/pkg/http/utils/app_errors"
)

type OrderService struct {
	repo      repo.OrderRepoInterface
	newPgRepo pgGorm.PGInterface
}

type OrderServiceInterface interface {
	CreateOrder(ctx context.Context, orderRequest model.CreateOrderRequest) (*model.CreateOrderResponse, error)
}

func NewOrderService(repo repo.OrderRepoInterface, newRepo pgGorm.PGInterface) *OrderService {
	return &OrderService{
		repo:      repo,
		newPgRepo: newRepo,
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

	return createOrderResponse, nil
}
