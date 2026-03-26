package integration

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"gophertrade/order/internal/domain"
	"gophertrade/order/internal/infrastructure/persistence/postgres"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupPostgres(t *testing.T) (*pgxpool.Pool, func()) {
	ctx := context.Background()

	migrationPath, err := filepath.Abs("../../internal/infrastructure/persistence/postgres/migrations/")
	require.NoError(t, err)

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("order"),
		tcpostgres.WithUsername("user"),
		tcpostgres.WithPassword("password"),
		tcpostgres.WithInitScripts(filepath.Join(migrationPath, "001_create_orders.up.sql")),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(10*time.Second)),
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

func TestOrderRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	pool, cleanup := setupPostgres(t)
	defer cleanup()

	repo := postgres.NewOrderRepository(pool)
	ctx := context.Background()

	productID1 := uuid.New()
	productID2 := uuid.New()

	items := []domain.OrderItem{
		{ProductID: productID1, Quantity: 2, UnitPrice: 1000},
		{ProductID: productID2, Quantity: 1, UnitPrice: 500},
	}

	order, err := domain.NewOrder(items)
	require.NoError(t, err)

	t.Run("Create and Get Order", func(t *testing.T) {
		err := repo.Create(ctx, order)
		assert.NoError(t, err)

		found, err := repo.GetByID(ctx, order.ID)
		assert.NoError(t, err)
		assert.Equal(t, order.ID, found.ID)
		assert.Equal(t, order.Status, found.Status)
		assert.Equal(t, order.TotalPrice, found.TotalPrice)
		assert.Len(t, found.Items, 2)
		
		// Map items by product ID for easy comparison
		foundItems := make(map[uuid.UUID]domain.OrderItem)
		for _, item := range found.Items {
			foundItems[item.ProductID] = item
		}

		assert.Equal(t, int32(2), foundItems[productID1].Quantity)
		assert.Equal(t, int64(1000), foundItems[productID1].UnitPrice)
		assert.Equal(t, int32(1), foundItems[productID2].Quantity)
		assert.Equal(t, int64(500), foundItems[productID2].UnitPrice)
	})

	t.Run("Get Non-Existent", func(t *testing.T) {
		_, err := repo.GetByID(ctx, uuid.New())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "order not found")
	})
}
