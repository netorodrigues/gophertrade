# Tasks: Order and Stock Management

**Input**: Design documents from `/specs/002-order-stock-management/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Mandatory per Constitution Principle III. Every user story includes corresponding unit and integration tests.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 Initialize `inventory` service as a Go module in `inventory/go.mod`
- [x] T002 Initialize `order` service as a Go module in `order/go.mod`
- [x] T003 [P] Setup linting (golangci-lint) and formatting in root `.golangci.yml`
- [x] T004 [P] Configure Docker Compose for local infrastructure (Postgres, Kafka, Firestore Emulator, ElasticSearch) in `docker-compose.yml`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

- [x] T005 [P] Setup Postgres migrations for `inventory` in `inventory/internal/infrastructure/persistence/postgres/migrations/` (Use `BIGINT` for cents)
- [x] T006 [P] Setup Postgres migrations for `order` in `order/internal/infrastructure/persistence/postgres/migrations/` (Use `BIGINT` for cents)
- [x] T007 [P] Generate gRPC code from proto files using Buf in `proto/inventory/v1/` and `proto/order/v1/`
- [x] T008 [P] Implement Mock Auth HTTP middleware in `internal/shared/middleware/auth.go` (Header-based)
- [x] T009 [P] Implement Mock Auth gRPC interceptors in `internal/shared/api/grpc/interceptors/auth.go`
- [x] T010 Configure Kafka producer and consumer base utilities in `inventory/internal/infrastructure/event/kafka/client.go`

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Manage Products and Stock (Priority: P1) 🎯 MVP

**Goal**: Create products and adjust their stock levels (Inventory service)

**Independent Test**: Can create a product via REST, verify via gRPC, and increment stock with optimistic locking verification.

### Tests for User Story 1 (MANDATORY) ⚠️

- [X] T011 [P] [US1] Unit test for product entity in `inventory/internal/domain/product_test.go`
- [X] T012 [P] [US1] Integration test for product repository in `inventory/tests/integration/product_repo_test.go`
- [X] T013 [P] [US1] Integration test for product REST API (Chi) in `inventory/tests/integration/product_http_test.go`

### Implementation for User Story 1

- [X] T014 [P] [US1] Create Product Entity with optimistic locking in `inventory/internal/domain/product.go`
- [X] T015 [US1] Implement Product Repository (Postgres/pgx) in `inventory/internal/infrastructure/persistence/postgres/product_repository.go`
- [X] T016 [US1] Implement Product Service with stock rules in `inventory/internal/application/product_service.go`
- [X] T017 [US1] Implement Product gRPC Handler (including `BatchUpdateStock`) in `inventory/internal/api/grpc/product_handler.go`
- [X] T018 [US1] Implement Product REST Handler (Chi) in `inventory/internal/api/http/product_handler.go`
- [X] T019 [US1] Integrate Kafka Producer to emit `ProductCreated` and `StockUpdated` events in `inventory/internal/infrastructure/event/product_events.go`

**Checkpoint**: User Story 1 (Inventory Management) fully functional and testable independently

---

## Phase 4: User Story 2 - Create Valid Orders (Priority: P1)

**Goal**: Create orders with multiple items while enforcing stock constraints (Order service)

**Independent Test**: Create an order via REST, which triggers a `BatchUpdateStock` gRPC call to Inventory. Verify success/failure scenarios.

### Tests for User Story 2 (MANDATORY) ⚠️

- [x] T020 [P] [US2] Unit test for order calculation and snapshots in `order/internal/domain/order_test.go`
- [x] T021 [P] [US2] Integration test for order creation journey (inter-service) in `order/tests/integration/order_journey_test.go`

### Implementation for User Story 2

- [x] T022 [P] [US2] Create Order and OrderItem Entities (price snapshots) in `order/internal/domain/order.go`
- [x] T023 [US2] Implement Order Repository (Postgres/pgx) in `order/internal/infrastructure/persistence/postgres/order_repository.go`
- [x] T024 [US2] Implement gRPC Client for Inventory Service in `order/internal/infrastructure/grpc/inventory_client.go`
- [x] T025 [US2] Implement Order Service (Coordination with `BatchUpdateStock`) in `order/internal/application/order_service.go`
- [x] T026 [US2] Implement Order REST Handler (Chi) in `order/internal/api/http/order_handler.go`
- [x] T027 [US2] Implement Order gRPC Handler in `order/internal/api/grpc/order_handler.go`
- [x] T028 [US2] Integrate Kafka Producer to emit `OrderCreated` events in `order/internal/infrastructure/event/order_events.go`

**Checkpoint**: User Story 2 fully functional - inter-service communication verified

---

## Phase 5: User Story 3 - View Orders (Priority: P2)

**Goal**: View order details and history using read-optimized stores (CQRS)

**Independent Test**: Create orders, wait for eventual consistency, then search via REST and verify Firestore/ElasticSearch results.

### Tests for User Story 3 (MANDATORY) ⚠️

- [x] T029 [P] [US3] Integration test for read model sync (Kafka -> Firestore/ES) in `order/tests/integration/read_model_test.go`
- [x] T030 [P] [US3] Contract test for query REST API in `order/tests/integration/query_api_test.go`

### Implementation for User Story 3

- [x] T031 [P] [US3] Implement Kafka Consumer for Firestore sync in `order/internal/infrastructure/event/firestore_consumer.go`
- [x] T032 [P] [US3] Implement Kafka Consumer for ElasticSearch sync in `order/internal/infrastructure/event/es_consumer.go`
- [x] T033 [US3] Implement Firestore Read Repository in `order/internal/infrastructure/persistence/firestore/order_read_repository.go`
- [x] T034 [US3] Implement ElasticSearch Search Repository in `order/internal/infrastructure/persistence/es/order_search_repository.go`
- [x] T035 [US3] Implement Query REST Handler (Chi) in `order/internal/api/http/query_handler.go`

**Checkpoint**: All user stories functional - CQRS architecture complete

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [x] T036 [P] [US1, US2] Integrate OpenTelemetry (Tracing/Metrics) across both services in `internal/shared/telemetry/otel.go`
- [x] T037 Performance audit: Validate P95 < 200ms for order creation in `order/tests/performance/`
- [x] T038 Add health check endpoints (/health) to both services in `api/http/health.go`
- [x] T039 [P] Final documentation updates in `quickstart.md`
- [x] T040 Final validation against all requirements in `spec.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Can start immediately.
- **Foundational (Phase 2)**: Depends on Phase 1 completion - BLOCKS all user stories.
- **User Stories (Phase 3+)**: All depend on Phase 2.
  - US1 can be completed before US2.
  - US2 depends on US1's gRPC service being defined (`BatchUpdateStock`).
  - US3 depends on US1 and US2's Kafka producers being active.
- **Polish (Final Phase)**: Depends on all user stories being functionally complete.

### Parallel Execution Examples

```bash
# Foundational (T005, T006, T007)
Task T005: Setup Inventory Postgres migrations
Task T006: Setup Order Postgres migrations
Task T007: Generate gRPC code from proto

# Implementation for US1
Task T014: Product Entity (domain)
Task T011: Unit tests (P)
Task T012: Repo integration tests (P)
```

---

## Implementation Strategy

### MVP First (User Story 1 & 2 Only)

1. Complete Setup and Foundational.
2. Complete US1 (Inventory Management) - Verify via REST/gRPC.
3. Complete US2 (Order Creation) - Verify inter-service `BatchUpdateStock` call to US1.
4. **MVP**: Orders can be created and stock is managed correctly.

### Incremental Delivery

1. Foundation ready.
2. US1 Ready (Inventory service).
3. US2 Ready (Order service + gRPC integration).
4. US3 Ready (Read side synchronization).
5. Polish (Telemetry + Performance targets).
