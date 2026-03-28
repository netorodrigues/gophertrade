package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tckafka "github.com/testcontainers/testcontainers-go/modules/kafka"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func getFreePort() (string, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return "", err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return "", err
	}
	defer l.Close()
	return fmt.Sprintf("%d", l.Addr().(*net.TCPAddr).Port), nil
}

func TestE2EOrderJourney(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Find free ports
	invHTTPPort, _ := getFreePort()
	invGRPCPort, _ := getFreePort()
	orderHTTPPort, _ := getFreePort()
	orderGRPCPort, _ := getFreePort()

	// 1. Setup Infrastructure
	kafkaContainer, err := tckafka.Run(ctx, "confluentinc/cp-kafka:7.2.1")
	require.NoError(t, err)
	defer kafkaContainer.Terminate(ctx)
	brokers, _ := kafkaContainer.Brokers(ctx)
	broker := brokers[0]

	pgContainer, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("gophertrade"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
	)
	require.NoError(t, err)
	defer pgContainer.Terminate(ctx)
	pgHost, _ := pgContainer.Host(ctx)
	pgPort, _ := pgContainer.MappedPort(ctx, "5432")
	dbURL := fmt.Sprintf("postgres://postgres:postgres@%s:%s/gophertrade?sslmode=disable", pgHost, pgPort.Port())

	// Run migrations
	runMigrations(t, dbURL, "../../../inventory/internal/infrastructure/persistence/postgres/migrations/")
	runMigrations(t, dbURL, "../../../order/internal/infrastructure/persistence/postgres/migrations/")

	// 2. Start Inventory Service
	invCmd := exec.CommandContext(ctx, "go", "run", "cmd/server/main.go")
	invCmd.Dir = "../../../inventory"
	invCmd.Env = append(os.Environ(),
		"DATABASE_URL="+dbURL,
		"KAFKA_BROKERS="+broker,
		"HTTP_PORT="+invHTTPPort,
		"GRPC_PORT="+invGRPCPort,
	)
	var invOut bytes.Buffer
	invCmd.Stdout = &invOut
	invCmd.Stderr = &invOut
	err = invCmd.Start()
	require.NoError(t, err)
	defer invCmd.Process.Kill()

	// 3. Start Order Service
	orderCmd := exec.CommandContext(ctx, "go", "run", "cmd/server/main.go")
	orderCmd.Dir = "../../../order"
	orderCmd.Env = append(os.Environ(),
		"DATABASE_URL="+dbURL,
		"KAFKA_BROKERS="+broker,
		"HTTP_PORT="+orderHTTPPort,
		"GRPC_PORT="+orderGRPCPort,
		"INVENTORY_GRPC_ADDR=localhost:"+invGRPCPort,
	)
	var orderOut bytes.Buffer
	orderCmd.Stdout = &orderOut
	orderCmd.Stderr = &orderOut
	err = orderCmd.Start()
	require.NoError(t, err)
	defer orderCmd.Process.Kill()

	// Wait for services to be ready
	waitForURL(t, "http://localhost:"+invHTTPPort+"/api/v1/products/health", 30*time.Second, &invOut)
	waitForURL(t, "http://localhost:"+orderHTTPPort+"/api/v1/orders/health", 30*time.Second, &orderOut)

	// 4. THE JOURNEY
	// Step A: Create a product in Inventory
	productID := ""
	t.Run("Create Product in Inventory", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name":          "E2E Gopher",
			"price_cents":   5000,
			"initial_stock": 10,
		}
		data, _ := json.Marshal(reqBody)
		resp, err := http.Post("http://localhost:"+invHTTPPort+"/api/v1/products", "application/json", bytes.NewBuffer(data))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var prod map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&prod)
		productID = prod["ID"].(string)
	})

	// Step B: Create an order in Order Service (which calls Inventory via gRPC)
	t.Run("Create Order in Order Service", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"items": []map[string]interface{}{
				{"product_id": productID, "quantity": 2},
			},
		}
		data, _ := json.Marshal(reqBody)
		resp, err := http.Post("http://localhost:"+orderHTTPPort+"/api/v1/orders", "application/json", bytes.NewBuffer(data))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var order map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&order)
		assert.Equal(t, float64(10000), order["TotalPrice"])
	})

	// Step C: Verify Stock Decrement in Inventory
	t.Run("Verify Stock Decrement", func(t *testing.T) {
		resp, err := http.Get("http://localhost:"+invHTTPPort+"/api/v1/products/" + productID)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var prod map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&prod)
		assert.Equal(t, float64(8), prod["StockQuantity"])
	})
}

func waitForURL(t *testing.T, url string, timeout time.Duration, output *bytes.Buffer) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			return
		}
		time.Sleep(1 * time.Second)
	}
	t.Fatalf("timed out waiting for %s. Service Output:\n%s", url, output.String())
}

func runMigrations(t *testing.T, dbURL string, migrationsDir string) {
	ctx := context.Background()
	var conn *pgx.Conn
	var err error

	// Retry connection
	for i := 0; i < 10; i++ {
		conn, err = pgx.Connect(ctx, dbURL)
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	require.NoError(t, err, "failed to connect to db after retries")
	defer conn.Close(ctx)

	absDir, _ := filepath.Abs(migrationsDir)
	files, err := filepath.Glob(filepath.Join(absDir, "*.up.sql"))
	require.NoError(t, err)

	for _, f := range files {
		content, err := os.ReadFile(f)
		require.NoError(t, err)

		_, err = conn.Exec(ctx, string(content))
		require.NoError(t, err)
	}
}
