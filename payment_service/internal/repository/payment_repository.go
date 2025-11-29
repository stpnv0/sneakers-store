package repository

import (
	"context"
	"fmt"
	"payment_service/internal/models"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentRepository struct {
	pool *pgxpool.Pool
}

func NewPaymentRepository(pool *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{pool: pool}
}

func (r *PaymentRepository) Create(ctx context.Context, payment *models.Payment) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO payments (order_id, yookassa_payment_id, amount, currency, status, confirmation_url, created_at, updated_at) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		payment.OrderID, payment.YooKassaPaymentID, payment.Amount, payment.Currency, payment.Status, payment.ConfirmationURL, time.Now(), time.Now(),
	).Scan(&payment.ID)
	if err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}
	return nil
}

func (r *PaymentRepository) UpdateStatus(ctx context.Context, yookassaID string, status string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE payments SET status = $1, updated_at = $2 WHERE yookassa_payment_id = $3`,
		status, time.Now(), yookassaID,
	)
	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}
	return nil
}

func (r *PaymentRepository) GetByYooKassaID(ctx context.Context, yookassaID string) (*models.Payment, error) {
	var payment models.Payment
	err := r.pool.QueryRow(ctx,
		`SELECT id, order_id, yookassa_payment_id, amount, currency, status, confirmation_url, created_at, updated_at 
		 FROM payments WHERE yookassa_payment_id = $1`,
		yookassaID,
	).Scan(&payment.ID, &payment.OrderID, &payment.YooKassaPaymentID, &payment.Amount, &payment.Currency, &payment.Status, &payment.ConfirmationURL, &payment.CreatedAt, &payment.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}
	return &payment, nil
}
