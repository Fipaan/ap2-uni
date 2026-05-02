This project implements three microservices connected via gRPC and RabbitMQ, following Event-Driven Architecture (EDA) principles.

```
[Client] --HTTP--> [Order Service] --gRPC--> [Payment Service] --AMQP--> [RabbitMQ] --AMQP--> [Notification Service]
```

---

# Services

| Service              | Role |
|----------------------|------|
| **Order Service**    | Accepts orders via HTTP, calls Payment Service via gRPC, streams status updates |
| **Payment Service**  | Processes payments via gRPC, publishes events to RabbitMQ |
| **Notify Service**   | Consumes payment events from RabbitMQ, sends email notifications |

---

# How to Run

### Local (requires PostgreSQL + RabbitMQ running)
```sh
git clone https://github.com/Fipaan/ap2-uni.git
cd ap2-uni/
go build -o nob
./nob -clean            # initialize all databases
./nob -l                # list all services
./nob -s <service-name> # start a service (order | payment | notify)
```

### Docker (recommended)
```sh
docker compose up --build
```

All services, databases, and RabbitMQ are orchestrated automatically. Databases are initialized on first start.

---

# Architecture

## Event Flow

```
POST /orders
    -> Order Service (HTTP)
        -> Payment Service (gRPC: ProcessPayment)
            -> payments DB
            -> RabbitMQ publish: payments.events / payment.completed
                -> Notification Service (AMQP consumer)
                    -> log: [Notify] Sent email to ...
```

## Service Architecture

Each service follows Clean Architecture:

```
├── cmd/                  - entry point
├── internal/
│   ├── app/              - wiring (dependency injection)
│   ├── domain/           - data structures
│   ├── repo/             - database access
│   ├── usecase/          - business logic and validation
│   ├── transport/
│   │   ├── grpc/         - gRPC server/client
│   │   └── http/         - HTTP handlers (order-service only)
│   └── infrastructure/
│       └── mq/           - RabbitMQ producer/consumer
├── migrations/           - SQL migration files
└── proto/v1/             - protobuf definitions
```

## Infrastructure

| Component             | Technology        |
|-----------------------|-------------------|
| Order ↔ Payment       | gRPC              |
| Payment → Notify      | RabbitMQ (AMQP)   |
| Order status updates  | PostgreSQL LISTEN/NOTIFY + gRPC streaming |
| Databases             | PostgreSQL (separate DB per service) |
| Containerization      | Docker Compose    |

---

# Reliability & Delivery Guarantees

| Concern           | Implementation |
|-------------------|---------------|
| **Manual ACKs**   | Notification Service ACKs only after successful processing |
| **Durable queues**| Queue and exchange declared with `durable=true` |
| **Persistent messages** | `DeliveryMode=Persistent` on all published messages |
| **Publisher confirms** | Payment Service waits for broker ACK before returning |
| **Idempotency**   | Notification Service deduplicates by `event_id` (in-memory store) |
| **DLQ**           | Messages failing after 3 retries are routed to `payment.completed.dlq` |
| **Graceful shutdown** | All services handle `SIGINT`/`SIGTERM` via `os/signal` |

---

# Idempotency Strategy

**Order Service**: deduplicates via `Idempotency-Key` HTTP header. Enforced both in application logic and via a unique DB index on `idempotency_key`.

**Notification Service**: deduplicates via `event_id` field in the event payload. On receipt, checks an in-memory store before processing. If already seen, ACKs and skips.

---

# ACK Logic

The Notification Service uses **manual acknowledgment**:
1. Message is received from RabbitMQ with `auto-ack=false`.
2. Message is deserialized and checked for duplicates.
3. Notifier (`EmailNotifier.Send`) is called.
4. **Only on success**: `msg.Ack(false)` is called.
5. On failure: retry counter is incremented and message is re-published. After 3 retries, moved to DLQ.

This guarantees **at-least-once delivery** — if the service crashes mid-processing, the unACKed message is redelivered by RabbitMQ.

---

# Failure Handling

HTTP error codes (Order Service):

| Code  | Meaning |
|-------|---------|
| `400` | Invalid input |
| `404` | Not found |
| `409` | Conflict (e.g. cancel non-pending order) |
| `503` | Downstream service unavailable |
| `500` | Internal error |

gRPC error codes (Payment Service):

| Code              | Meaning |
|-------------------|---------|
| `InvalidArgument` | Bad input |
| `Unavailable`     | Service unreachable |
| `DeadlineExceeded`| Timeout |
| `Internal`        | Unexpected error |
