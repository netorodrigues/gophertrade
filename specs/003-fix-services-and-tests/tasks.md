# Tasks: Fix Services and Tests

**Input**: Design documents from `/specs/003-fix-services-and-tests/`
**Prerequisites**: plan.md, research.md, data-model.md, quickstart.md
**Context**: Implement test infrastructure, API tests, E2E tests, and make services runnable. Garanteed validation via test execution for every step.

## Phase 1: Setup & Foundational (Entry Points & Testcontainers)

**Purpose**: Provide the fundamental infrastructure to run the applications and spin up test containers for robust integration testing.

- [x] T001 [P] Create entry point for the inventory service in `inventory/cmd/server/main.go`
- [x] T002 [P] Create entry point for the order service in `order/cmd/server/main.go`
- [x] T003 Validate Phase 1: Run `go build ./cmd/server/main.go` in both `inventory` and `order` to ensure they compile and can be started.

---

## Phase 2: User Story 1 - Comprehensive API Testing (Priority: P1) 🎯 MVP

**Goal**: Implement comprehensive tests for the REST and gRPC APIs across both services.

**Independent Test**: The API test suites for both `order` and `inventory` execute successfully using standard `go test` and `bufconn`/`httptest`.

### Implementation for User Story 1

- [x] T004 [P] [US1] Implement gRPC API integration tests for the inventory service in `inventory/tests/integration/grpc_test.go`
- [x] T005 [P] [US1] Implement REST API integration tests for the order service in `order/tests/integration/api_test.go`
- [x] T006 [P] [US1] Implement gRPC API integration tests for the order service in `order/tests/integration/api_test.go`
- [x] T007 [US1] Validation Step: Execute `go test -v ./tests/integration/...` in both directories to validate the functioning of all API endpoints. Do not mark US1 as done until tests pass.

---

## Phase 3: User Story 2 - Event-Driven Architecture & CQRS Testing (Priority: P2)

**Goal**: Ensure Kafka integrations (producers and consumers) and CQRS flows (updating Firestore and Elasticsearch) are fully tested using real instances via Testcontainers.

**Independent Test**: Kafka messages can be produced and consumed, and data propagates to read models, tested via Testcontainers.

### Implementation for User Story 2

- [x] T008 [P] [US2] Implement Testcontainers setup and Kafka integration tests for the inventory service in `inventory/tests/integration/kafka_test.go`
- [x] T009 [P] [US2] Implement Testcontainers setup and Kafka integration tests for the order service in `order/tests/integration/kafka_test.go`
- [x] T010 [P] [US2] Update `order/tests/integration/read_model_test.go` (if existing) or add tests in `kafka_test.go` to validate CQRS flow using Testcontainers for Firestore and Elasticsearch.
- [x] T011 [US2] Validation Step: Execute `go test -v ./tests/integration/...` with Docker running to validate Kafka and CQRS. Do not mark US2 as done until Testcontainer-backed tests pass successfully.

---

## Phase 4: User Story 3 - End-to-End Inter-Service Journey (Priority: P3)

**Goal**: Validate the complete "order creation journey", verifying that the gRPC client in the order service successfully communicates with the inventory service.

**Independent Test**: An E2E test runs successfully, verifying inter-service communication and full state transitions.

### Implementation for User Story 3

- [x] T012 [US3] Implement the E2E order journey test spanning both services in `order/tests/e2e/journey_test.go`
- [x] T013 [US3] Validation Step: Execute `go test -v ./tests/e2e/journey_test.go` to validate the entire workflow end-to-end. Do not mark US3 as done until the E2E journey successfully passes.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Fix misleading documentation and perform final verifications.

- [x] T014 [P] Update `tasks.md` or any misleading documentation in the project to accurately reflect the actual implementation state.
- [x] T015 Final Validation: Run the full test suite across the entire project `go test -v ./...` in both `order` and `inventory` directories.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Must be completed first to allow services to run.
- **User Story 1 (Phase 2)**: Can be started independently for API boundaries.
- **User Story 2 (Phase 3)**: Independent of US1, focuses on event-driven infrastructure.
- **User Story 3 (Phase 4)**: Depends on the APIs and services running (Phase 1 & 2), and preferably Kafka (Phase 3) for the full flow.
- **Polish (Phase 5)**: Executed after all functionality is in place.

### Parallel Opportunities

- T001 and T002 (Foundational Entry Points) can be done in parallel.
- API Tests for Inventory (T004) and Order (T005, T006) can be built in parallel.
- Kafka Tests for Inventory (T008) and Order (T009) can be developed concurrently.

### Implementation Strategy

Follow the priority order. Implement the foundational `main.go` scripts, then build API tests, followed by the Kafka integrations, and finally the comprehensive E2E journey. Crucially, **ensure every validation step (e.g., T003, T007, T011, T013) is successfully executed** prior to proceeding, fulfilling the mandate for rigorously validated work.