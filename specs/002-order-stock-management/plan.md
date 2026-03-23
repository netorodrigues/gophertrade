# Implementation Plan: Backend Order and Stock Management

**Branch**: `002-order-stock-management` | **Date**: 2026-03-22 | **Spec**: [Link to Spec](spec.md)
**Input**: Feature specification from `/specs/002-order-stock-management/spec.md`

## Summary

Build a backend system comprising two microservices: `order` (order management) and `inventory` (product/stock management). They will expose REST/HTTP APIs using **Chi** for external interaction and communicate via **gRPC** for inter-service calls. The architecture follows DDD with vertical slices and CQRS. PostgreSQL is the write DB, while Firestore (ID-based) and ElasticSearch (search-based) are read DBs, synchronized via Kafka event bus. Observability will be handled by OpenTelemetry/Jaeger.

## Technical Context

**Language/Version**: Go (Latest Stable)
**Primary Dependencies**: 
- **REST**: [Chi](https://github.com/go-chi/chi) (standardized for idiomatic Go routing)
- **gRPC**: Standard Go gRPC with [Buf](https://buf.build/) for tooling
- **Database**: `pgx/v5` (Postgres), Firestore SDK, ElasticSearch v8 client
- **Messaging**: `kafka-go` (Segment)
- **Observability**: OpenTelemetry SDK + Jaeger exporter
- **Authentication**: **Mock/Header-based** middleware (reads `User-ID` from header)
**Storage**: PostgreSQL (Write - Strong Consistency), Firestore (Read - Key/Value), ElasticSearch (Read - Search).
**Communication**: External: REST/HTTP (Chi) | Internal: gRPC.
**Testing**: Go `testing`, `testcontainers-go` for integration (Postgres/Kafka).
**Target Platform**: Linux containers (Docker/Kubernetes).
**Project Type**: Microservices (Backend).
**Performance Goals**: 
- **Latency**: < 200ms for order creation (P95).
- **Throughput**: Support ~500 concurrent order creation requests/sec.
- **Consistency**: Strong consistency for inventory via PostgreSQL + Optimistic Locking.
**Constraints**: 
- All monetary values stored as **Integers (Cents)** in `BIGINT` or `INTEGER` columns.
- Atomic stock decrement via single gRPC batch call.

## Constitution Check

- **I. Clean Code**: Using Chi (lightweight) and Repository pattern to keep domain logic isolated.
- **II. Domain Driven Design**: Split into `inventory` and `order` bounded contexts.
- **III. Mandatory Testing**: Integrated into all implementation phases via `testcontainers-go`.
- **IV. Observability**: OTEL integration is part of the core infrastructure.

## Project Structure

### Documentation

```text
specs/002-order-stock-management/
├── plan.md              # This file
├── research.md          # Technology decisions and Rationale
├── data-model.md        # Database schemas and entities
├── quickstart.md        # Setup guide
├── contracts/           # Protobuf and API definitions
└── tasks.md             # Implementation tasks
```

### Source Code

```text
order/
├── cmd/server/          # App entry point
├── internal/
│   ├── domain/          # Order and OrderItem entities
│   ├── application/     # Order creation and listing logic
│   ├── infrastructure/  # Postgres, gRPC Client, Kafka Publisher
│   └── api/             # Chi REST handlers and gRPC handlers
├── tests/               # Integration tests (testcontainers)
└── go.mod

inventory/
├── cmd/server/
├── internal/
│   ├── domain/          # Product entity + Optimistic Locking
│   ├── application/     # Stock management logic
│   ├── infrastructure/  # Postgres persistence, Kafka Publisher
│   └── api/             # gRPC handlers and Chi REST handlers
├── tests/
└── go.mod

internal/shared/         # Shared utilities (Mock Auth, OTEL, Middleware)

proto/                   # Protobuf definitions (v1)
docker-compose.yml       # Infrastructure (Postgres, Kafka, ES, Firestore Emulator)
```

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| CQRS + Dual Read DBs | User requirement for specific search/read patterns | Simple CRUD insufficient for scale/search requirements |
| Microservices | Separation of Order and Inventory domains | Monolith less scalable for distinct load profiles |
