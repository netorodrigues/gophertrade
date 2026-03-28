package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"gophertrade/internal/shared/telemetry"
	api_grpc "gophertrade/inventory/internal/api/grpc"
	api_http "gophertrade/inventory/internal/api/http"
	"gophertrade/inventory/internal/application"
	"gophertrade/inventory/internal/config"
	"gophertrade/inventory/internal/infrastructure/event"
	"gophertrade/inventory/internal/infrastructure/event/kafka"
	"gophertrade/inventory/internal/infrastructure/persistence/postgres"
	inventoryv1 "gophertrade/proto/inventory/v1"
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
	shutdownTracer, err := telemetry.InitTracer(ctx, "inventory-service")
	if err != nil {
		log.Printf("failed to init telemetry: %v", err)
	} else {
		defer shutdownTracer(ctx)
	}

	// 3. PostgreSQL Database
	pool, err := pgxpool.New(ctx, cfg.DB.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to db: %w", err)
	}
	defer pool.Close()
	repo := postgres.NewProductRepository(pool)

	// 4. Kafka
	kafkaWriter := kafka.NewWriter(cfg.Kafka.Brokers, "products")
	kafkaClient := &kafka.Client{
		Writer: kafkaWriter,
	}
	var publisher application.ProductEventPublisher = event.NewProductEventPublisher(kafkaClient)
	defer kafkaClient.Close()

	// 5. Service
	service := application.NewProductService(repo, publisher)

	// 6. HTTP Handler
	httpHandler := api_http.NewProductHandler(service)
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Mount("/api/v1/products", httpHandler.Routes())

	// 7. GRPC Handler
	grpcHandler := api_grpc.NewProductHandler(service)
	grpcServer := grpc.NewServer()
	inventoryv1.RegisterInventoryServiceServer(grpcServer, grpcHandler)
	reflection.Register(grpcServer)

	// 8. Start Servers
	errChan := make(chan error, 2)

	go func() {
		log.Printf("Starting HTTP server on %s", cfg.Web.APIHost)
		if err := http.ListenAndServe(cfg.Web.APIHost, r); err != nil {
			errChan <- fmt.Errorf("http server failed: %w", err)
		}
	}()

	go func() {
		lis, err := net.Listen("tcp", ":"+cfg.GRPC.Port)
		if err != nil {
			errChan <- fmt.Errorf("failed to listen on grpc port: %w", err)
			return
		}
		log.Printf("Starting gRPC server on :%s", cfg.GRPC.Port)
		if err := grpcServer.Serve(lis); err != nil {
			errChan <- fmt.Errorf("grpc server failed: %w", err)
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		log.Println("Shutting down...")
		grpcServer.GracefulStop()
		return nil
	}
}
