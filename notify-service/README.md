# Notification Service

Consumes `PaymentCompletedEvent` messages from RabbitMQ and simulates sending email notifications. Fully decoupled — has no knowledge of Order or Payment services.

---

# Event Consumer

Listens on queue `payment.completed.q`, bound to exchange `payments.events` with routing key `payment.completed`.

On each message:
1. Deserializes the `PaymentCompletedEvent`.
2. Checks idempotency store — skips if already processed.
3. Calls notifier — logs the notification:
```
[Notify] Sent email to user@example.com for Order <order_id>. Amount: $99.99
```
4. ACKs the message only after successful processing.

---

# Reliability

| Feature | Implementation |
|---------|---------------|
| **Manual ACKs** | `auto-ack` is disabled. Message is ACKed only after the log is successfully printed. |
| **Durable queue** | `payment.completed.q` is declared with `durable=true` — survives broker restart. |
| **Persistent messages** | Producer sets `DeliveryMode=Persistent`. |
| **QoS** | Prefetch count set to `1` — processes one message at a time. |

---

# Idempotency

Uses an in-memory store (`idempotency.Store`) keyed by `event_id`. If the same event is delivered twice, the second delivery is ACKed and skipped without re-processing.

NOTE: store is in-memory. For production, it's better to use persistent store (e.g. Redis or a DB table).

---

# Dead Letter Queue (Bonus)

Implements a DLQ for permanent failures:

- Main exchange: `payments.events` (topic)
- DLQ exchange: `payments.dlx` (direct)
- DLQ queue: `payment.completed.dlq`

**Retry logic:**
- On processing failure, the message is re-published to the main exchange with an incremented `x-retry-count` header.
- After **3 failed attempts**, the message is moved to the DLQ with an `x-dead-letter-reason` header explaining the failure.

**To simulate a DLQ scenario**, send an order with `customer_email: "fail@example.com"` — the notifier will permanently fail for this address.

---

# Graceful Shutdown

Uses `os/signal` with `SIGINT`/`SIGTERM`. On shutdown:
- Consumer loop exits cleanly.
- RabbitMQ channel and connection are closed via `consumer.Close()`.
