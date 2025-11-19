package workers

import (
	"context"
	"encoding/json"
	"log"
	"order/internal/events"

	"github.com/google/uuid"
	"order/internal/models"
	"order/internal/services"
)

type PaymentEventWorker struct {
	orderService services.OrderServiceInterface
}

func NewPaymentEventWorker(orderService services.OrderServiceInterface) *PaymentEventWorker {
	return &PaymentEventWorker{orderService: orderService}
}

func (w *PaymentEventWorker) Handle(ctx context.Context, data []byte) {
	var evt models.PaymentAuthorizedEvent
	if err := json.Unmarshal(data, &evt); err != nil {
		log.Printf("failed to unmarshal payment event: %v", err)
		return
	}

	orderID, err := uuid.Parse(evt.OrderID)
	if err != nil {
		log.Printf("invalid order id in payment event: %v", err)
		return
	}

	if evt.Status != events.PaymentAuthorized {
		log.Printf("ignore payment event with status: %s", events.PaymentAuthorized)
		return
	}

	// gọi service để cập nhật trạng thái đơn hàng
	if err := w.orderService.UpdateOrderStatus(ctx, orderID, events.PaymentAuthorized); err != nil {
		log.Printf("failed to update order status: %v", err)
		return
	}

	log.Printf("order %s updated to Authorized by payment event", orderID.String())
}
