# Feature Specification: Backend Order and Stock Management

**Feature Branch**: `002-order-stock-management`  
**Created**: 2026-03-21  
**Status**: Draft  
**Input**: User description: "Construa uma aplicação de backend capaz de gerenciar pedidos. Todo pedido deve conter um ou mais itens, de diferentes quantidades, com um preço total associado. O usuário deve ser capaz de gerenciar tanto os pedidos quanto o estoque dos produtos. Um pedido não pode ser criado caso o item solicitado nele não exista ou não esteja disponível na quantidade solicitada no pedido. Foque apenas nas regras de negócio da aplicação."

## Clarifications

### Session 2026-03-21
- Q: How should monetary values be stored to prevent precision errors? → A: **Integers (Cents)**: Store prices as the smallest currency unit (e.g., cents).
- Q: How should concurrent stock updates be handled? → A: **Optimistic Locking**: Use a version number on the `Product` table. Updates fail if the version has changed.
- Q: Should order prices change if product prices change later? → A: **No (Snapshot)**: Order items must store the product unit price at the time of creation.
- Q: Which data access pattern should be used? → A: **Repository Pattern**: Abstract data access behind interfaces (e.g., `ProductRepository`). Best for DDD/Clean Code.
- Q: Estratégia de Atomicidade no Decremento de Estoque → A: **Batch gRPC no Inventory**: O serviço de Pedidos envia todos os itens em uma única chamada gRPC. O Inventário processa tudo em uma única transação SQL (rollback se algum falhar).
- Q: Qual biblioteca HTTP/REST deve ser utilizada? → A: **Chi**: Framework leve, idiomático e focado na biblioteca padrão (`net/http`). Ideal para middlewares e roteamento limpo.
- Q: Qual ferramenta para gRPC deve ser utilizada? → A: **Buf**: Ferramenta moderna para gerenciar Protobuf, incluindo linting e geração de código simplificada.
- Q: Qual o escopo da Autenticação/Autorização para este protótipo? → A: **Mock/Placeholder**: Middleware simples que aceita `User-ID` via header para focar nas regras de negócio de pedidos e estoque.
- Q: Quais bancos de dados para modelos de leitura (CQRS) devem ser utilizados? → A: **Firestore + ElasticSearch**: Implementação completa para consultas por ID (Firestore) e buscas avançadas (ES), sincronizados via Kafka.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Manage Products and Stock (Priority: P1)

As a user, I want to create products and adjust their stock levels so that I have an inventory to sell.

**Why this priority**: Products and stock must exist before orders can be processed. This is the foundational data layer.

**Independent Test**: Can be fully tested by creating a product, adding stock, and verifying the stock count persists.

**Acceptance Scenarios**:

1. **Given** a new product detail (name, price), **When** I request to create it, **Then** the product is saved with 0 initial stock.
2. **Given** an existing product, **When** I add quantity X to its stock, **Then** the available quantity increases by X.
3. **Given** an existing product, **When** I remove quantity X from its stock, **Then** the available quantity decreases by X.

---

### User Story 2 - Create Valid Orders (Priority: P1)

As a user, I want to create an order with multiple items so that I can process sales, but only if stock permits.

**Why this priority**: This is the core business value of the application—selling items while enforcing inventory constraints.

**Independent Test**: Can be tested by attempting to create orders with valid and invalid quantities and checking success/failure responses.

**Acceptance Scenarios**:

1. **Given** products A and B exist with sufficient stock, **When** I create an order for 1 of A and 2 of B, **Then** the order is created, the total price is calculated correctly, and stock for A and B is reduced.
2. **Given** product A has 5 items in stock, **When** I attempt to order 6 of A, **Then** the order creation fails with an error and stock remains unchanged.
3. **Given** I attempt to order a product ID that does not exist, **When** I submit the order, **Then** the creation fails with an error.
4. **Given** I attempt to create an order with an empty item list, **When** I submit the order, **Then** the creation fails (must have one or more items).

---

### User Story 3 - View Orders (Priority: P2)

As a user, I want to view the details of created orders so that I can track what has been sold.

**Why this priority**: Visibility into past transactions is essential for management but secondary to the ability to actually process them.

**Independent Test**: Can be tested by creating orders and then querying the list/details to ensure data matches.

**Acceptance Scenarios**:

1. **Given** existing orders, **When** I request a list of all orders, **Then** I receive a summary of orders including their IDs, total prices, and dates.
2. **Given** a specific order ID, **When** I request its details, **Then** I receive the full item list, quantities, unit prices at time of purchase, and total price.

### Edge Cases

- What happens when two orders try to buy the last item simultaneously? (Concurrency handling expected: one succeeds, one fails).
- How does the system handle negative stock updates? (Should be prevented or handled as "removal").
- What if a product price changes after an order is created? (Order history should preserve the price at the moment of creation).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST allow creating a product with a name and unit price.
- **FR-002**: System MUST allow updating the stock quantity of a product (increment/decrement).
- **FR-003**: System MUST prevent stock from becoming negative.
- **FR-004**: System MUST allow creating an order containing a list of item IDs and quantities.
- **FR-005**: System MUST calculate the total price of the order automatically based on product unit prices.
- **FR-006**: System MUST reject order creation if any requested item ID does not exist.
- **FR-007**: System MUST reject order creation if any requested item has insufficient stock.
- **FR-008**: System MUST atomically decrement stock for all items in an order via a single transactional batch update in the Inventory service.
- **FR-009**: System MUST require at least one item per order.
- **FR-010**: System MUST allow retrieving a list of all orders.
- **FR-011**: System MUST allow retrieving details of a specific order by ID.
- **FR-012**: System MUST store and calculate all monetary values as integers representing the smallest currency unit (e.g., cents) to avoid floating-point errors.
- **FR-013**: System MUST implement optimistic locking for stock updates to handle concurrent order creation safely.
- **FR-014**: System MUST snapshot the unit price of each product into the order item at the time of creation to ensure historical accuracy.

### Key Entities *(include if feature involves data)*

- **Product**: Represents an item for sale. Attributes: ID, Name, Unit Price (in cents), Current Stock Quantity, Version (for optimistic locking).
- **Order**: Represents a completed transaction. Attributes: ID, Date, Total Price (in cents), Status (e.g., Created).
- **OrderItem**: Link between Order and Product. Attributes: Product ID, Quantity, Unit Price (snapshot in cents).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of successful orders result in correct stock deduction (verified via audit test).
- **SC-002**: 0% of orders are created if they exceed available stock.
- **SC-003**: Order total price calculation is accurate to the cent for 100% of transactions.
- **SC-004**: System handles concurrent order requests for the same item without data corruption (race condition safety).
