package models

import "order/internal/events"

type PaymentAuthorizedEvent struct {
	PaymentID      string               `json:"payment_id"`
	OrderID        string               `json:"order_id"`
	IdempotencyKey string               `json:"idempotency_key"`
	Amount         float64              `json:"amount"`
	Status         events.PaymentStatus `json:"status"`
}
