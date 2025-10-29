package handlers

import (
	"context"
	"order/internal/services"
	pb "order/pkg/proto/orderpb"
)

type OrderHandler struct {
	pb.UnimplementedOrderServiceServer
	service services.OrderService
}

func NewOrderHandler(s services.OrderService) *OrderHandler {
	return &OrderHandler{service: s}
}

func (h *OrderHandler) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	return &pb.CreateOrderResponse{}, nil
}
