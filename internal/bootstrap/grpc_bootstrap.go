package bootstrap

import (
	"log"
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

	newPgRepo := app.PGRepo
	orderRepo := repo.NewOrderRepository(newPgRepo)
	orderService := services.NewOrderService(orderRepo, newPgRepo)

	handler := handlers.NewOrderHandler(*orderService)

	grpcServer := server.NewGRPCServer(handler)

	go func() {
		if err := grpcServer.Run(grpcPort); err != nil {
			// choose appropriate logging/handling instead of panic in production
			panic(err)
		}
	}()

	return grpcServer, nil
}
