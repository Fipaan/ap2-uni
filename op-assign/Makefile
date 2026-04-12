.PHONY: clean
clean:
	psql -U postgres \
		-c "DROP DATABASE order_db; " \
		-c "DROP DATABASE payment_db;" \
		-c "CREATE DATABASE order_db; " \
		-c "CREATE DATABASE payment_db;" \
		-c "\q"
	psql -U postgres -d order_db -f ./order-service/migrations/init.sql
	psql -U postgres -d payment_db -f ./payment-service/migrations/init.sql

.PHONY: order
order:
	cd ./order-service && \
	go mod tidy && \
	go run ./cmd

.PHONY: clean-order
clean-order: clean order

.PHONY: payment
payment:
	cd ./payment-service && \
	go mod tidy && \
	go run ./cmd

.PHONY: clean-payment
clean-payment: clean payment
