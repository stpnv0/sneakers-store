package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"order_service/internal/models"
)

type PaymentRepository struct {
	pool *pgxpool.Pool
}

func NewPaymentRepository(pool *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{pool: pool}
}

func (r *PaymentRepository) Create(ctx context.Context, payment *models.Payment) error {
	const op = "repository.PaymentRepository.Create"

	now := time.Now()
	err := r.pool.QueryRow(ctx,
		`INSERT INTO payments
		   (order_id, yookassa_payment_id, amount, currency, status, confirmation_url, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id`,
		payment.OrderID, payment.YooKassaPaymentID, payment.Amount,
		payment.Currency, payment.Status, payment.ConfirmationURL, now, now,
	).Scan(&payment.ID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *PaymentRepository) UpdateStatusAndGet(ctx context.Context, yookassaID, newStatus string) (*models.Payment, error) {
	const op = "repository.PaymentRepository.UpdateStatusAndGet"

	var p models.Payment
	err := r.pool.QueryRow(ctx,
		`UPDATE payments SET status = $1, updated_at = $2
		 WHERE yookassa_payment_id = $3 AND status != $1
		 RETURNING id, order_id, yookassa_payment_id, status, amount, currency,
		           confirmation_url, created_at, updated_at`,
		newStatus, time.Now(), yookassaID,
	).Scan(
		&p.ID, &p.OrderID, &p.YooKassaPaymentID, &p.Status, &p.Amount,
		&p.Currency, &p.ConfirmationURL, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // idempotent: already processed
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &p, nil
}

func (r *PaymentRepository) GetByYooKassaID(ctx context.Context, yookassaID string) (*models.Payment, error) {
	const op = "repository.PaymentRepository.GetByYooKassaID"

	var p models.Payment
	err := r.pool.QueryRow(ctx,
		`SELECT id, order_id, yookassa_payment_id, amount, currency, status,
		        confirmation_url, created_at, updated_at
		 FROM payments WHERE yookassa_payment_id = $1`, yookassaID,
	).Scan(
		&p.ID, &p.OrderID, &p.YooKassaPaymentID, &p.Amount,
		&p.Currency, &p.Status, &p.ConfirmationURL, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &p, nil
}
