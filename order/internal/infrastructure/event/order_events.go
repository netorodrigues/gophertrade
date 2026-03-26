package event

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gophertrade/order/internal/domain"
	"gophertrade/order/internal/infrastructure/event/kafka"
)

type OrderEventPublisher struct {
	client *kafka.Client
}

func NewOrderEventPublisher(client *kafka.Client) *OrderEventPublisher {
	return &OrderEventPublisher{client: client}
}

type OrderCreatedEvent struct {
	OrderID    string      `json:"order_id"`
	TotalPrice int64       `json:"total_price"`
	Status     string      `json:"status"`
	CreatedAt  time.Time   `json:"created_at"`
	Items      []OrderItem `json:"items"`
}

type OrderItem struct {
	ProductID string `json:"product_id"`
	Quantity  int32  `json:"quantity"`
	UnitPrice int64  `json:"unit_price"`
}

func (p *OrderEventPublisher) PublishOrderCreated(ctx context.Context, order *domain.Order) error {
	var items []OrderItem
	for _, item := range order.Items {
		items = append(items, OrderItem{
			ProductID: item.ProductID.String(),
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		})
	}

	event := OrderCreatedEvent{
		OrderID:    order.ID.String(),
		TotalPrice: order.TotalPrice,
		Status:     string(order.Status),
		CreatedAt:  order.CreatedAt,
		Items:      items,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal order created event: %w", err)
	}

	return p.client.Publish(ctx, []byte(order.ID.String()), data)
}
