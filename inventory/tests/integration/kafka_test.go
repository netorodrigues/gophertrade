package integration

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"gophertrade/inventory/internal/domain"
	"gophertrade/inventory/internal/infrastructure/event"
	"gophertrade/inventory/internal/infrastructure/event/kafka"

	segmentiokafka "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tckafka "github.com/testcontainers/testcontainers-go/modules/kafka"
)

func setupKafka(t *testing.T) (string, func()) {
	ctx := context.Background()

	kafkaContainer, err := tckafka.Run(ctx, "confluentinc/cp-kafka:7.2.1")
	require.NoError(t, err)

	brokers, err := kafkaContainer.Brokers(ctx)
	require.NoError(t, err)

	return brokers[0], func() {
		kafkaContainer.Terminate(ctx)
	}
}

func TestKafkaIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	broker, cleanup := setupKafka(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	t.Run("Publish Product Created", func(t *testing.T) {
		topic := "inventory.products.created"
		createTopic(ctx, t, broker, topic)

		writer := kafka.NewWriter(broker, topic)
		client := &kafka.Client{Writer: writer}
		publisher := event.NewProductEventPublisher(client)

		product := domain.NewProduct("Test Product", 1000, 50)
		err := publisher.PublishProductCreated(ctx, product)
		require.NoError(t, err)

		// Read back using a raw kafka-go reader
		reader := kafka.NewReader(broker, topic, "test-group-created")
		defer reader.Close()

		msg, err := reader.ReadMessage(ctx)
		require.NoError(t, err)

		var received event.ProductCreatedEvent
		err = json.Unmarshal(msg.Value, &received)
		require.NoError(t, err)

		assert.Equal(t, product.ID.String(), received.ProductID)
		assert.Equal(t, product.Name, received.Name)
		assert.Equal(t, product.Price, received.Price)
		assert.Equal(t, product.StockQuantity, received.Stock)
	})

	t.Run("Publish Stock Updated", func(t *testing.T) {
		topic := "inventory.products.updated"
		createTopic(ctx, t, broker, topic)

		writer := kafka.NewWriter(broker, topic)
		client := &kafka.Client{Writer: writer}
		publisher := event.NewProductEventPublisher(client)

		product := domain.NewProduct("Stock Product", 2000, 10)
		product.UpdateStock(-5)
		product.Version = 2

		err := publisher.PublishStockUpdated(ctx, product)
		require.NoError(t, err)

		// Read back
		reader := kafka.NewReader(broker, topic, "test-group-updated")
		defer reader.Close()

		msg, err := reader.ReadMessage(ctx)
		require.NoError(t, err)

		var received event.StockUpdatedEvent
		err = json.Unmarshal(msg.Value, &received)
		require.NoError(t, err)

		assert.Equal(t, product.ID.String(), received.ProductID)
		assert.Equal(t, int64(5), received.NewStock)
		assert.Equal(t, int32(2), received.Version)
	})
}

func createTopic(ctx context.Context, t *testing.T, broker string, topic string) {
	conn, err := segmentiokafka.DialContext(ctx, "tcp", broker)
	require.NoError(t, err)
	defer conn.Close()

	err = conn.CreateTopics(segmentiokafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	})
	require.NoError(t, err)

	// Wait for topic to be visible
	for i := 0; i < 20; i++ {
		partitions, err := conn.ReadPartitions(topic)
		if err == nil && len(partitions) > 0 {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("topic %s not ready after 10 seconds", topic)
}
