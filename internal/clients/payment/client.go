package paymentclient

import (
	"context"
	pbPayment "order_service/pkg/proto/paymentpb"
)

type PaymentClient interface {
	Pay(ctx context.Context, req *pbPayment.PayRequest) (*pbPayment.PayResponse, error)
}
