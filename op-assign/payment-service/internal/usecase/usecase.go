package usecase

import (
	"context"
	"errors"

	"github.com/fipaan/ap2-uni/op-assign/payment-service/internal/domain"

	"github.com/google/uuid"
)

var ErrInvalidOrderID = errors.New("order id is empty")
var ErrInvalidAmount  = errors.New("amount must be > 0")

type PaymentRepository interface {
	Create(ctx context.Context, p *domain.Payment) error
	GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error)
	GetAll(ctx context.Context) (*[]domain.Payment, error)
	GetAllByStatus(ctx context.Context, status string) (*[]domain.Payment, error)
}

type PaymentUsecase struct {
	repo PaymentRepository
}

func NewPaymentUsecase(r PaymentRepository) *PaymentUsecase {
	return &PaymentUsecase{repo: r}
}

func validateProcess(orderID string, amount int64) error {
	if orderID == "" {
		return ErrInvalidOrderID
	}
	if amount <= 0 {
		return ErrInvalidAmount
	}
	return nil
}

func (uc *PaymentUsecase) Process(ctx context.Context, orderID string, amount int64) (p *domain.Payment, err error) {
	err = validateProcess(orderID, amount)
	if err != nil { return }

	status := "Authorized"
	if amount > 100000 { status = "Declined" }

	p = &domain.Payment{
		ID:            uuid.New().String(),
		OrderID:       orderID,
		TransactionID: uuid.New().String(),
		Amount:        amount,
		Status:        status,
	}

	err = uc.repo.Create(ctx, p)
	return p, err
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
