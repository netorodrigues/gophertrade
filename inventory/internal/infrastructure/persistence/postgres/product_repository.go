package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gophertrade/inventory/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepository struct {
	pool *pgxpool.Pool
}

func NewProductRepository(pool *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{pool: pool}
}

func (r *ProductRepository) Create(ctx context.Context, p *domain.Product) error {
	query := `
		INSERT INTO products (id, name, price_cents, stock_quantity, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.pool.Exec(ctx, query, p.ID, p.Name, p.Price, p.StockQuantity, p.Version, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}
	return nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	query := `
		SELECT id, name, price_cents, stock_quantity, version, created_at, updated_at
		FROM products
		WHERE id = $1
	`
	var p domain.Product
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Price, &p.StockQuantity, &p.Version, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	return &p, nil
}

func (r *ProductRepository) UpdateStock(ctx context.Context, id uuid.UUID, delta int64, expectedVersion int32) error {
	query := `
		UPDATE products
		SET stock_quantity = stock_quantity + $1,
		    version = version + 1,
		    updated_at = $2
		WHERE id = $3 AND version = $4 AND stock_quantity + $1 >= 0
	`
	tag, err := r.pool.Exec(ctx, query, delta, time.Now().UTC(), id, expectedVersion)
	if err != nil {
		return fmt.Errorf("failed to update stock: %w", err)
	}

	if tag.RowsAffected() == 0 {
		// Check if product exists or if it was a version conflict or insufficient stock
		var currentStock int64
		var currentVersion int32
		err := r.pool.QueryRow(ctx, "SELECT stock_quantity, version FROM products WHERE id = $1", id).Scan(&currentStock, &currentVersion)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.ErrProductNotFound
			}
			return fmt.Errorf("failed to verify update failure: %w", err)
		}

		if currentVersion != expectedVersion {
			return domain.ErrConflict
		}
		if currentStock+delta < 0 {
			return domain.ErrInsufficientStock
		}
		return fmt.Errorf("update failed for unknown reason")
	}

	return nil
}

func (r *ProductRepository) BatchUpdateStock(ctx context.Context, updates []domain.StockUpdateItem) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		UPDATE products
		SET stock_quantity = stock_quantity + $1,
		    version = version + 1,
		    updated_at = $2
		WHERE id = $3 AND version = $4 AND stock_quantity + $1 >= 0
	`

	for _, up := range updates {
		tag, err := tx.Exec(ctx, query, up.Delta, time.Now().UTC(), up.ProductID, up.Version)
		if err != nil {
			return fmt.Errorf("failed to update stock for product %s: %w", up.ProductID, err)
		}

		if tag.RowsAffected() == 0 {
			// Determine why it failed
			var currentStock int64
			var currentVersion int32
			err := tx.QueryRow(ctx, "SELECT stock_quantity, version FROM products WHERE id = $1", up.ProductID).Scan(&currentStock, &currentVersion)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return fmt.Errorf("product %s not found: %w", up.ProductID, domain.ErrProductNotFound)
				}
				return fmt.Errorf("failed to verify update failure for product %s: %w", up.ProductID, err)
			}

			if currentVersion != up.Version {
				return fmt.Errorf("version conflict for product %s: %w", up.ProductID, domain.ErrConflict)
			}
			if currentStock+up.Delta < 0 {
				return fmt.Errorf("insufficient stock for product %s: %w", up.ProductID, domain.ErrInsufficientStock)
			}
			return fmt.Errorf("update failed for product %s for unknown reason", up.ProductID)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
