package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Fipaan/ap2-uni/config"
	"github.com/Fipaan/ap2-uni/notify-service/internal/idempotency"
	"github.com/Fipaan/ap2-uni/notify-service/internal/infrastructure/mq"
	"github.com/Fipaan/ap2-uni/notify-service/internal/infrastructure/notify"
)

func main() {
	amqpURL   := config.RabbitMQAddr()
	redisAddr := config.RedisAddr()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

    redisClient := redis.NewClient(&redis.Options{Addr: redisAddr})
    defer redisClient.Close()

	store    := idempotency.NewStore(redisClient, 24*time.Hour)
	notifier := notify.NewFromEnv()

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
