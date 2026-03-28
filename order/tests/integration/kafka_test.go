package integration

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"gophertrade/order/internal/domain"
	"gophertrade/order/internal/infrastructure/event"
	"gophertrade/order/internal/infrastructure/event/kafka"

	"github.com/google/uuid"
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

func TestOrderKafkaIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	broker, cleanup := setupKafka(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	t.Run("Publish Order Created", func(t *testing.T) {
		topic := "orders.created"
		createTopic(ctx, t, broker, topic)

		writer := kafka.NewWriter(broker, topic)
		client := &kafka.Client{Writer: writer}
		publisher := event.NewOrderEventPublisher(client)

		items := []domain.OrderItem{
			{ProductID: uuid.New(), Quantity: 2, UnitPrice: 1000},
			{ProductID: uuid.New(), Quantity: 1, UnitPrice: 500},
		}
		order, err := domain.NewOrder(items)
		require.NoError(t, err)

		err = publisher.PublishOrderCreated(ctx, order)
		require.NoError(t, err)

		// Read back
		reader := kafka.NewReader(broker, topic, "test-order-group")
		defer reader.Close()

		msg, err := reader.ReadMessage(ctx)
		require.NoError(t, err)

		var received event.OrderCreatedEvent
		err = json.Unmarshal(msg.Value, &received)
		require.NoError(t, err)

		assert.Equal(t, order.ID.String(), received.OrderID)
		assert.Equal(t, int64(2500), received.TotalPrice)
		assert.Equal(t, 2, len(received.Items))
	})
}
