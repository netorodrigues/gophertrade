package event

import (
	"context"
	"encoding/json"
	"log"

	"gophertrade/order/internal/infrastructure/event/kafka"
	"gophertrade/order/internal/infrastructure/persistence/es"
)

type ESSyncConsumer struct {
	client *kafka.Client
	repo   *es.OrderSearchRepository
}

func NewESSyncConsumer(client *kafka.Client, repo *es.OrderSearchRepository) *ESSyncConsumer {
	return &ESSyncConsumer{
		client: client,
		repo:   repo,
	}
}

func (c *ESSyncConsumer) Start(ctx context.Context) {
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

		doc := &es.OrderDoc{
			OrderID:    event.OrderID,
			TotalPrice: event.TotalPrice,
			Status:     event.Status,
			CreatedAt:  event.CreatedAt.String(),
		}

		if err := c.repo.Save(ctx, doc); err != nil {
			log.Printf("failed to sync order to elasticsearch: %v", err)
		}
	}
}
