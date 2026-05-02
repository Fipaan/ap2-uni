package repo

import (
	"context"
	"database/sql"
	"github.com/Fipaan/ap2-uni/payment-service/internal/domain"
)

type PaymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(ctx context.Context, p *domain.Payment) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO payments (id, order_id, transaction_id, amount, status, customer_email)
		 VALUES ($1,$2,$3,$4,$5,$6)`,
		p.ID, p.OrderID, p.TransactionID, p.Amount, p.Status, p.CustomerEmail,
	)
	return err
}

func (r *PaymentRepository) GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, order_id, transaction_id, amount, status, customer_email FROM payments WHERE order_id=$1`,
		orderID,
	)

	var p domain.Payment
	err := row.Scan(&p.ID, &p.OrderID, &p.TransactionID, &p.Amount, &p.Status, &p.CustomerEmail)
	return &p, err
}

func (r *PaymentRepository) GetAll(ctx context.Context) (*[]domain.Payment, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, order_id, transaction_id, amount, status, customer_email FROM payments`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []domain.Payment

	for rows.Next() {
		var p domain.Payment
		if err := rows.Scan(&p.ID, &p.OrderID, &p.TransactionID, &p.Amount, &p.Status, &p.CustomerEmail); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &payments, nil
}

func (r *PaymentRepository) GetAllByStatus(ctx context.Context, status string) (*[]domain.Payment, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, order_id, transaction_id, amount, status, customer_email FROM payments WHERE status=$1`,
		status,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []domain.Payment

	for rows.Next() {
		var p domain.Payment
		if err := rows.Scan(&p.ID, &p.OrderID, &p.TransactionID, &p.Amount, &p.Status, &p.CustomerEmail); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &payments, nil
}
