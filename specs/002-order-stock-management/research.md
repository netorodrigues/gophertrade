# Research: Backend Order and Stock Management

## Technology Decisions

### 1. Inter-service Communication
- **Decision**: gRPC
- **Rationale**: User Requirement. Provides strict contract (Protobuf), high performance, and strongly typed code generation.
- **Alternatives Considered**: REST (JSON) - Rejected due to user requirement and lower performance for internal communication.

### 2. Event Bus
- **Decision**: Kafka
- **Rationale**: User Requirement. Standard for high-throughput event streaming and decoupling write/read models in CQRS.
- **Library**: `github.com/segmentio/kafka-go` (Idiomatic Go) or `github.com/confluentinc/confluent-kafka-go` (CGO wrapper, high perf).
  - **Selection**: `github.com/segmentio/kafka-go` for pure Go implementation and ease of build (no C dependencies).

### 3. Database Drivers
- **PostgreSQL**: `github.com/jackc/pgx/v5` - High performance, standard for Go.
- **Firestore**: `cloud.google.com/go/firestore` - Official SDK.
- **ElasticSearch**: `github.com/elastic/go-elasticsearch/v8` - Official Typed Client.

### 4. Observability
- **Decision**: OpenTelemetry (OTEL)
- **Rationale**: User Requirement. Industry standard, vendor-neutral.
- **Exporter**: OTLP exporter to send data to Jaeger.

## Pattern Decisions

### 1. CQRS Implementation
- **Write Side**:
  - Receives Command (gRPC/REST).
  - Validates Domain Rules.
  - Persists to PostgreSQL (Single Source of Truth).
  - Publishes Event to Kafka (Transactional Outbox Pattern recommended for consistency, but simple publish-after-commit acceptable for MVP if ack is handled).
- **Read Side**:
  - Kafka Consumer Group reads events.
  - Updates Firestore (Key-Value lookups).
  - Updates ElasticSearch (Full-text search).

### 2. Vertical Slices structure in Go
- Instead of "Layered" (Controller -> Service -> Repo), organize by Feature/Slice within the module.
- Example: `internal/application/commands/create_order.go` contains the handler, validation, and domain logic for that specific action.
