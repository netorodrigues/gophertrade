# Quickstart: Running and Testing the Services

## Running the Services
To start the services locally:

1. Ensure infrastructure is running:
   ```bash
   docker-compose up -d
   ```
2. Start the inventory service:
   ```bash
   cd inventory
   go run cmd/server/main.go
   ```
3. Start the order service:
   ```bash
   cd order
   go run cmd/server/main.go
   ```

## Running the Tests
This project relies on `testcontainers-go` for integration testing. Docker must be running on your machine.

To run all tests including integration and Kafka/CQRS tests:
```bash
cd order && go test -v ./...
cd ../inventory && go test -v ./...
```