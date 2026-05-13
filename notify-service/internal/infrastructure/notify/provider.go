package notify

import (
    "context"
    "github.com/Fipaan/ap2-uni/notify-service/internal/domain"
)

type Provider interface {
    Send(ctx context.Context, evt domain.PaymentCompletedEvent) error
}
