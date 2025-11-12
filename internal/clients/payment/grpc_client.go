package paymentclient

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	pbPayment "order/pkg/proto/paymentpb"
)

type PaymentGRPCClient struct {
	client pbPayment.PaymentServiceClient
}

func NewPaymentGRPCClient(conn *grpc.ClientConn) *PaymentGRPCClient {
	return &PaymentGRPCClient{client: pbPayment.NewPaymentServiceClient(conn)}
}

func (c *PaymentGRPCClient) Pay(ctx context.Context, req *pbPayment.PayRequest) (*pbPayment.PayResponse, error) {
	headers := map[string]string{}
	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(headers))
	md := metadata.New(headers)
	ctx = metadata.NewOutgoingContext(ctx, md)

	return c.client.Pay(ctx, req)
}
