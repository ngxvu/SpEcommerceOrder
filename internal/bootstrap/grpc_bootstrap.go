package bootstrap

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"order/internal/grpc/clients/payment"
	"order/internal/grpc/handlers"
	"order/internal/grpc/server"
	"order/internal/metrics"
	repo "order/internal/repositories"
	"order/internal/services"
	"order/internal/workers"
	"strconv"
	"time"
)

func StartGRPC(app *AppSetup) (*server.GRPCServer, func(), error) {

	// grpcPort using for grpc server to transport gRPC requests
	grpcPort, err := strconv.Atoi(app.AppConfig.GRPCPort)
	if err != nil || grpcPort == 0 {
		grpcPort = 50051
	}
	grpcAddr := fmt.Sprintf(":%d", grpcPort)

	// httpPort using for http server of grpc gateway to transport HTTP requests that will be converted to gRPC requests
	httpPort, err := strconv.Atoi(app.AppConfig.HTTPPort)
	if err != nil || httpPort == 0 {
		httpPort = 8080
	}
	httpAddr := fmt.Sprintf(":%d", httpPort)

	newPgRepo := app.PGRepoInterface
	orderRepo := repo.NewOrderRepository(newPgRepo)
	outboxRepo := repo.NewOutboxRepository(newPgRepo)

	// create gRPC connection to payment service and build payment client
	paymentAddr := app.AppConfig.PaymentServiceAddr
	if paymentAddr == "" {
		paymentAddr = "localhost:50052"
	}

	// Dial context with timeout
	dialCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// conn is connection to payment gRPC service
	connection, err := grpc.DialContext(
		dialCtx,
		paymentAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),

		// outgoing from this order service to payment service will be measured here
		grpc.WithUnaryInterceptor(metrics.UnaryClientInterceptor("order")),
	)
	if err != nil {
		return nil, nil, err
	}

	paymentClient := paymentclient.NewPaymentGRPCClient(connection)
	orderService := services.NewOrderService(orderRepo, newPgRepo, paymentClient, outboxRepo)

	_, stopKafka := InitKafka(context.Background(), *orderService)

	handler := handlers.NewOrderHandler(orderService)

	grpcServer := server.NewGRPCServer(handler, grpcAddr, httpAddr)

	// start outbox worker properly (was previously discarded with `_ = ...`)
	ctx := context.Background()

	worker := workers.NewOutboxWorkerInit(newPgRepo, paymentClient)
	go worker.Run(ctx)

	go func() {
		if err := grpcServer.Run(ctx); err != nil {
			panic(err)
		}
	}()

	return grpcServer, stopKafka, nil
}
