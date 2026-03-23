package event

import (
	"context"
	"encoding/json"
	"fmt"

	"gophertrade/inventory/internal/application"
	"gophertrade/inventory/internal/domain"
	"gophertrade/inventory/internal/infrastructure/event/kafka"
)

type ProductEventPublisher struct {
	client *kafka.Client
}

func NewProductEventPublisher(client *kafka.Client) *ProductEventPublisher {
	return &ProductEventPublisher{client: client}
}

type ProductCreatedEvent struct {
	ProductID string `json:"product_id"`
	Name      string `json:"name"`
	Price     int64  `json:"price"`
	Stock     int64  `json:"stock"`
}

type StockUpdatedEvent struct {
	ProductID string `json:"product_id"`
	NewStock  int64  `json:"new_stock"`
	Version   int32  `json:"version"`
}

func (p *ProductEventPublisher) PublishProductCreated(ctx context.Context, product *domain.Product) error {
	event := ProductCreatedEvent{
		ProductID: product.ID.String(),
		Name:      product.Name,
		Price:     product.Price,
		Stock:     product.StockQuantity,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal product created event: %w", err)
	}

	return p.client.Publish(ctx, []byte(product.ID.String()), data)
}

func (p *ProductEventPublisher) PublishStockUpdated(ctx context.Context, product *domain.Product) error {
	event := StockUpdatedEvent{
		ProductID: product.ID.String(),
		NewStock:  product.StockQuantity,
		Version:   product.Version,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal stock updated event: %w", err)
	}

	return p.client.Publish(ctx, []byte(product.ID.String()), data)
}

// Ensure implementation
var _ application.ProductEventPublisher = (*ProductEventPublisher)(nil)
