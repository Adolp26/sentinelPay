package payment

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("payment not found")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, p *Payment) error {
	const query = `
		INSERT INTO payments (id, merchant_id, customer_id, amount, currency, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		p.ID, p.MerchantID, p.CustomerID, p.Amount, p.Currency, p.Status,
	).Scan(&p.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert payment: %w", err)
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (Payment, error) {
	const query = `
		SELECT id, merchant_id, customer_id, amount, currency, status, created_at
		FROM payments
		WHERE id = $1`

	if uuid.Validate(id) != nil {
		return Payment{}, ErrNotFound
	}

	var p Payment
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.MerchantID, &p.CustomerID, &p.Amount, &p.Currency, &p.Status, &p.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Payment{}, ErrNotFound
	}
	if err != nil {
		return Payment{}, fmt.Errorf("get payment: %w", err)
	}
	return p, nil
}
