package performance

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"gophertrade/order/internal/application"
	"gophertrade/order/internal/domain"
	infra "gophertrade/order/internal/infrastructure/grpc"
	"gophertrade/order/internal/infrastructure/persistence/postgres"
	inventoryv1 "gophertrade/proto/inventory/v1"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"
)

type mockInventoryService struct {
	inventoryv1.UnimplementedInventoryServiceServer
}

func (m *mockInventoryService) GetProduct(ctx context.Context, req *inventoryv1.GetProductRequest) (*inventoryv1.GetProductResponse, error) {
	return &inventoryv1.GetProductResponse{
		ProductId:  req.ProductId,
		Name:       "Test Product",
		PriceCents: 1000,
	}, nil
}

func (m *mockInventoryService) BatchUpdateStock(ctx context.Context, req *inventoryv1.BatchUpdateStockRequest) (*inventoryv1.BatchUpdateStockResponse, error) {
	return &inventoryv1.BatchUpdateStockResponse{Success: true}, nil
}

type mockPublisher struct{}

func (m *mockPublisher) PublishOrderCreated(ctx context.Context, order *domain.Order) error {
	return nil
}

func setupPostgres(t *testing.T) (*pgxpool.Pool, func()) {
	ctx := context.Background()

	migrationPath, err := filepath.Abs("../../internal/infrastructure/persistence/postgres/migrations/")
	require.NoError(t, err)

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("order_perf"),
		tcpostgres.WithUsername("user"),
		tcpostgres.WithPassword("password"),
		tcpostgres.WithInitScripts(filepath.Join(migrationPath, "001_create_orders.up.sql")),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(15*time.Second)),
	)
	require.NoError(t, err)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)

	return pool, func() {
		pool.Close()
		pgContainer.Terminate(ctx)
	}
}

func TestOrderCreationPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test")
	}

	ctx := context.Background()

	// 1. Setup Postgres
	pool, cleanup := setupPostgres(t)
	defer cleanup()

	repo := postgres.NewOrderRepository(pool)

	// 2. Setup Mock Inventory Service
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	
	mockInv := &mockInventoryService{}
	s := grpc.NewServer()
	inventoryv1.RegisterInventoryServiceServer(s, mockInv)
	go s.Serve(lis)
	defer s.Stop()

	invClient, err := infra.NewInventoryClient(lis.Addr().String())
	require.NoError(t, err)

	svc := application.NewOrderService(repo, invClient, &mockPublisher{})

	productID := uuid.New().String()
	numRequests := 100
	latencies := make([]time.Duration, 0, numRequests)

	// Warm up
	for i := 0; i < 5; i++ {
		_, _ = svc.CreateOrder(ctx, application.CreateOrderRequest{
			Items: []struct {
				ProductID string
				Quantity  int32
			}{
				{ProductID: productID, Quantity: 1},
			},
		})
	}

	for i := 0; i < numRequests; i++ {
		req := application.CreateOrderRequest{
			Items: []struct {
				ProductID string
				Quantity  int32
			}{
				{ProductID: productID, Quantity: 1},
			},
		}

		start := time.Now()
		_, err := svc.CreateOrder(ctx, req)
		latencies = append(latencies, time.Since(start))
		require.NoError(t, err)
	}

	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	p95Index := int(float64(numRequests) * 0.95)
	p95Latency := latencies[p95Index]

	fmt.Printf("P95 Latency (with real DB): %v\n", p95Latency)
	if p95Latency > 200*time.Millisecond {
		t.Errorf("P95 latency %v exceeded 200ms", p95Latency)
	}
}
