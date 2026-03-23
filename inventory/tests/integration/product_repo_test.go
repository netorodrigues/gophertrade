package integration

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"gophertrade/inventory/internal/domain"
	"gophertrade/inventory/internal/infrastructure/persistence/postgres"

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
		tcpostgres.WithDatabase("inventory"),
		tcpostgres.WithUsername("user"),
		tcpostgres.WithPassword("password"),
		tcpostgres.WithInitScripts(filepath.Join(migrationPath, "001_create_products.up.sql")),
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

func TestProductRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	pool, cleanup := setupPostgres(t)
	defer cleanup()

	repo := postgres.NewProductRepository(pool)
	ctx := context.Background()

	product := domain.NewProduct("Gopher Plushie", 2500, 100)

	t.Run("Create and Get Product", func(t *testing.T) {
		err := repo.Create(ctx, product)
		assert.NoError(t, err)

		found, err := repo.GetByID(ctx, product.ID)
		assert.NoError(t, err)
		assert.Equal(t, product.Name, found.Name)
		assert.Equal(t, product.StockQuantity, found.StockQuantity)
		assert.Equal(t, product.Version, found.Version)
	})

	t.Run("Update Stock Success", func(t *testing.T) {
		err := repo.UpdateStock(ctx, product.ID, -10, product.Version)
		assert.NoError(t, err)

		found, err := repo.GetByID(ctx, product.ID)
		assert.NoError(t, err)
		assert.Equal(t, int64(90), found.StockQuantity)
		assert.Equal(t, product.Version+1, found.Version)

		// Update local product for next tests
		product = found
	})

	t.Run("Update Stock Conflict", func(t *testing.T) {
		err := repo.UpdateStock(ctx, product.ID, -10, product.Version-1)
		assert.ErrorIs(t, err, domain.ErrConflict)
	})

	t.Run("Update Stock Insufficient", func(t *testing.T) {
		err := repo.UpdateStock(ctx, product.ID, -100, product.Version)
		assert.ErrorIs(t, err, domain.ErrInsufficientStock)
	})

	t.Run("Get Non-Existent", func(t *testing.T) {
		_, err := repo.GetByID(ctx, uuid.New())
		assert.ErrorIs(t, err, domain.ErrProductNotFound)
	})
}
