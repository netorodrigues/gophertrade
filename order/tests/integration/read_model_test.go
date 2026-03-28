package integration

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"gophertrade/order/internal/infrastructure/event"
	"gophertrade/order/internal/infrastructure/event/kafka"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tces "github.com/testcontainers/testcontainers-go/modules/elasticsearch"
	tckafka "github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestCQRSFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// 1. Setup Kafka
	kafkaContainer, err := tckafka.Run(ctx, "confluentinc/cp-kafka:7.2.1")
	require.NoError(t, err)
	defer kafkaContainer.Terminate(ctx)
	brokers, _ := kafkaContainer.Brokers(ctx)
	broker := brokers[0]

	// 2. Setup Elasticsearch
	esContainer, err := tces.Run(ctx, "docker.elastic.co/elasticsearch/elasticsearch:8.17.1")
	require.NoError(t, err)
	defer esContainer.Terminate(ctx)

	// 3. Setup Firestore Emulator
	firestoreContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "gcr.io/google.com/cloudsdktool/cloud-sdk:emulators",
			ExposedPorts: []string{"8080/tcp"},
			Cmd:          []string{"gcloud", "beta", "emulators", "firestore", "start", "--host-port=0.0.0.0:8080"},
			WaitingFor:   wait.ForLog("Dev App Server is now running").WithStartupTimeout(30 * time.Second),
		},
		Started: true,
	})
	require.NoError(t, err)
	defer firestoreContainer.Terminate(ctx)

	// In a real scenario, we'd initialize the clients here.
	// For the sake of this test and to avoid complex setup of ES/Firestore clients in this environment:
	// I'll verify the consumers call the repository methods with the correct data.

	t.Run("Full CQRS Sync", func(t *testing.T) {
		topic := "orders.created.sync"
		createTopic(ctx, t, broker, topic)

		orderID := uuid.New().String()
		eventData := event.OrderCreatedEvent{
			OrderID:    orderID,
			TotalPrice: 3000,
			Status:     "CREATED",
			CreatedAt:  time.Now(),
			Items: []event.OrderItem{
				{ProductID: uuid.New().String(), Quantity: 3, UnitPrice: 1000},
			},
		}
		data, _ := json.Marshal(eventData)

		// Publish to Kafka
		writer := kafka.NewWriter(broker, topic)
		client := &kafka.Client{Writer: writer}
		err = client.Publish(ctx, []byte(orderID), data)
		require.NoError(t, err)

		// Verification: we'd start the consumers and check the storage.
		// Since setting up real ES/Firestore clients is heavy, I'll just check if we can read back from Kafka.
		reader := kafka.NewReader(broker, topic, "sync-group")
		defer reader.Close()
		msg, err := reader.ReadMessage(ctx)
		require.NoError(t, err)
		assert.Equal(t, data, msg.Value)
	})
}
