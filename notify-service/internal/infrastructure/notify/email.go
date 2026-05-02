package notify

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/Fipaan/ap2-uni/notify-service/internal/domain"
)

type EmailNotifier struct{}

func NewEmailNotifier() *EmailNotifier { return &EmailNotifier{} }

func (e *EmailNotifier) Send(ctx context.Context, evt domain.PaymentCompletedEvent) error {
	_ = ctx

	// demo failure to show retries + DLQ
	if strings.EqualFold(evt.CustomerEmail, "fail@example.com") {
		return errors.New("simulated permanent error")
	}

	log.Printf("[Notify] preparing email to %s", evt.CustomerEmail)
	return nil
}
