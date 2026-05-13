# Order Service

Allows you to interact (create, access and cancel) with orders. Communicates with Payment Service via gRPC to process payments, and streams order status updates to subscribers via gRPC streaming backed by PostgreSQL `LISTEN/NOTIFY`.

---

# HTTP Endpoints

| Method  | Path                  | Description |
|---------|-----------------------|-------------|
| `POST`  | `/orders`             | Create a new order. Triggers payment processing via gRPC. Returns order with current status. |
| `GET`   | `/orders/:id`         | Fetch an existing order by ID. |
| `GET`   | `/orders/list`        | List payments (proxied from Payment Service). Accepts optional `?status=` query param. |
| `PATCH` | `/orders/:id/cancel`  | Cancel an order if it is still `Pending`, or within the cancellation window for `Authorized` orders. |

### POST `/orders` — Request Body
```json
{
  "customer_id": "cust-1",
  "customer_email": "user@example.com",
  "item_name": "server",
  "amount": 9999
}
```

---

# Idempotency

Supply an `Idempotency-Key` header to protect against duplicate order creation. If the same key is submitted again, the existing order is returned with HTTP `200` instead of creating a new one.

Idempotency is enforced at two levels:
- **Application layer**: checks for existing key before insert.
- **Database layer**: unique index on `idempotency_key` column as a safety net for race conditions.

---

# Order Lifecycle

```
Pending -> Authorized -> Paid
                      -> Failed
       -> Failed
       -> Cancelled
```

- Orders start as `Pending`.
- Payment Service responds with `Authorized` or `Declined`.
- A background goroutine transitions `Authorized` orders to `Paid` after a short delay.
- Orders can be cancelled while `Pending`, or within a short window after `Authorized`.

---

# gRPC

Exposes `SubscribeToOrderUpdates(OrderRequest) returns (stream OrderStatusUpdate)` — allows clients to stream real-time status changes for a given order ID. Updates are pushed via PostgreSQL `LISTEN/NOTIFY` trigger on the `orders` table.

---

# Error Codes

| Code  | Meaning |
|-------|---------|
| `400` | Invalid input (missing customer ID, item name, or non-positive amount) |
| `404` | Order not found |
| `409` | Order cannot be cancelled (wrong state or window expired) |
| `503` | Payment Service unavailable |
| `500` | Internal error |

---

# Graceful Shutdown

Uses `os/signal` with `SIGINT`/`SIGTERM`. On shutdown:
- HTTP server drains active connections (5s timeout).
- gRPC server stops gracefully.
- Payment gRPC client connection is closed.
- PostgreSQL listener is stopped.
- Redis cache connection is closed.

# Caching

Order Service uses a **cache-aside** pattern backed by Redis:
- `GET /orders/:id` checks Redis before hitting the DB.
- Cache entries have a configurable TTL (default `5m`, set via `CACHE_TTL` env var).
- Cache is invalidated on every status update to prevent stale reads.
