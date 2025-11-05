package bootstrap

import (
	"context"
	"fmt"
	"order/internal/grpc/handlers"
	"order/internal/grpc/server"
	repo "order/internal/repositories"
	"order/internal/services"
	"strconv"
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
	orderService := services.NewOrderService(orderRepo, newPgRepo)
	handler := handlers.NewOrderHandler(*orderService)

	grpcServer := server.NewGRPCServer(handler, grpcAddr, httpAddr)

	ctx := context.Background()

	go func() {
		if err := grpcServer.Run(ctx); err != nil {
			// choose appropriate logging/handling instead of panic in production
			panic(err)
		}
	}()

	return grpcServer, nil
}
