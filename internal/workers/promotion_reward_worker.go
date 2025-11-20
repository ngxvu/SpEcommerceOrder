package workers

import (
	"context"
	"encoding/json"
	"log"

	"order/internal/models"
	"order/internal/services"
)

type PromotionRewardEventWorker struct {
	promotionService services.PromotionServiceInterface
}

func NewPromotionRewardWorker(promotionService services.PromotionServiceInterface) *PromotionRewardEventWorker {
	return &PromotionRewardEventWorker{promotionService: promotionService}
}

func (w *PromotionRewardEventWorker) Handle(ctx context.Context, data []byte) {
	var evt models.PromotionRewardEvent
	if err := json.Unmarshal(data, &evt); err != nil {
		log.Printf("failed to unmarshal promotion event: %v", err)
		return
	}

	if err := w.promotionService.HandlePromotion(ctx, evt); err != nil {
		log.Printf("promotion handling failed: %v", err)
		// don't ack or let consumer retry depending on consumer framework
		return
	}

	log.Printf("processed promotion event for order %s", evt.OrderID)
}
