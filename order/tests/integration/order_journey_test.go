package integration

import (
	"context"
	"fmt"
	"net"
	"testing"

	"gophertrade/order/internal/application"
	"gophertrade/order/internal/domain"
	infra "gophertrade/order/internal/infrastructure/grpc"
	"gophertrade/order/internal/infrastructure/persistence/postgres"
	inventoryv1 "gophertrade/proto/inventory/v1"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type mockInventoryService struct {
	inventoryv1.UnimplementedInventoryServiceServer
	products map[string]*inventoryv1.GetProductResponse
	updates  []*inventoryv1.StockUpdateItem
}

func (m *mockInventoryService) GetProduct(ctx context.Context, req *inventoryv1.GetProductRequest) (*inventoryv1.GetProductResponse, error) {
	p, ok := m.products[req.ProductId]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return p, nil
}

func (m *mockInventoryService) BatchUpdateStock(ctx context.Context, req *inventoryv1.BatchUpdateStockRequest) (*inventoryv1.BatchUpdateStockResponse, error) {
	m.updates = append(m.updates, req.Updates...)
	return &inventoryv1.BatchUpdateStockResponse{Success: true}, nil
}

type mockPublisher struct{}

func (m *mockPublisher) PublishOrderCreated(ctx context.Context, order *domain.Order) error {
	return nil
}

func TestOrderCreationJourney(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	// 1. Setup Postgres
	pool, cleanup := setupPostgres(t)
	defer cleanup()

	// 2. Setup Mock Inventory Service
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	
	mockInv := &mockInventoryService{
		products: make(map[string]*inventoryv1.GetProductResponse),
	}
	
	productID := uuid.New().String()
	mockInv.products[productID] = &inventoryv1.GetProductResponse{
		ProductId: productID,
		Name:      "Test Product",
		PriceCents: 1000,
	}

	s := grpc.NewServer()
	inventoryv1.RegisterInventoryServiceServer(s, mockInv)
	go s.Serve(lis)
	defer s.Stop()

	// 3. Setup Order Service
	repo := postgres.NewOrderRepository(pool)
	invClient, err := infra.NewInventoryClient(lis.Addr().String())
	require.NoError(t, err)

	svc := application.NewOrderService(repo, invClient, &mockPublisher{})

	// 4. Create Order
	req := application.CreateOrderRequest{
		Items: []struct {
			ProductID string
			Quantity  int32
		}{
			{ProductID: productID, Quantity: 2},
		},
	}

	order, err := svc.CreateOrder(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, int64(2000), order.TotalPrice)

	// 5. Verify Repository
	found, err := repo.GetByID(ctx, order.ID)
	assert.NoError(t, err)
	assert.Equal(t, order.ID, found.ID)
	assert.Len(t, found.Items, 1)

	// 6. Verify gRPC call
	assert.Len(t, mockInv.updates, 1)
	assert.Equal(t, productID, mockInv.updates[0].ProductId)
	assert.Equal(t, int32(-2), mockInv.updates[0].QuantityDelta)
}
