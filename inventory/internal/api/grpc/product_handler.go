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

func (h *ProductHandler) CreateProduct(ctx context.Context, req *inventoryv1.CreateProductRequest) (*inventoryv1.CreateProductResponse, error) {
	product, err := h.service.CreateProduct(ctx, req.Name, req.PriceCents, int64(req.InitialStock))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create product: %v", err)
	}

	return &inventoryv1.CreateProductResponse{
		ProductId: product.ID.String(),
	}, nil
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

func (h *ProductHandler) UpdateStock(ctx context.Context, req *inventoryv1.UpdateStockRequest) (*inventoryv1.UpdateStockResponse, error) {
	id, err := uuid.Parse(req.ProductId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid product id: %v", err)
	}

	err = h.service.UpdateStock(ctx, id, int64(req.QuantityDelta))
	if err != nil {
		if errors.Is(err, domain.ErrInsufficientStock) {
			return &inventoryv1.UpdateStockResponse{Success: false, Message: "insufficient stock"}, nil
		}
		if errors.Is(err, domain.ErrConflict) {
			return &inventoryv1.UpdateStockResponse{Success: false, Message: "version conflict"}, nil
		}
		return nil, status.Errorf(codes.Internal, "failed to update stock: %v", err)
	}

	// Fetch new version
	p, _ := h.service.GetProduct(ctx, id)
	var newVersion int32
	if p != nil {
		newVersion = p.Version
	}

	return &inventoryv1.UpdateStockResponse{
		Success:    true,
		NewVersion: newVersion,
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
