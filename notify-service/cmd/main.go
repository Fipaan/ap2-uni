package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Fipaan/ap2-uni/notify-service/internal/idempotency"
	"github.com/Fipaan/ap2-uni/notify-service/internal/infrastructure/mq"
	"github.com/Fipaan/ap2-uni/notify-service/internal/infrastructure/notify"
)

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	amqpURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	store := idempotency.NewStore()
	notifier := notify.NewEmailNotifier()

	consumer, err := mq.NewConsumer(amqpURL, store, notifier)
	if err != nil {
		log.Fatal(err)
	}
	defer consumer.Close()

	log.Println("notify service started")
	if err := consumer.Start(ctx); err != nil {
		log.Fatal(err)
	}
	log.Println("notify service stopped")
}
