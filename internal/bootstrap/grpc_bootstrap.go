package bootstrap

//import (
//	"base-source/internal/grpc/handlers"
//	"base-source/internal/grpc/server"
//	"base-source/internal/services"
//)
//
//func StartGRPC() {
//	service := services.NewOrderService()
//	handler := handlers.NewOrderHandler(service)
//	grpcServer := server.New(handler)
//
//	if err := grpcServer.Run(50051); err != nil {
//		panic(err)
//	}
//}
