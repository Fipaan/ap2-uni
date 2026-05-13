package notify

import (
    "context"
    "errors"
    "log"
    "math/rand"
    "strings"
    "time"

    "github.com/Fipaan/ap2-uni/notify-service/internal/domain"
)

type SimulatedProvider struct{}

func NewSimulatedProvider() *SimulatedProvider { return &SimulatedProvider{} }

func (s *SimulatedProvider) Send(ctx context.Context, evt domain.PaymentCompletedEvent) error {
    time.Sleep(time.Duration(100+rand.Intn(400)) * time.Millisecond) // simulate latency

    if strings.EqualFold(evt.CustomerEmail, "fail@example.com") {
        return errors.New("simulated permanent error")
    }
    if rand.Float32() < 0.2 { // 20% random transient failure
        return errors.New("simulated transient error")
    }

    log.Printf("[Notify] Sent email to %s for Order %s. Amount: $%.2f",
        evt.CustomerEmail, evt.OrderID, float64(evt.Amount)/100.0)
    return nil
}
