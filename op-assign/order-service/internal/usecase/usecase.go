package usecase

import (
	"context"
	"errors"
	"time"
	"database/sql"

	"github.com/fipaan/ap2-uni/op-assign/order-service/internal/domain"
	paymentdomain "github.com/fipaan/ap2-uni/op-assign/payment-service/internal/client/grpc"
	client "github.com/fipaan/ap2-uni/op-assign/order-service/internal/client/grpc"

	"github.com/google/uuid"
)

var ErrNoErrIdempotency    = errors.New("key duplicate")
var ErrInvalidCustomerID   = errors.New("customer id is empty")
var ErrInvalidItem         = errors.New("item name is empty")
var ErrInvalidAmount       = errors.New("amount must be > 0")
var ErrCancelNonExist      = errors.New("order does not exist")
var ErrNotPending          = errors.New("order is not pending")
var ErrPaymentNotAvailable = client.ErrPaymentNotAvailable

type OrderRepository interface {
	Create(ctx context.Context, o *domain.Order) error
	GetByID(ctx context.Context, id string) (*domain.Order, error)
	ListPayments(ctx context.Context, status string) (*[]paymentdomain.Payment, error)
	UpdateStatus(ctx context.Context, id string, status string) error
	GetByIdempotencyKey(ctx context.Context, key string) (*domain.Order, error)
}

type OrderUsecase struct {
	repo    OrderRepository
	payment client.PaymentClient
}

func NewOrderUsecase(r OrderRepository, p client.PaymentClient) *OrderUsecase {
	return &OrderUsecase{repo: r, payment: p}
}

func validateCreate(customerID, itemName string, amount int64) error {
	if customerID == "" { return ErrInvalidCustomerID }
	if itemName   == "" { return ErrInvalidItem       }
	if amount     <= 0  { return ErrInvalidAmount     }
	return nil
}

func (uc *OrderUsecase) getByIdempotencyKey(ctx context.Context, idempotencyKey string) (order *domain.Order) {
	if idempotencyKey == "" { return nil }
	existing, err := uc.repo.GetByIdempotencyKey(ctx, idempotencyKey)
	if err != nil { return nil }
	return existing
}

func (uc *OrderUsecase) processPayment(orderID string) {
	time.Sleep(5 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	order, err := uc.repo.GetByID(ctx, orderID)
	if err != nil || order.Status != "Authorized" {
		return
	}

	if err == nil && order.Status == "Authorized" {
		order.Status = "Paid"
	} else {
		order.Status = "Failed"
	}
	_ = uc.repo.UpdateStatus(ctx, order.ID, order.Status)
}

func (uc *OrderUsecase) Create(ctx context.Context, customerID, itemName string, amount int64, idempotencyKey string) (order *domain.Order, err error) {
	err = validateCreate(customerID, itemName, amount)
	if err != nil { return }

	order = uc.getByIdempotencyKey(ctx, idempotencyKey)
	if order != nil {
		err = ErrNoErrIdempotency
		return
	}

	order = &domain.Order{
		ID:             uuid.New().String(),
		CustomerID:     customerID,
		ItemName:       itemName,
		Amount:         amount,
		Status:         "Pending",
		CreatedAt:      time.Now().UTC().Truncate(time.Microsecond), // that's how psql stores it
		IdempotencyKey: idempotencyKey,
	}

	err = uc.repo.Create(ctx, order)
	if err != nil {
		order = uc.getByIdempotencyKey(ctx, idempotencyKey)
		if order != nil { err = ErrNoErrIdempotency }
		return
	}

	status, err := uc.payment.Pay(ctx, order.ID, order.Amount)
	if err != nil {
		order.Status = "Failed"
		err = ErrPaymentNotAvailable
	} else {
		go uc.processPayment(order.ID)
   	
		if status == "Authorized" {
			order.Status = "Authorized"
		} else {
			order.Status = "Failed"
		}
	}

	_ = uc.repo.UpdateStatus(ctx, order.ID, order.Status)

	return
}

func (uc *OrderUsecase) Get(ctx context.Context, id string) (*domain.Order, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *OrderUsecase) ListPayments(ctx context.Context, status string) (*[]paymentdomain.Payment, error) {
	payments, err := uc.payment.ListPayments(ctx, status)
	if err != nil {
		err = ErrPaymentNotAvailable
	}
	return payments, err
}

func (uc *OrderUsecase) Cancel(ctx context.Context, id string) error {
	order, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrCancelNonExist
		}
		return err
	}

	if order.Status == "Pending" ||
	  (order.Status == "Authorized" && time.Since(order.CreatedAt) < 5*time.Second) {
		return uc.repo.UpdateStatus(ctx, id, "Cancelled")
	}
	return ErrNotPending
}
