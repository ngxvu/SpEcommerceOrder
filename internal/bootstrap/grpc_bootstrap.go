package bootstrap

import (
	"order/internal/grpc/handlers"
	"order/internal/grpc/server"
	repo "order/internal/repositories"
	"order/internal/services"
	"strconv"
)

func StartGRPC(app *App) {
	grpcPort, _ := strconv.Atoi(app.config.GRPCPort)

	newPgRepo := app.PGRepo
	orderRepo := repo.NewOrderRepository(newPgRepo)
	orderService := services.NewOrderService(orderRepo, newPgRepo)
	handler := handlers.NewOrderHandler(*orderService)
	grpcServer := server.NewGRPCServer(handler)

	if err := grpcServer.Run(grpcPort); err != nil {
		panic(err)
	}
}
