package paymentclient

import (
	"context"
	"google.golang.org/grpc"
	pbPayment "order_service/pkg/proto/paymentpb"
)

type PaymentGRPCClient struct {
	client pbPayment.PaymentServiceClient
}

func NewPaymentGRPCClient(conn *grpc.ClientConn) *PaymentGRPCClient {
	return &PaymentGRPCClient{client: pbPayment.NewPaymentServiceClient(conn)}
}

func (c *PaymentGRPCClient) Pay(ctx context.Context, req *pbPayment.PayRequest) (*pbPayment.PayResponse, error) {
	return c.client.Pay(ctx, req)
}
