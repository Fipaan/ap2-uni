package config

import "os"

const (
	defaultOrderPORT       = "8000"
	defaultPaymentPORT     = "8080"
	defaultOrderGRPCAddr   = "127.0.0.1:9091"
	defaultPaymentGRPCAddr = "127.0.0.1:9090"
	defaultOrderDB_DSN     = "postgres://postgres:password@localhost:5432/order_db?sslmode=disable"
	defaultPaymentDB_DSN   = "postgres://postgres:password@localhost:5432/payment_db?sslmode=disable"
	defaultRMQAddr         = "amqp://guest:guest@rabbitmq:5672/"
)

func tryEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func OrderPORT() string {
	return tryEnv("ORDER_PORT", defaultOrderPORT)
}
func PaymentPORT() string {
	return tryEnv("PAYMENT_PORT", defaultPaymentPORT)
}
func OrderGRPCAddr() string {
	return tryEnv("ORDER_GRPC_ADDR", defaultOrderGRPCAddr)
}
func PaymentGRPCAddr() string {
	return tryEnv("PAYMENT_GRPC_ADDR", defaultPaymentGRPCAddr)
}
func OrderDB_DSN() string {
	return tryEnv("ORDER_DB_DSN", defaultOrderDB_DSN)
}
func PaymentDB_DSN() string {
	return tryEnv("PAYMENT_DB_DSN", defaultPaymentDB_DSN)
}
func RabbitMQAddr() string {
	return tryEnv("RABBITMQ_ADDR", defaultRMQAddr)
}
