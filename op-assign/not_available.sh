#!/usr/bin/bash

set -xe

curl -X POST http://localhost:8000/orders \
      -H "Content-Type: application/json" \
      -d '{
    "customer_id": "cust-1",
    "item_name": "server",
    "amount": 200000
  }' | jq
