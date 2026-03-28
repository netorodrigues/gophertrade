package integration

import (
	"context"
	"net"
	"testing"

	"gophertrade/inventory/internal/api/grpc"
	"gophertrade/inventory/internal/application"
	"gophertrade/inventory/internal/domain"
	inventoryv1 "gophertrade/proto/inventory/v1"

	"github.com/google/uuid"
	real_grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

// mock repo and publisher for grpc tests
type mockRepo struct {
	products map[uuid.UUID]*domain.Product
}

func (m *mockRepo) Create(ctx context.Context, p *domain.Product) error {
	m.products[p.ID] = p
	return nil
}
func (m *mockRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	if p, ok := m.products[id]; ok {
		copy := *p
		return &copy, nil
	}
	return nil, domain.ErrProductNotFound
}
func (m *mockRepo) UpdateStock(ctx context.Context, id uuid.UUID, delta int64, expectedVersion int32) error {
	p, ok := m.products[id]
	if !ok {
		return domain.ErrProductNotFound
	}
	if p.Version != expectedVersion {
		return domain.ErrConflict
	}
	err := p.UpdateStock(delta)
	if err != nil {
		return err
	}
	p.Version++
	m.products[id] = p
	return nil
}

func (m *mockRepo) BatchUpdateStock(ctx context.Context, updates []domain.StockUpdateItem) error {
	for _, u := range updates {
		err := m.UpdateStock(ctx, u.ProductID, u.Delta, u.Version)
		if err != nil {
			return err
		}
	}
	return nil
}

type mockPub struct{}

func (m *mockPub) PublishProductCreated(ctx context.Context, p *domain.Product) error { return nil }
func (m *mockPub) PublishStockUpdated(ctx context.Context, p *domain.Product) error   { return nil }

func setupGRPCServer() (*real_grpc.Server, *bufconn.Listener) {
	lis := bufconn.Listen(1024 * 1024)
	server := real_grpc.NewServer()

	repo := &mockRepo{products: make(map[uuid.UUID]*domain.Product)}
	pub := &mockPub{}
	service := application.NewProductService(repo, pub)
	handler := grpc.NewProductHandler(service)

	inventoryv1.RegisterInventoryServiceServer(server, handler)

	go func() {
		if err := server.Serve(lis); err != nil {
			panic(err)
		}
	}()

	return server, lis
}

func bufDialer(lis *bufconn.Listener) func(context.Context, string) (net.Conn, error) {
	return func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
}

func TestGRPC_CreateProduct(t *testing.T) {
	server, lis := setupGRPCServer()
	defer server.Stop()

	ctx := context.Background()
	conn, err := real_grpc.DialContext(ctx, "bufnet", real_grpc.WithContextDialer(bufDialer(lis)), real_grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := inventoryv1.NewInventoryServiceClient(conn)

	req := &inventoryv1.CreateProductRequest{
		Name:         "Test Product",
		PriceCents:   1000,
		InitialStock: 50,
	}

	resp, err := client.CreateProduct(ctx, req)
	if err != nil {
		t.Fatalf("CreateProduct failed: %v", err)
	}

	if resp.ProductId == "" {
		t.Fatal("Expected product ID in response, got empty")
	}

	// Verify creation
	getReq := &inventoryv1.GetProductRequest{ProductId: resp.ProductId}
	getResp, err := client.GetProduct(ctx, getReq)
	if err != nil {
		t.Fatalf("GetProduct failed: %v", err)
	}

	if getResp.Name != "Test Product" {
		t.Errorf("Expected product name 'Test Product', got '%s'", getResp.Name)
	}
	if getResp.StockQuantity != 50 {
		t.Errorf("Expected product stock 50, got '%d'", getResp.StockQuantity)
	}
}

func TestGRPC_UpdateStock(t *testing.T) {
	server, lis := setupGRPCServer()
	defer server.Stop()

	ctx := context.Background()
	conn, err := real_grpc.DialContext(ctx, "bufnet", real_grpc.WithContextDialer(bufDialer(lis)), real_grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := inventoryv1.NewInventoryServiceClient(conn)

	// First create a product
	reqCreate := &inventoryv1.CreateProductRequest{
		Name:         "Stock Product",
		PriceCents:   2000,
		InitialStock: 10,
	}
	respCreate, err := client.CreateProduct(ctx, reqCreate)
	if err != nil {
		t.Fatalf("CreateProduct failed: %v", err)
	}
	prodID := respCreate.ProductId

	// Update stock (decrement)
	reqUpdate := &inventoryv1.UpdateStockRequest{
		ProductId:     prodID,
		QuantityDelta: -5,
	}
	respUpdate, err := client.UpdateStock(ctx, reqUpdate)
	if err != nil {
		t.Fatalf("UpdateStock failed: %v", err)
	}

	if !respUpdate.Success {
		t.Error("Expected UpdateStock to succeed")
	}

	// Verify final stock
	reqGet := &inventoryv1.GetProductRequest{
		ProductId: prodID,
	}
	respGet, err := client.GetProduct(ctx, reqGet)
	if err != nil {
		t.Fatalf("GetProduct failed: %v", err)
	}
	if respGet.StockQuantity != 5 {
		t.Errorf("Expected stock 5, got %d", respGet.StockQuantity)
	}
}