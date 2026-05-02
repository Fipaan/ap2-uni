# Payment Service

Handles payment processing. Exposes a gRPC interface consumed by Order Service. After a successful payment, publishes a `PaymentCompletedEvent` to RabbitMQ for downstream consumers (e.g. Notification Service).

---

# gRPC Endpoints

### `ProcessPayment(PaymentRequest) returns (PaymentResponse)`

Processes a payment for a given order.

**Request:**
```protobuf
message PaymentRequest {
  string order_id       = 1;
  int64  amount         = 2;
  string customer_email = 3;
}
```

**Response:**
```protobuf
message PaymentResponse {
  string payment_id     = 1;
  string transaction_id = 2;
  string status         = 3;
  Timestamp processed_at = 4;
}
```

Status is `Authorized` if `amount <= 100000`, otherwise `Declined`.

---

### `ListPayments(ListPaymentsRequest) returns (ListPaymentsResponse)`

Returns all payments, optionally filtered by status.

**Request:**
```protobuf
message ListPaymentsRequest {
  string status = 1; // optional
}
```

---

# Event Publishing (RabbitMQ)

After a payment is committed to the database, an event is published to the `payments.events` exchange with routing key `payment.completed`.

**Payload:**
```json
{
  "event_id":       "uuid",
  "order_id":       "uuid",
  "amount":         9999,
  "customer_email": "user@example.com",
  "status":         "Authorized",
  "created_at":     "2026-05-03T00:00:00Z"
}
```

**Reliability guarantees:**
- Exchange and messages are durable — survive broker restart.
- Publisher confirms are enabled — publish blocks until broker ACKs or times out (5s).
- `DeliveryMode` is set to `Persistent`.

---

# Error Codes (gRPC)

| Code                | Meaning |
|---------------------|---------|
| `InvalidArgument`   | Empty order ID or non-positive amount or invalid email |
| `Internal`          | Unexpected error |

---

# Graceful Shutdown

Uses `os/signal` with `SIGINT`/`SIGTERM`. On shutdown:
- gRPC server stops gracefully.
- RabbitMQ channel and connection are closed cleanly via `Publisher.Close()`.
