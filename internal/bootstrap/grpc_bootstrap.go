package bootstrap

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	paymentclient "order_service/internal/clients/payment"
	"order_service/internal/grpc/handlers"
	"order_service/internal/grpc/server"
	repo "order_service/internal/repositories"
	"order_service/internal/services"
	"strconv"
	"time"
)

func StartGRPC(app *App) (*server.GRPCServer, error) {

	grpcPort, err := strconv.Atoi(app.Config.GRPCPort)
	if err != nil || grpcPort == 0 {
		grpcPort = 50051
	}
	httpPort, err := strconv.Atoi(app.Config.HTTPPort)
	if err != nil || httpPort == 0 {
		httpPort = 8080
	}

	grpcAddr := fmt.Sprintf(":%d", grpcPort)
	httpAddr := fmt.Sprintf(":%d", httpPort)

	newPgRepo := app.PGRepo
	orderRepo := repo.NewOrderRepository(newPgRepo)
	outboxRepo := repo.NewOutboxRepository(newPgRepo)
	// create gRPC connection to payment service and build payment client
	paymentAddr := app.Config.PaymentServiceAddr
	if paymentAddr == "" {
		paymentAddr = "payment-service:50051"
	}

	dialCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(dialCtx, paymentAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	paymentClient := paymentclient.NewPaymentGRPCClient(conn)
	orderService := services.NewOrderService(orderRepo, newPgRepo, paymentClient, outboxRepo)
	handler := handlers.NewOrderHandler(*orderService)

	grpcServer := server.NewGRPCServer(handler, grpcAddr, httpAddr)

	ctx := context.Background()

	go func() {
		if err := grpcServer.Run(ctx); err != nil {
			panic(err)
		}
	}()

	return grpcServer, nil
}
