package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrInvalidDelta      = errors.New("invalid stock delta")
	ErrProductNotFound   = errors.New("product not found")
	ErrConflict          = errors.New("version conflict")
)

type ProductRepository interface {
	Create(ctx context.Context, product *Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*Product, error)
	UpdateStock(ctx context.Context, id uuid.UUID, delta int64, expectedVersion int32) error
	BatchUpdateStock(ctx context.Context, updates []StockUpdateItem) error
}

type StockUpdateItem struct {
	ProductID uuid.UUID
	Delta     int64
	Version   int32
}

type Product struct {
	ID            uuid.UUID
	Name          string
	Price         int64 // Cents
	StockQuantity int64
	Version       int32
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewProduct(name string, price int64, initialStock int64) *Product {
	now := time.Now().UTC()
	return &Product{
		ID:            uuid.New(),
		Name:          name,
		Price:         price,
		StockQuantity: initialStock,
		Version:       1,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func (p *Product) UpdateStock(delta int64) error {
	newQuantity := p.StockQuantity + delta
	if newQuantity < 0 {
		return ErrInsufficientStock
	}
	p.StockQuantity = newQuantity
	p.UpdatedAt = time.Now().UTC()
	// Version is handled by Repository for optimistic locking
	return nil
}
