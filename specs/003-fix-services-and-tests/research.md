# Phase 0: Research

## Testing gRPC and REST APIs in Go
- **Decision:** Use standard `testing` package with `net/http/httptest` for REST and `google.golang.org/grpc/test/bufconn` or real ports for gRPC integration tests.
- **Rationale:** Built-in tools and `bufconn` allow fast, in-memory networking for API tests without port conflicts, ensuring reliable CI runs.
- **Alternatives considered:** Spin up full servers on real ports for every test (slower, port conflicts).

## End-to-End Inter-Service Communication
- **Decision:** Implement a true E2E test suite in a separate `tests/e2e` directory or as part of integration tests using Docker Compose or Testcontainers.
- **Rationale:** Verifies the "order creation journey" across `order` and `inventory` services as mandated by the Constitution (Principle III: Integration Tests).
- **Alternatives considered:** Mocks only (violates the need to test the actual integration).

## Testing Event-Driven Architecture (Kafka) & CQRS (Firestore, ElasticSearch)
- **Decision:** Use `testcontainers-go` to spin up ephemeral Kafka, Postgres, Firestore (emulator), and ElasticSearch containers during `go test`.
- **Rationale:** Provides real dependencies for CQRS and event-driven flows, completely replacing the placeholder `0.00s` execution tests.
- **Alternatives considered:** Mocking Kafka and databases (leads to false confidence and untested integration code).

## Entry Points (Runnable Services)
- **Decision:** Create standard Go layout entry points: `order/cmd/server/main.go` and `inventory/cmd/server/main.go`.
- **Rationale:** Fulfills the "Services are Not Runnable" requirement, providing the necessary `main` packages to initialize dependencies and start HTTP/gRPC servers.
- **Alternatives considered:** Putting `main.go` in the root of the service directories.