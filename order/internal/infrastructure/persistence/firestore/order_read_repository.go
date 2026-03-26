package firestore

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
)

type OrderReadRepository struct {
	client *firestore.Client
}

func NewOrderReadRepository(client *firestore.Client) *OrderReadRepository {
	return &OrderReadRepository{client: client}
}

type OrderView struct {
	OrderID    string      `firestore:"order_id"`
	TotalPrice int64       `firestore:"total_price"`
	Status     string      `firestore:"status"`
	CreatedAt  string      `firestore:"created_at"`
	Items      []OrderItem `firestore:"items"`
}

type OrderItem struct {
	ProductID string `firestore:"product_id"`
	Quantity  int32  `firestore:"quantity"`
	UnitPrice int64  `firestore:"unit_price"`
}

func (r *OrderReadRepository) Save(ctx context.Context, order *OrderView) error {
	_, err := r.client.Collection("orders_view").Doc(order.OrderID).Set(ctx, order)
	if err != nil {
		return fmt.Errorf("failed to save order to firestore: %w", err)
	}
	return nil
}

func (r *OrderReadRepository) GetByID(ctx context.Context, id string) (*OrderView, error) {
	doc, err := r.client.Collection("orders_view").Doc(id).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get order from firestore: %w", err)
	}

	var order OrderView
	if err := doc.DataTo(&order); err != nil {
		return nil, fmt.Errorf("failed to unmarshal firestore data: %w", err)
	}

	return &order, nil
}
