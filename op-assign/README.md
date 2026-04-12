# Project

This project implements two services:
* Order Service
* Payment Service

---

# Architecture

## Top Architecture

Because services are independent they have small common layer to not accidently collide with each other (like services' port).
But basically, it just two compound structures that doesn't share code.

## Service Architecture

```
Each service uses Clean Architecture with separation of concerns.
├── cmd/ - service entry
├── internal/ - inner code, that used by service
│   ├── app/ - wiring layer, that combines everything in coherent structure
│   ├── client/ - code, that interacts with other services
│   ├── domain/ - data structures, that used across service
│   ├── repo/ - direct database interaction
│   ├── transport/ (http/) - connecting HTTP requests/responses to actual data manipulation
│   └── usecase/ - verifies data and performs actions on database
└── migrations/ - database files that define database entities and upgrades them on new changes
```

---

# Bounded Context

Even if codebases are separated, we still have some boundaries that allows us to interact between services.
For example, `order-service` depends on `payment-service`, but because we expect them to be independent,
we can't use `payment-service` directly, instead we are doing HTTP request that tries to retrieve data
from that service, in result we still get data, but if service is not available, we are trying to handle it
already on our side (e.g. send appropriate error).

---

# Failure Handling

We are working with HTTP, so we need to handle errors appropriately. I'm utilizng error messages and error codes:
* 400 - bad input data
* 404 - not found
* 409 - data conflict
* 500 - internal errors (to not leak any sensitive data)
* 503 - service is not available (when interacting with different services)

That way you can clearly identify root of the problem, and take specific action if needed, for example, use another service.
