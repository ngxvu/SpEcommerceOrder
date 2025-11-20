package services

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"order/internal/models"
	repo "order/internal/repositories"
	"time"
)

var (
	ErrNoActivePromotion      = errors.New("no active promotion")
	ErrOrderBelowMinValue     = errors.New("order below minimum value for promotion")
	ErrCustomerAlreadyReward  = errors.New("customer already received promotion")
	ErrPromotionCustomerLimit = errors.New("promotion customer limit reached")
	ErrPromotionTotalExhaust  = errors.New("promotion total rewards exhausted")
)

type PromotionService struct {
	promoRepo  repo.PromotionRepoInterface
	outboxRepo repo.OutboxRepoInterface
	orderRepo  repo.OrderRepoInterface
	nowFunc    func() time.Time
}

func NewPromotionService(p repo.PromotionRepoInterface, o repo.OutboxRepoInterface, ord repo.OrderRepoInterface) *PromotionService {
	return &PromotionService{
		promoRepo:  p,
		outboxRepo: o,
		orderRepo:  ord,
		nowFunc:    time.Now,
	}
}

type PromotionServiceInterface interface {
	HandlePromotion(ctx context.Context, evt models.PromotionRewardEvent) error
}

// HandlePromotion processes a PromotionRewardEvent (sent after payment authorized).
// It verifies campaign constraints, creates a PromotionReward record and an Outbox entry.
func (prom *PromotionService) HandlePromotion(ctx context.Context, evt models.PromotionRewardEvent) error {
	// parse order id
	orderID, err := uuid.Parse(evt.OrderID)
	if err != nil {
		return err
	}

	// fetch order (must provide TotalAmount and CustomerID via orderRepo)
	order, err := prom.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	// get active promotion (assumes single active campaign; adjust to name if needed)
	now := prom.nowFunc()
	promo, err := prom.promoRepo.GetActivePromotion(ctx, now)
	if err != nil {
		return ErrNoActivePromotion
	}

	// check order amount against promotion minimum
	if order.TotalAmount < promo.MinOrderValue {
		return ErrOrderBelowMinValue
	}

	// check per-customer: only one reward per customer
	already, err := prom.promoRepo.HasCustomerReceived(ctx, promo.ID, order.CustomerID)
	if err != nil {
		return err
	}
	if already {
		return ErrCustomerAlreadyReward
	}

	// check customer limit (first N customers who satisfy conditions)
	customerCount, err := prom.promoRepo.CountDistinctCustomers(ctx, promo.ID)
	if err != nil {
		return err
	}

	if promo.CustomerLimit > 0 && int(customerCount) >= promo.CustomerLimit {
		return ErrPromotionCustomerLimit
	}

	// check total rewards given (global cap)
	totalGiven, err := prom.promoRepo.CountRewards(ctx, promo.ID)
	if err != nil {
		return err
	}
	// Interpret RewardLimit on PromotionConfig as total available for campaign.
	if promo.RewardLimit > 0 && int(totalGiven) >= promo.RewardLimit {
		return ErrPromotionTotalExhaust
	}

	// create PromotionReward
	reward := &models.PromotionReward{
		PromotionConfigID: promo.ID,
		OrderID:           order.ID,
		CustomerID:        order.CustomerID,
		ReceivedAt:        prom.nowFunc(),
	}
	if err := prom.promoRepo.CreateReward(ctx, reward); err != nil {
		return err
	}

	// create outbox event (reliable publish)
	outPayload, _ := json.Marshal(struct {
		RewardID string `json:"reward_id"`
		OrderID  string `json:"order_id"`
	}{
		RewardID: reward.ID.String(),
		OrderID:  order.ID.String(),
	})
	outbox := &models.Outbox{
		EventID:       uuid.New(),
		EventType:     "promotion.reward.created",
		Payload:       string(outPayload),
		AggregateType: "promotion_reward",
		AggregateID:   reward.ID,
		Status:        models.OutboxStatusPending,
		Attempts:      0,
		NextAttemptAt: prom.nowFunc(),
	}
	if err = prom.outboxRepo.CreateOutbox(ctx, nil, outbox); err != nil {
		// NOTE: reward already persisted; outbox failure should be handled (retry by caller) — return error so it can be retried.
		return err
	}

	return nil
}
