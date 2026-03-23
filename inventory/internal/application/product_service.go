package application

import (
	"context"
	"fmt"

	"gophertrade/inventory/internal/domain"

	"github.com/google/uuid"
)

type ProductEventPublisher interface {
	PublishProductCreated(ctx context.Context, p *domain.Product) error
	PublishStockUpdated(ctx context.Context, p *domain.Product) error
}

type ProductService struct {
	repo      domain.ProductRepository
	publisher ProductEventPublisher
}

func NewProductService(repo domain.ProductRepository, publisher ProductEventPublisher) *ProductService {
	return &ProductService{
		repo:      repo,
		publisher: publisher,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, name string, price int64, initialStock int64) (*domain.Product, error) {
	product := domain.NewProduct(name, price, initialStock)
	if err := s.repo.Create(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	if s.publisher != nil {
		_ = s.publisher.PublishProductCreated(ctx, product)
	}

	return product, nil
}

func (s *ProductService) UpdateStock(ctx context.Context, id uuid.UUID, delta int64) error {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := product.UpdateStock(delta); err != nil {
		return err
	}

	if err := s.repo.UpdateStock(ctx, id, delta, product.Version); err != nil {
		return err
	}

	// Fetch updated product for event
	updated, err := s.repo.GetByID(ctx, id)
	if err == nil && s.publisher != nil {
		_ = s.publisher.PublishStockUpdated(ctx, updated)
	}

	return nil
}

type StockUpdate struct {
	ProductID uuid.UUID
	Delta     int64
}

func (s *ProductService) BatchUpdateStock(ctx context.Context, updates []StockUpdate) error {
	repoUpdates := make([]domain.StockUpdateItem, len(updates))
	for i, u := range updates {
		p, err := s.repo.GetByID(ctx, u.ProductID)
		if err != nil {
			return err
		}
		repoUpdates[i] = domain.StockUpdateItem{
			ProductID: u.ProductID,
			Delta:     u.Delta,
			Version:   p.Version,
		}
	}

	if err := s.repo.BatchUpdateStock(ctx, repoUpdates); err != nil {
		return err
	}

	// Publish events for each product (could be optimized)
	if s.publisher != nil {
		for _, u := range updates {
			updated, err := s.repo.GetByID(ctx, u.ProductID)
			if err == nil {
				_ = s.publisher.PublishStockUpdated(ctx, updated)
			}
		}
	}

	return nil
}

func (s *ProductService) GetProduct(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	return s.repo.GetByID(ctx, id)
}
