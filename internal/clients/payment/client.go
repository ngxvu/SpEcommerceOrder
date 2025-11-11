package paymentclient

import (
	"context"
	pbPayment "order/pkg/proto/paymentpb"
)

type PaymentClient interface {
	Pay(ctx context.Context, req *pbPayment.PayRequest) (*pbPayment.PayResponse, error)
}
