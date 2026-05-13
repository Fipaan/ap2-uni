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

Uses Redis (`idempotency.Store`) keyed by `event_id` with a 24h TTL. If the same event is delivered twice, the second delivery is ACKed and skipped. Persists across service restarts.

---

# Dead Letter Queue (Bonus)

Implements a DLQ for permanent failures:

- Main exchange: `payments.events` (topic)
- DLQ exchange: `payments.dlx` (direct)
- DLQ queue: `payment.completed.dlq`

**Retry logic:**
- On processing failure, the worker sleeps with **exponential backoff** (2s, 4s, 8s) before re-publishing.
- After **3 failed attempts**, the message is moved to the DLQ with an `x-dead-letter-reason` header explaining the failure.

**To simulate a DLQ scenario**, send an order with `customer_email: "fail@example.com"` — the notifier will permanently fail for this address.

---

# Graceful Shutdown

Uses `os/signal` with `SIGINT`/`SIGTERM`. On shutdown:
- Consumer loop exits cleanly.
- RabbitMQ channel and connection are closed via `consumer.Close()`.

# Notification Provider

Uses the **Adapter Pattern** — business logic depends on a `Provider` interface, not a concrete implementation.

Set `PROVIDER_MODE` env var to switch:
| Value       | Behavior |
|-------------|----------|
| `SIMULATED` | Logs notification, simulates latency (100-500ms) and 20% random transient failures |
| `REAL`      | SMTP/Mailjet integration (not yet implemented) |
