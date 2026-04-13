# Project

This project implements two services:
* Order Service
* Payment Service

---

# How to run?

```sh
$ git clone https://github.com/Fipaan/ap2-uni.git
$ cd ap2-uni/op-assign/
$ go build -o nob
$ ./nob -clean            # clean all databases
$ ./nob -l                # list all services
$ ./nob -s <service-name> # start service
```

---

# Architecture

## Top Architecture

Because services are independent they have small common layer to not accidently collide with each other (like services' port). But basically, it just two compound structures that doesn't share code. Common configuration module sets-up environment variables, and fallbacks to default ones if necessary.

## Service Architecture

```
Each service uses Clean Architecture with separation of concerns.
├── cmd/ - service entry
├── internal/ - inner code, that used by service
│   ├── app/ - wiring layer, that combines everything in coherent structure
│   ├── client/ - code, that interacts with other services
│   ├── domain/ - data structures, that used across service
│   ├── repo/ - direct database interaction
│   ├── transport/ - transport layer for different kinds of communication
│   │   ├── grpc/ - actual interaction between services using gRPC for fast communication
│   │   └── http/ - user exposed HTTP communication
│   ├── usecase/ - verifies data and performs actions on database
│   └── proto/ (v1/) - gRPC `.proto` files that implement services, message structures and etc.
└── migrations/ - database files that define database entities and upgrades them on new changes
```

---

# Bounded Context

Even if codebases are separated, we still have some boundaries that allows us to interact between services.
For example, `order-service` depends on `payment-service`, but because we expect them to be independent,
we can't use `payment-service` directly, instead we are sending gRPC request that tries to retrieve data
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
* ...

That way you can clearly identify root of the problem, and take specific action if needed, for example, use another service.

Also gRPC exposes constant errors and doesn't rely on such kind of errors. It uses:
* codes.Unavailable      - service is currently unavailable
* codes.DeadlineExceeded - operation expired before completion
* ...
