package application

import (
	"context"
	"fmt"

	"gophertrade/order/internal/domain"
	inventoryv1 "gophertrade/proto/inventory/v1"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type InventoryService interface {
	GetProduct(ctx context.Context, productID string) (*inventoryv1.GetProductResponse, error)
	BatchUpdateStock(ctx context.Context, items map[string]int32) error
}

type OrderEventPublisher interface {
	PublishOrderCreated(ctx context.Context, order *domain.Order) error
}

type OrderService struct {
	repo             domain.OrderRepository
	inventoryService InventoryService
	publisher        OrderEventPublisher
	tracer           trace.Tracer
}

func NewOrderService(repo domain.OrderRepository, inventoryService InventoryService, publisher OrderEventPublisher) *OrderService {
	return &OrderService{
		repo:             repo,
		inventoryService: inventoryService,
		publisher:        publisher,
		tracer:           otel.Tracer("order-service"),
	}
}

type CreateOrderRequest struct {
	Items []struct {
		ProductID string
		Quantity  int32
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, req CreateOrderRequest) (*domain.Order, error) {
	ctx, span := s.tracer.Start(ctx, "CreateOrder")
	defer span.End()

	var orderItems []domain.OrderItem
	stockUpdates := make(map[string]int32)

	for _, reqItem := range req.Items {
		productID, err := uuid.Parse(reqItem.ProductID)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("invalid product ID %s: %w", reqItem.ProductID, err)
		}

		// Fetch product to get current price (snapshot)
		resp, err := s.inventoryService.GetProduct(ctx, reqItem.ProductID)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to fetch product %s: %w", reqItem.ProductID, err)
		}

		orderItems = append(orderItems, domain.OrderItem{
			ProductID: productID,
			Quantity:  reqItem.Quantity,
			UnitPrice: resp.PriceCents,
		})

		// Prepare stock update (negative delta)
		stockUpdates[reqItem.ProductID] = -reqItem.Quantity
	}

	// 1. Create Order entity (calculates total)
	order, err := domain.NewOrder(orderItems)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to create order entity: %w", err)
	}

	// 2. Decrement stock in Inventory service (Atomic via gRPC)
	if err := s.inventoryService.BatchUpdateStock(ctx, stockUpdates); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to update stock: %w", err)
	}

	// 3. Persist Order to DB
	if err := s.repo.Create(ctx, order); err != nil {
		span.RecordError(err)
		// CRITICAL: In a real system, we should trigger a compensating action here
		// because stock was already decremented.
		return nil, fmt.Errorf("failed to persist order: %w", err)
	}

	// 4. Publish Event
	if s.publisher != nil {
		_ = s.publisher.PublishOrderCreated(ctx, order)
	}

	return order, nil
}

func (s *OrderService) GetOrder(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	ctx, span := s.tracer.Start(ctx, "GetOrder")
	defer span.End()

	return s.repo.GetByID(ctx, id)
}
