# Quickstart: Backend Order and Stock Management

## Prerequisites

- Go 1.22+
- Docker & Docker Compose
- `protoc` compiler
- `grpcurl` (for testing gRPC)

## Running the Stack

1. **Start Infrastructure**:
   ```bash
   docker-compose up -d postgres kafka firestore elasticsearch jaeger
   ```

2. **Run Migrations**:
   ```bash
   go run cmd/migrate/main.go up
   ```

3. **Start Services**:
   ```bash
   # Terminal 1
   go run order/cmd/server/main.go

   # Terminal 2
   go run inventory/cmd/server/main.go
   ```

## Testing Manually

### 1. Create a Product (Inventory Service)
```bash
grpcurl -plaintext -d '{"name": "Gopher Plushie", "price_cents": 2500, "initial_stock": 100}' localhost:50052 inventory.v1.InventoryService/CreateProduct
# Returns: {"product_id": "uuid-..."}
```

### 2. Create an Order (Order Service)
```bash
grpcurl -plaintext -d '{"items": [{"product_id": "uuid-...", "quantity": 2}]}' localhost:50051 order.v1.OrderService/CreateOrder
```

### 3. Verify Stock Update
```bash
grpcurl -plaintext -d '{"product_id": "uuid-..."}' localhost:50052 inventory.v1.InventoryService/GetProduct
# Should show stock_quantity: 98
```

## Running Tests

```bash
# Unit Tests
go test ./...

# Integration Tests (requires Docker)
go test -tags=integration ./...
```
