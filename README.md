# GopherTrade

GopherTrade é uma plataforma de e-commerce baseada em microsserviços escrita em Go. O sistema é dividido internamente entre os microsserviços de **Order** e **Inventory**, utilizando padrões arquiteturais robustos como Kafka Event Streaming, CQRS e rastreamento distribuído através de OpenTelemetry.

---

## 🛠 Pré-requisitos e Dependências

A aplicação foi desenhada para facilitar o onboarding. Tudo que você precisa ter instalado no seu host local é:

1. **Go (1.21+)** - Para build e compilação (*Certifique-se de configurar seu `$GOPATH` corretamente*).
2. **Docker e Docker Compose** - Essencial para rodar toda a malha de infraestrutura auxiliar.
3. Dependências internas da infra (Serão orquestradas sozinhas via containers):
	- PostgreSQL (Banco de Escrita Primário)
	- Apache Kafka + Zookeeper (Event Broker)
	- Elasticsearch / NoSQL (Motores de CQRS/Busca)
	- Jaeger (Motor de OpenTelemetry)

---

## 🚀 Como Executar o Projeto

1. **Inicie a infraestrutura e os bancos**
Rodando esse comando você ativará o Compose com todo o scaffolding e injetará as migrações (`.sql`) nativamente dentro dos volumes do PostgreSQL assim que criados:
```bash
docker compose down -v  # Para limpar sujeiras passadas
docker compose up -d    # Subida daemon
```

2. **Baixe as dependências do Go**
GopherTrade funciona num Worktree (Mono-repo logico). Instale os módulos individualmente:
```bash
cd inventory && go mod tidy
cd ../order && go mod tidy
```

3. **Suba as aplicações independentemente**
Em terminais separados suba os dois workers. *Atenção: Garanta que o `.env` esteja na raiz do repositório para a lib `ardanlabs/conf/v3` mapeá-los nativamente*
```bash
# Terminal 1 - Inventory
cd inventory
go run cmd/server/main.go

# Terminal 2 - Order
cd order
go run cmd/server/main.go
```

**As portas padrão configuradas no seu .env serão:**
- HTTP Order API: `8082`
- HTTP Inventory API: `8081`
- gRPC interno (Inventory): `9091`

---

## 🧪 Como rodar Testes
Utilizamos testes automatizados usando a biblioteca padrão do golang `testing`.

Para executar a suite completa na base da aplicação (Incluindo Integration tests Mockados):
```bash
# Na base de ambos as pastas
go test ./... -v
```
 
---
📝 *Um arquivo de **Postman Collection** pode ser encontrado também na pasta raiz (gophertrade.postman_collection.json) contendo todos requests possíveis!*
