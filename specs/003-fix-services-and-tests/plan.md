# Implementation Plan: Fix Services and Tests

**Branch**: `003-fix-services-and-tests` | **Date**: 2026-03-26 | **Spec**: [spec.md](./spec.md)
**Input**: Fix services not being runnable (missing main.go), missing API tests, untested inter-service communication, untested event-driven architecture, and misleading documentation.

## Summary

This plan addresses critical gaps in the `order` and `inventory` services: adding missing application entry points (`cmd/server/main.go`), implementing comprehensive API and integration tests for REST, gRPC, and Kafka, and establishing a robust Testcontainers-based testing environment to validate the CQRS and Event-Driven Architecture.

## Technical Context

**Language/Version**: Go (Latest Stable)
**Primary Dependencies**: gRPC, pgx, Firestore SDK, ElasticSearch client, Kafka client, OpenTelemetry SDK, testcontainers-go
**Storage**: PostgreSQL, Firestore, ElasticSearch, Kafka
**Testing**: Go `testing` package, `net/http/httptest`, `testcontainers-go`
**Target Platform**: Linux server / Docker
**Project Type**: Microservices
**Performance Goals**: N/A (Testing/Bugfix focus)
**Constraints**: Tests must run reliably in CI without local dependencies installed (use Testcontainers)
**Scale/Scope**: 2 Services (order, inventory)

## Constitution Check

* I. Clean Code: Passes. Changes will follow Go idioms.
* II. Domain Driven Design (DDD): Passes. Tests will validate domain rules and bounded contexts.
* III. Mandatory Unit & Integration Testing (NON-NEGOTIABLE): Passes. This plan directly addresses major violations of this principle by implementing missing integration and E2E tests.
* IV. Observability: Passes. Will ensure test setups initialize telemetry.

## Project Structure

### Documentation (this feature)

```text
specs/003-fix-services-and-tests/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
└── tasks.md
```

### Source Code

```text
inventory/
├── cmd/
│   └── server/
│       └── main.go       # NEW: Entry point
├── tests/
│   └── integration/
│       ├── grpc_test.go  # NEW: gRPC API tests
│       └── kafka_test.go # NEW: Kafka integration tests

order/
├── cmd/
│   └── server/
│       └── main.go       # NEW: Entry point
├── tests/
│   ├── integration/
│   │   ├── api_test.go   # NEW: REST & gRPC API tests
│   │   └── kafka_test.go # NEW: Kafka integration tests
│   └── e2e/
│       └── journey_test.go # NEW: Order creation journey E2E test
```

## Complexity Tracking

No constitution violations needing justification.