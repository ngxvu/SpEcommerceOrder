package handlers

import (
	"context"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"order/internal/models"
	"order/internal/services"
	pbOrder "order/pkg/proto"
)

type OrderHandler struct {
	pbOrder.UnimplementedOrderServiceServer
	service services.OrderService
}

func NewOrderHandler(s services.OrderService) *OrderHandler {
	return &OrderHandler{service: s}
}

func (h *OrderHandler) CreateOrder(ctx context.Context, req *pbOrder.CreateOrderRequest) (*pbOrder.CreateOrderResponse, error) {

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	// string req.CustomerId to uuid
	customerID, err := uuid.Parse(req.CustomerId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid customer id: %v", err)
	}

	var listOrderItems []models.CreateOrderItemRequest

	var orderItems models.CreateOrderItemRequest

	for _, v := range req.OrderItems {
		orderItems.UniquePrice = v.Price
		orderItems.ProductID, err = uuid.Parse(v.ProductId)
		orderItems.Quantity = int(v.Quantity)
		listOrderItems = append(listOrderItems, orderItems)
	}

	servicesRequest := models.CreateOrderRequest{
		CustomerID:  customerID,
		TotalAmount: req.TotalAmount,
		Status:      req.Status,
		OrderItems:  listOrderItems,
	}

	createOrderResp, err := h.service.CreateOrder(ctx, servicesRequest)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create order failed: %v", err)
	}

	grpcResponse := &pbOrder.CreateOrderResponse{
		Message: createOrderResp.Data,
	}

	return grpcResponse, nil
}
