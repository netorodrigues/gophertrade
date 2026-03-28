package grpc

import (
	"context"
	"errors"

	"gophertrade/inventory/internal/application"
	"gophertrade/inventory/internal/domain"
	inventoryv1 "gophertrade/proto/inventory/v1"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProductHandler struct {
	inventoryv1.UnimplementedInventoryServiceServer
	service *application.ProductService
}

func NewProductHandler(service *application.ProductService) *ProductHandler {
	return &ProductHandler{
		service: service,
	}
}


func (h *ProductHandler) GetProduct(ctx context.Context, req *inventoryv1.GetProductRequest) (*inventoryv1.GetProductResponse, error) {
	id, err := uuid.Parse(req.ProductId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid product id: %v", err)
	}

	product, err := h.service.GetProduct(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			return nil, status.Error(codes.NotFound, "product not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get product: %v", err)
	}

	return &inventoryv1.GetProductResponse{
		ProductId:     product.ID.String(),
		Name:          product.Name,
		PriceCents:    product.Price,
		StockQuantity: int32(product.StockQuantity),
		Version:       product.Version,
	}, nil
}


func (h *ProductHandler) BatchUpdateStock(ctx context.Context, req *inventoryv1.BatchUpdateStockRequest) (*inventoryv1.BatchUpdateStockResponse, error) {
	updates := make([]application.StockUpdate, len(req.Updates))
	for i, u := range req.Updates {
		id, err := uuid.Parse(u.ProductId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid product id %s: %v", u.ProductId, err)
		}
		updates[i] = application.StockUpdate{
			ProductID: id,
			Delta:     int64(u.QuantityDelta),
		}
	}

	err := h.service.BatchUpdateStock(ctx, updates)
	if err != nil {
		if errors.Is(err, domain.ErrInsufficientStock) || errors.Is(err, domain.ErrConflict) {
			return &inventoryv1.BatchUpdateStockResponse{Success: false, Message: err.Error()}, nil
		}
		return nil, status.Errorf(codes.Internal, "failed to batch update stock: %v", err)
	}

	return &inventoryv1.BatchUpdateStockResponse{
		Success: true,
	}, nil
}
