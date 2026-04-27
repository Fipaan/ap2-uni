# Purpose

Allows you to interact (create and access) with payment entries.

# HTTP entries

* POST:  `/payments`           - Accepts json with `order_id`, `amount`, and registers payment entry.
* GET:   `/payments/:order_id` - Allows user to access existing payment entry
