package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"cloud.google.com/go/firestore"
	"github.com/elastic/go-elasticsearch/v8"

	"gophertrade/internal/shared/telemetry"
	api_http "gophertrade/order/internal/api/http"
	"gophertrade/order/internal/application"
	"gophertrade/order/internal/config"
	"gophertrade/order/internal/infrastructure/event"
	"gophertrade/order/internal/infrastructure/event/kafka"
	inf_grpc "gophertrade/order/internal/infrastructure/grpc"
	"gophertrade/order/internal/infrastructure/persistence/es"
	fsrepo "gophertrade/order/internal/infrastructure/persistence/firestore"
	"gophertrade/order/internal/infrastructure/persistence/postgres"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		log.Fatalf("service failed: %v", err)
	}
}

func run(ctx context.Context) error {
	// 1. Config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// 2. Telemetry
	shutdownTracer, err := telemetry.InitTracer(ctx, "order-service")
	if err != nil {
		log.Printf("failed to init telemetry: %v", err)
	} else {
		defer shutdownTracer(ctx)
	}

	// 3. PostgreSQL Database
	pool, err := pgxpool.New(ctx, cfg.DB.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer pool.Close()
	repo := postgres.NewOrderRepository(pool)

	// 4. CQRS Read Components (Firestore & Elasticsearch)
	os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8080")
	fsClient, err := firestore.NewClient(ctx, "gophertrade")
	if err != nil {
		return fmt.Errorf("failed to init firestore client: %w", err)
	}
	defer fsClient.Close()
	fsRepo := fsrepo.NewOrderReadRepository(fsClient)

	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
	    Addresses: []string{"http://localhost:9200"},
	})
	if err != nil {
		return fmt.Errorf("failed to init elasticsearch client: %w", err)
	}
	esRepo := es.NewOrderSearchRepository(esClient)

	// 5. Kafka
	kafkaWriter := kafka.NewWriter(cfg.Kafka.Brokers, "orders")
	publishClient := &kafka.Client{Writer: kafkaWriter}
	defer publishClient.Close()

	readerFS := kafka.NewReader(cfg.Kafka.Brokers, "orders", "order-sync-firestore")
	clientFS := &kafka.Client{Reader: readerFS}
	defer clientFS.Close()

	readerES := kafka.NewReader(cfg.Kafka.Brokers, "orders", "order-sync-elasticsearch")
	clientES := &kafka.Client{Reader: readerES}
	defer clientES.Close()

	publisher := event.NewOrderEventPublisher(publishClient)

	// 6. Start Consumers in Background
	fsConsumer := event.NewFirestoreSyncConsumer(clientFS, fsRepo)
	go fsConsumer.Start(ctx)

	esConsumer := event.NewESSyncConsumer(clientES, esRepo)
	go esConsumer.Start(ctx)

	// 7. Inventory GRPC Client
	inventoryClient, err := inf_grpc.NewInventoryClient(cfg.GRPC.InventoryAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to inventory service: %w", err)
	}

	// 8. Services & Handlers
	service := application.NewOrderService(repo, inventoryClient, publisher)
	httpHandler := api_http.NewOrderHandler(service)
	queryHandler := api_http.NewQueryHandler(fsRepo, esRepo)

	// 9. Router Setup (Transparent CQRS mapping)
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/v1/orders", func(r chi.Router) {
		r.Get("/health", api_http.HealthCheck)
		r.Post("/", httpHandler.CreateOrder)       // Write -> Postgres
		r.Get("/{id}", queryHandler.ViewOrder)     // Read by ID -> Firestore
		r.Get("/", queryHandler.SearchOrders)      // Read collection/search -> Elasticsearch
	})

	// 10. Start Server
	errChan := make(chan error, 1)

	go func() {
		log.Printf("Starting HTTP server on %s", cfg.Web.APIHost)
		if err := http.ListenAndServe(cfg.Web.APIHost, r); err != nil {
			errChan <- fmt.Errorf("http server failed: %w", err)
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		log.Println("Shutting down...")
		return nil
	}
}
