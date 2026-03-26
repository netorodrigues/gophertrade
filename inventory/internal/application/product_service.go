package application

import (
	"context"
	"fmt"

	"gophertrade/inventory/internal/domain"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type ProductEventPublisher interface {
	PublishProductCreated(ctx context.Context, p *domain.Product) error
	PublishStockUpdated(ctx context.Context, p *domain.Product) error
}

type ProductService struct {
	repo      domain.ProductRepository
	publisher ProductEventPublisher
	tracer    trace.Tracer
}

func NewProductService(repo domain.ProductRepository, publisher ProductEventPublisher) *ProductService {
	return &ProductService{
		repo:      repo,
		publisher: publisher,
		tracer:    otel.Tracer("inventory-service"),
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, name string, price int64, initialStock int64) (*domain.Product, error) {
	ctx, span := s.tracer.Start(ctx, "CreateProduct")
	defer span.End()

	product := domain.NewProduct(name, price, initialStock)
	if err := s.repo.Create(ctx, product); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	if s.publisher != nil {
		_ = s.publisher.PublishProductCreated(ctx, product)
	}

	return product, nil
}

func (s *ProductService) UpdateStock(ctx context.Context, id uuid.UUID, delta int64) error {
	ctx, span := s.tracer.Start(ctx, "UpdateStock")
	defer span.End()

	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		return err
	}

	if err := product.UpdateStock(delta); err != nil {
		span.RecordError(err)
		return err
	}

	if err := s.repo.UpdateStock(ctx, id, delta, product.Version); err != nil {
		span.RecordError(err)
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
	ctx, span := s.tracer.Start(ctx, "BatchUpdateStock")
	defer span.End()

	repoUpdates := make([]domain.StockUpdateItem, len(updates))
	for i, u := range updates {
		p, err := s.repo.GetByID(ctx, u.ProductID)
		if err != nil {
			span.RecordError(err)
			return err
		}
		repoUpdates[i] = domain.StockUpdateItem{
			ProductID: u.ProductID,
			Delta:     u.Delta,
			Version:   p.Version,
		}
	}

	if err := s.repo.BatchUpdateStock(ctx, repoUpdates); err != nil {
		span.RecordError(err)
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
	ctx, span := s.tracer.Start(ctx, "GetProduct")
	defer span.End()

	return s.repo.GetByID(ctx, id)
}
