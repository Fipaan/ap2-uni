package mq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/Fipaan/ap2-uni/notify-service/internal/domain"
	"github.com/Fipaan/ap2-uni/notify-service/internal/idempotency"
)

type Notifier interface {
	Send(ctx context.Context, evt domain.PaymentCompletedEvent) error
}

type Consumer struct {
	conn   *amqp.Connection
	ch     *amqp.Channel
	store  *idempotency.Store
	notifier Notifier

	mainExchange string
	dlqExchange  string
	mainQueue    string
	dlqQueue     string
}

func NewConsumer(amqpURL string, store *idempotency.Store, notifier Notifier) (*Consumer, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("open channel: %w", err)
	}

	c := &Consumer{
		conn:         conn,
		ch:           ch,
		store:        store,
		notifier:     notifier,
		mainExchange: "payments.events",
		dlqExchange:  "payments.dlx",
		mainQueue:    "payment.completed.q",
		dlqQueue:     "payment.completed.dlq",
	}

	if err := c.setup(); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}

	return c, nil
}

func (c *Consumer) setup() error {
	if err := c.ch.ExchangeDeclare(c.mainExchange, "topic", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare main exchange: %w", err)
	}
	if err := c.ch.ExchangeDeclare(c.dlqExchange, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare dlq exchange: %w", err)
	}

	_, err := c.ch.QueueDeclare(c.mainQueue, true, false, false, false, amqp.Table{
	    "x-dead-letter-exchange":    c.dlqExchange,
	    "x-dead-letter-routing-key": "payment.completed.dlq",
	})
	if err != nil {
		return fmt.Errorf("declare main queue: %w", err)
	}
	_, err = c.ch.QueueDeclare(c.dlqQueue, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("declare dlq queue: %w", err)
	}

	if err := c.ch.QueueBind(c.mainQueue, "payment.completed", c.mainExchange, false, nil); err != nil {
		return fmt.Errorf("bind main queue: %w", err)
	}
	if err := c.ch.QueueBind(c.dlqQueue, "payment.completed.dlq", c.dlqExchange, false, nil); err != nil {
		return fmt.Errorf("bind dlq queue: %w", err)
	}

	if err := c.ch.Qos(1, 0, false); err != nil {
		return fmt.Errorf("qos: %w", err)
	}

	return nil
}

func (c *Consumer) Start(ctx context.Context) error {
	deliveries, err := c.ch.Consume(
		c.mainQueue,
		"",
		false, // auto-ack disabled
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-deliveries:
			if !ok {
				return errors.New("delivery channel closed")
			}
			if err := c.handle(ctx, msg); err != nil {
				log.Printf("[Notify] handler error: %v", err)
			}
		}
	}
}

func (c *Consumer) handle(ctx context.Context, msg amqp.Delivery) error {
	var evt domain.PaymentCompletedEvent
	if err := json.Unmarshal(msg.Body, &evt); err != nil {
		return c.moveToDLQ(ctx, msg, "invalid json: "+err.Error())
	}

	if c.store.Seen(evt.EventID) {
		return msg.Ack(false)
	}

	if err := c.notifier.Send(ctx, evt); err != nil {
		c.store.Forget(evt.EventID)

		retryCount := retryFromHeaders(msg.Headers)
		if retryCount >= 3 {
			return c.moveToDLQ(ctx, msg, err.Error())
		}
		return c.republishForRetry(ctx, msg, retryCount+1)
	}

	log.Printf("[Notify] Sent email to %s for Order %s. Amount: $%.2f",
		evt.CustomerEmail, evt.OrderID, float64(evt.Amount)/100.0)

	return msg.Ack(false)
}

func (c *Consumer) republishForRetry(ctx context.Context, msg amqp.Delivery, retryCount int) error {
	headers := amqp.Table{}
	for k, v := range msg.Headers {
		headers[k] = v
	}
	headers["x-retry-count"] = retryCount

	if err := c.ch.PublishWithContext(
		ctx,
		c.mainExchange,
		"payment.completed",
		false,
		false,
		amqp.Publishing{
			ContentType:  msg.ContentType,
			DeliveryMode:  amqp.Persistent,
			Body:         msg.Body,
			MessageId:    msg.MessageId,
			Headers:      headers,
		},
	); err != nil {
		return fmt.Errorf("republish retry: %w", err)
	}

	return msg.Ack(false)
}

func (c *Consumer) moveToDLQ(ctx context.Context, msg amqp.Delivery, reason string) error {
	headers := amqp.Table{}
	for k, v := range msg.Headers {
		headers[k] = v
	}
	headers["x-dead-letter-reason"] = reason

	if err := c.ch.PublishWithContext(
		ctx,
		c.dlqExchange,
		"payment.completed.dlq",
		false,
		false,
		amqp.Publishing{
			ContentType:  msg.ContentType,
			DeliveryMode:  amqp.Persistent,
			Body:         msg.Body,
			MessageId:    msg.MessageId,
			Headers:      headers,
		},
	); err != nil {
		return fmt.Errorf("publish dlq: %w", err)
	}

	return msg.Ack(false)
}

func retryFromHeaders(h amqp.Table) int {
	raw, ok := h["x-retry-count"]
	if !ok {
		return 0
	}

	switch v := raw.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case byte:
		return int(v)
	case string:
		n, _ := strconv.Atoi(v)
		return n
	default:
		return 0
	}
}

func (c *Consumer) Close() error {
	var first error
	if c.ch != nil {
		if err := c.ch.Close(); err != nil && first == nil {
			first = err
		}
	}
	if c.conn != nil {
		if err := c.conn.Close(); err != nil && first == nil {
			first = err
		}
	}
	return first
}
