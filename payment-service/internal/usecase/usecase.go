package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Fipaan/ap2-uni/payment-service/internal/domain"

	"github.com/google/uuid"
)

var ErrInvalidOrderID = errors.New("order id is empty")
var ErrInvalidAmount  = errors.New("amount must be > 0")
var ErrInvalidEmail   = errors.New("customer email is invalid")

type PaymentRepository interface {
	Create(ctx context.Context, p *domain.Payment) error
	GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error)
	GetAll(ctx context.Context) (*[]domain.Payment, error)
	GetAllByStatus(ctx context.Context, status string) (*[]domain.Payment, error)
}

type EventPublisher interface {
	PublishPaymentCompleted(ctx context.Context, p *domain.PaymentCompletedEvent) error
}

type PaymentUsecase struct {
	repo      PaymentRepository
	publisher EventPublisher
}

func NewPaymentUsecase(r PaymentRepository, pub EventPublisher) *PaymentUsecase {
	return &PaymentUsecase{
		repo:      r,
		publisher: pub,
	}
}

func validateProcess(orderID string, amount int64, customerEmail string) error {
	if orderID == "" {
		return ErrInvalidOrderID
	}
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if !strings.Contains(customerEmail, "@") {
		return ErrInvalidEmail
	}
	return nil
}

func (uc *PaymentUsecase) Process(ctx context.Context, orderID string, amount int64, customerEmail string) (p *domain.Payment, err error) {
	err = validateProcess(orderID, amount, customerEmail)
	if err != nil { return }

	status := "Authorized"
	if amount > 100000 { status = "Declined" }

	p = &domain.Payment{
		ID:            uuid.New().String(),
		OrderID:       orderID,
		TransactionID: uuid.New().String(),
		Amount:        amount,
		Status:        status,
		CustomerEmail: customerEmail,
	}

	err = uc.repo.Create(ctx, p)
	if err != nil { return }

	evt := domain.PaymentCompletedEvent{
		EventID:       uuid.NewString(),
		OrderID:       p.OrderID,
		Amount:        p.Amount,
		CustomerEmail: p.CustomerEmail,
		Status:        p.Status,
		CreatedAt:     time.Now().UTC(),
	}
	
	if err := uc.publisher.PublishPaymentCompleted(ctx, evt); err != nil {
		err = fmt.Errorf("publish payment completed: %w", err)
		return
	}

	return p, nil
}

func (uc *PaymentUsecase) List(ctx context.Context, status string) (p *[]domain.Payment, err error) {
	if status == "" {
		return uc.repo.GetAll(ctx)
	} else {
		return uc.repo.GetAllByStatus(ctx, status)
	}
}

func (uc *PaymentUsecase) Get(ctx context.Context, orderID string) (*domain.Payment, error) {
	return uc.repo.GetByOrderID(ctx, orderID)
}
