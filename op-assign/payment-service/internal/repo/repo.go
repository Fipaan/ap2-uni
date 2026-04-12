package repo

import (
	"context"
	"database/sql"
	"github.com/fipaan/ap2-uni/op-assign/payment-service/internal/domain"
)

type PaymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(ctx context.Context, p *domain.Payment) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO payments (id, order_id, transaction_id, amount, status)
		 VALUES ($1,$2,$3,$4,$5)`,
		p.ID, p.OrderID, p.TransactionID, p.Amount, p.Status,
	)
	return err
}

func (r *PaymentRepository) GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, order_id, transaction_id, amount, status FROM payments WHERE order_id=$1`,
		orderID,
	)

	var p domain.Payment
	err := row.Scan(&p.ID, &p.OrderID, &p.TransactionID, &p.Amount, &p.Status)
	return &p, err
}
