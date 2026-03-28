# System Architecture & Design (GopherTrade)

A arquitetura do GopherTrade prioriza desacoplamento, rastreamento ativo e leituras ultrarrápidas distribuídas num ecossistema políglota de bancos de dados. Esta é The Big Picture:

---

## 🏗 Topologia dos Microsserviços

O design é dividido pela granularidade dos seus domínios core:

### 1. Inventory Service
- **Responsabilidade**: Detém a "SOT" (Source of truth) absoluta de catálogo de produtos e quantitativos físicos (stock).
- **Interface Pública**: RESTful HTTP (8081) para cadastros, e **gRPC Server** (9091) focado exclusivamente na ponte síncrona interna para abater rapidamente o estoque com o `BatchUpdateStock`.
- **Database**: PostgreSQL (`inventory` schema).

### 2. Order Service
- **Responsabilidade**: Lidar com carrinhos e pedidos consolidados e interface principal do e-commerce. 
- **Interface Pública**: Somente HTTP (8082), unificado pelas rotas lógicas transparentes `/api/v1/orders`.
- **Database**: 
	- PostgreSQL (`order` schema) **(Writer)**
	- Firestore & Elasticsearch **(Readers)**

---

## 🏛 Design Patterns e Decisões

### 1. CQRS (Command Query Responsibility Segregation) c/ Consistência Eventual
O projeto abraçou o CQRS de maneira **transparente** na camada API. O cliente consome tudo de `/api/v1/orders/`, porem nós aplicamos a segregação por debaixo dos panos:
- **Write Path (Command)**: Quando um pedido é feito (`POST`), gravamos unicamente no banco super consistente e transacional (PostgreSQL).
- **Async Event Bus**: Assim que essa persistência é confirmada, disparamos um evento `OrderCreated` pro tópico `orders` do **Apache Kafka**.
- **Read Path (Query)**: Duas *background routines* ("consumers") capturam esse evento e indexam passivamente as projeções numectificadas.
	- Se você der `GET /{id}` o framework desvia pra pegar milissegundos do **Firestore** (perfeito em buscar key-value unitários).
	- Se você dar request de busca contendo string status (ex: `?q=Pendente`), ele é despachado pro **Elasticsearch**, otimizado estritamente contra table-scans demoradas.

### 2. Comunicação Híbrida Inteligente
Em vez de focar tudo apenas em coreografia assíncrona, a aplicação sabe a hora de travar. 
Antes de um Order ser oficialmente gerado num `POST`, ele liga via síncrono **(gRPC Call)** em milissegundos ao app `Inventory` para abater o estoque do produto. Somente com sucesso a transação roda. Trata-se do *Orchestration Mode* híbrido que protege consistência de venda fantasma.

### 3. Ardanlabs Config Management (The 12-Factor App)
As definições como IPs e Portas seguem strict parity baseada na metodologia "12 factor app" lendo as primitivas do OS provindas de um `.env` central sob a library de struct bonding `conf/v3`. 

### 4. Distributed Tracing c/ OpenTelemetry (OTEL)
Para não ficarmos cegos no breu dos containers e dos event busses, nós subimos um endpoint da Jaeger (`localhost:16686`).
Toda transação gera um Span encapsulado no provider do OTEL que injeta headers tanto numa chamada GPRC normal (propagador B3) quanto nas filas de Evento, costurando o painel visual fim-a-fim.
