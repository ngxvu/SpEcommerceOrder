package handlers

import (
	"context"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"order/internal/models"
	"order/internal/services"
	pbOrder "order/pkg/proto"
)

type OrderHandler struct {
	pbOrder.UnimplementedOrderServiceServer
	service services.OrderServiceInterface
}

func NewOrderHandler(s services.OrderServiceInterface) *OrderHandler {
	return &OrderHandler{
		service: s}
}

func (h *OrderHandler) CreateOrder(ctx context.Context, req *pbOrder.CreateOrderRequest) (*pbOrder.CreateOrderResponse, error) {

	tracer := otel.Tracer("order/handler")
	ctx, span := tracer.Start(ctx, "OrderHandler.CreateOrder",
		trace.WithAttributes(attribute.String("grpc.method", "CreateOrder")))
	defer span.End()

	if req == nil {
		span.SetAttributes(attribute.Bool("invalid_request", true))
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
		span.RecordError(err)
		return nil, status.Errorf(codes.Internal, "create order failed: %v", err)
	}

	grpcResponse := &pbOrder.CreateOrderResponse{
		OrderId:     createOrderResp.Data.OrderID.String(),
		CustomerId:  createOrderResp.Data.CustomerID.String(),
		TotalAmount: createOrderResp.Data.TotalAmount,
		Status:      createOrderResp.Data.Status,
	}

	return grpcResponse, nil
}
