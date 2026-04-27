package repo

import (
	"context"
	"database/sql"

	"github.com/Fipaan/ap2-uni/order-service/internal/domain"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, o *domain.Order) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO orders (id, customer_id, item_name, amount, status, created_at, idempotency_key)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		o.ID, o.CustomerID, o.ItemName, o.Amount, o.Status, o.CreatedAt, o.IdempotencyKey,
	)
	return err
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, customer_id, item_name, amount, status, created_at, idempotency_key
		 FROM orders WHERE id=$1`,
		id,
	)

	var o domain.Order
	err := row.Scan(&o.ID, &o.CustomerID, &o.ItemName, &o.Amount, &o.Status, &o.CreatedAt, &o.IdempotencyKey)
	return &o, err
}

func (r *OrderRepository) GetByIdempotencyKey(ctx context.Context, key string) (*domain.Order, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, customer_id, item_name, amount, status, created_at, idempotency_key
		 FROM orders WHERE idempotency_key=$1`,
		key,
	)

	var o domain.Order
	err := row.Scan(&o.ID, &o.CustomerID, &o.ItemName, &o.Amount, &o.Status, &o.CreatedAt, &o.IdempotencyKey)
	return &o, err
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE orders SET status=$1 WHERE id=$2`,
		status, id,
	)
	return err
}
