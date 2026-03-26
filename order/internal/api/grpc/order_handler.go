package grpc

import (
	"context"

	"gophertrade/order/internal/application"
	orderv1 "gophertrade/proto/order/v1"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderHandler struct {
	orderv1.UnimplementedOrderServiceServer
	service *application.OrderService
}

func NewOrderHandler(service *application.OrderService) *OrderHandler {
	return &OrderHandler{
		service: service,
	}
}

func (h *OrderHandler) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (*orderv1.CreateOrderResponse, error) {
	appReq := application.CreateOrderRequest{}
	for _, item := range req.Items {
		appReq.Items = append(appReq.Items, struct {
			ProductID string
			Quantity  int32
		}{
			ProductID: item.ProductId,
			Quantity:  item.Quantity,
		})
	}

	order, err := h.service.CreateOrder(ctx, appReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create order: %v", err)
	}

	return &orderv1.CreateOrderResponse{
		OrderId:         order.ID.String(),
		TotalPriceCents: order.TotalPrice,
		Status:          string(order.Status),
	}, nil
}

func (h *OrderHandler) GetOrder(ctx context.Context, req *orderv1.GetOrderRequest) (*orderv1.GetOrderResponse, error) {
	id, err := uuid.Parse(req.OrderId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid order ID: %v", err)
	}

	order, err := h.service.GetOrder(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "order not found: %v", err)
	}

	var items []*orderv1.OrderItem
	for _, item := range order.Items {
		items = append(items, &orderv1.OrderItem{
			ProductId:      item.ProductID.String(),
			Quantity:       item.Quantity,
			UnitPriceCents: item.UnitPrice,
		})
	}

	return &orderv1.GetOrderResponse{
		OrderId:         order.ID.String(),
		TotalPriceCents: order.TotalPrice,
		Status:          string(order.Status),
		CreatedAt:       order.CreatedAt.Unix(),
		Items:           items,
	}, nil
}
