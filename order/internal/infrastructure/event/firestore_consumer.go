package event

import (
	"context"
	"encoding/json"
	"log"

	"gophertrade/order/internal/infrastructure/event/kafka"
	"gophertrade/order/internal/infrastructure/persistence/firestore"
)

type FirestoreSyncConsumer struct {
	client *kafka.Client
	repo   *firestore.OrderReadRepository
}

func NewFirestoreSyncConsumer(client *kafka.Client, repo *firestore.OrderReadRepository) *FirestoreSyncConsumer {
	return &FirestoreSyncConsumer{
		client: client,
		repo:   repo,
	}
}

func (c *FirestoreSyncConsumer) Start(ctx context.Context) {
	for {
		msg, err := c.client.Reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("failed to read message from kafka: %v", err)
			continue
		}

		var event OrderCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("failed to unmarshal order created event: %v", err)
			continue
		}

		// Convert to Firestore View
		var items []firestore.OrderItem
		for _, item := range event.Items {
			items = append(items, firestore.OrderItem{
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
				UnitPrice: item.UnitPrice,
			})
		}

		view := &firestore.OrderView{
			OrderID:    event.OrderID,
			TotalPrice: event.TotalPrice,
			Status:     event.Status,
			CreatedAt:  event.CreatedAt.String(),
			Items:      items,
		}

		if err := c.repo.Save(ctx, view); err != nil {
			log.Printf("failed to sync order to firestore: %v", err)
		}
	}
}
