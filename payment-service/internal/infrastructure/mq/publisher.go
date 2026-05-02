package mq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/Fipaan/ap2-uni/payment-service/internal/domain"
)

type Publisher struct {
	conn     *amqp.Connection
	ch       *amqp.Channel
	confirms <-chan amqp.Confirmation
	exchange string
	mu       sync.Mutex
}

func NewPublisher(amqpURL, exchange string) (*Publisher, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("open channel: %w", err)
	}

	if err := ch.ExchangeDeclare(
		exchange,
		"topic",
		true,  // durable
		false, // auto-delete
		false,
		false,
		nil,
	); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("declare exchange: %w", err)
	}

	if err := ch.Confirm(false); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("enable confirms: %w", err)
	}

	confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 1))

	return &Publisher{
		conn:     conn,
		ch:       ch,
		confirms: confirms,
		exchange: exchange,
	}, nil
}

func (p *Publisher) PublishPaymentCompleted(ctx context.Context, evt *domain.PaymentCompletedEvent) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	body, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	if err := p.ch.PublishWithContext(
		ctx,
		p.exchange,
		"payment.completed",
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode:  amqp.Persistent,
			MessageId:     evt.EventID,
			Timestamp:     time.Now().UTC(),
			Body:          body,
			Headers:       amqp.Table{"event_id": evt.EventID},
		},
	); err != nil {
		return fmt.Errorf("publish: %w", err)
	}

	select {
	case conf := <-p.confirms:
		if !conf.Ack {
			return errors.New("broker nacked published message")
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		return errors.New("publish confirm timeout")
	}
}

func (p *Publisher) Close() error {
	var first error
	if p.ch != nil {
		if err := p.ch.Close(); err != nil && first == nil {
			first = err
		}
	}
	if p.conn != nil {
		if err := p.conn.Close(); err != nil && first == nil {
			first = err
		}
	}
	return first
}
