Allows you to interact (create, access and cancel) with orders.

# HTTP entries

* POST:  `/orders`            - Accepts json with `customer_id`, `item_name`, `amount`, and registers purchase. Give small window to allow user cancel purchase
* GET:   `/orders/:id`        - Allows user to access existing order
* PATCH: `/orders/:id/cancel` - Allows user to cancel existing order if it still pending or small window for cancel didn't passed.

# Idempotency

If you supply `Idempotency-Key` header, it will protect order entry from duplication, and will return existing entry on multiple orders
