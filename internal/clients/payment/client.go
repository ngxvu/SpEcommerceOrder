package paymentclient

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"order/internal/metrics"
	pbPayment "order/pkg/proto/paymentpb"
)

type PaymentClient interface {
	Pay(ctx context.Context, req *pbPayment.PayRequest) (*pbPayment.PayResponse, error)
}

func newPaymentConn(target string) (*grpc.ClientConn, error) {
	return grpc.Dial(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(metrics.UnaryClientInterceptor("order_client")),
	)
}
