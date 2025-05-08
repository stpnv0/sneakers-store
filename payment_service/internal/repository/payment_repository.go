// internal/repository/postgres_repository.go
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"payment/internal/domain/models"
	"time"

	"github.com/jmoiron/sqlx"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment *models.Payment, idempotencyKey string) error
	Update(ctx context.Context, payment *models.Payment) error
	GetByID(ctx context.Context, id string) (*models.Payment, error)
	GetByOrderID(ctx context.Context, orderID string) (*models.Payment, error)
	GetByExternalID(ctx context.Context, externalID string) (*models.Payment, error)
	GetByIdempotencyKey(ctx context.Context, key string) (*models.Payment, error)
	GetPendingPayments(ctx context.Context, olderThan time.Time) ([]*models.Payment, error)
}

type PostgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, payment *models.Payment, idempotencyKey string) error {
	metadataJSON, err := json.Marshal(payment.Metadata)
	if err != nil {
		return err
	}

	query := `
        INSERT INTO payments (
            id, order_id, user_id, amount, currency, status, payment_method,
            created_at, updated_at, payment_url, external_id, metadata, description, return_url, idempotency_key
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
        )
    `

	_, err = r.db.ExecContext(
		ctx,
		query,
		payment.ID,
		payment.OrderID,
		payment.UserID,
		payment.Amount,
		payment.Currency,
		payment.Status,
		payment.PaymentMethod,
		payment.CreatedAt,
		payment.UpdatedAt,
		payment.PaymentURL,
		payment.ExternalID,
		metadataJSON,
		payment.Description,
		payment.ReturnURL,
		idempotencyKey,
	)

	return err
}

func (r *PostgresRepository) Update(ctx context.Context, payment *models.Payment) error {
	metadataJSON, err := json.Marshal(payment.Metadata)
	if err != nil {
		return err
	}

	query := `
        UPDATE payments
        SET 
            status = $1,
            payment_method = $2,
            updated_at = $3,
            payment_url = $4,
            external_id = $5,
            metadata = $6
        WHERE id = $7
    `

	_, err = r.db.ExecContext(
		ctx,
		query,
		payment.Status,
		payment.PaymentMethod,
		payment.UpdatedAt,
		payment.PaymentURL,
		payment.ExternalID,
		metadataJSON,
		payment.ID,
	)

	return err
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*models.Payment, error) {
	query := `
        SELECT 
            id, order_id, user_id, amount, currency, status, payment_method,
            created_at, updated_at, payment_url, external_id, metadata, description, return_url
        FROM payments
        WHERE id = $1
    `

	var payment models.Payment
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.UserID,
		&payment.Amount,
		&payment.Currency,
		&payment.Status,
		&payment.PaymentMethod,
		&payment.CreatedAt,
		&payment.UpdatedAt,
		&payment.PaymentURL,
		&payment.ExternalID,
		&metadataJSON,
		&payment.Description,
		&payment.ReturnURL,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &payment.Metadata); err != nil {
			return nil, err
		}
	}

	return &payment, nil
}

func (r *PostgresRepository) GetByOrderID(ctx context.Context, orderID string) (*models.Payment, error) {
	query := `
        SELECT 
            id, order_id, user_id, amount, currency, status, payment_method,
            created_at, updated_at, payment_url, external_id, metadata, description, return_url
        FROM payments
        WHERE order_id = $1
        ORDER BY created_at DESC
        LIMIT 1
    `

	var payment models.Payment
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.UserID,
		&payment.Amount,
		&payment.Currency,
		&payment.Status,
		&payment.PaymentMethod,
		&payment.CreatedAt,
		&payment.UpdatedAt,
		&payment.PaymentURL,
		&payment.ExternalID,
		&metadataJSON,
		&payment.Description,
		&payment.ReturnURL,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &payment.Metadata); err != nil {
			return nil, err
		}
	}

	return &payment, nil
}

func (r *PostgresRepository) GetByExternalID(ctx context.Context, externalID string) (*models.Payment, error) {
	query := `
        SELECT 
            id, order_id, user_id, amount, currency, status, payment_method,
            created_at, updated_at, payment_url, external_id, metadata, description, return_url
        FROM payments
        WHERE external_id = $1
    `

	var payment models.Payment
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, externalID).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.UserID,
		&payment.Amount,
		&payment.Currency,
		&payment.Status,
		&payment.PaymentMethod,
		&payment.CreatedAt,
		&payment.UpdatedAt,
		&payment.PaymentURL,
		&payment.ExternalID,
		&metadataJSON,
		&payment.Description,
		&payment.ReturnURL,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &payment.Metadata); err != nil {
			return nil, err
		}
	}

	return &payment, nil
}

func (r *PostgresRepository) GetByIdempotencyKey(ctx context.Context, key string) (*models.Payment, error) {
	query := `
        SELECT 
            id, order_id, user_id, amount, currency, status, payment_method,
            created_at, updated_at, payment_url, external_id, metadata, description, return_url
        FROM payments
        WHERE idempotency_key = $1
    `

	var payment models.Payment
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, key).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.UserID,
		&payment.Amount,
		&payment.Currency,
		&payment.Status,
		&payment.PaymentMethod,
		&payment.CreatedAt,
		&payment.UpdatedAt,
		&payment.PaymentURL,
		&payment.ExternalID,
		&metadataJSON,
		&payment.Description,
		&payment.ReturnURL,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &payment.Metadata); err != nil {
			return nil, err
		}
	}

	return &payment, nil
}

func (r *PostgresRepository) GetPendingPayments(ctx context.Context, olderThan time.Time) ([]*models.Payment, error) {
	query := `
        SELECT 
            id, order_id, user_id, amount, currency, status, payment_method,
            created_at, updated_at, payment_url, external_id, metadata, description, return_url
        FROM payments
        WHERE (status = $1 OR status = $2) AND created_at < $3
    `

	rows, err := r.db.QueryContext(
		ctx,
		query,
		models.PaymentStatusPending,
		models.PaymentStatusWaiting,
		olderThan,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*models.Payment

	for rows.Next() {
		var payment models.Payment
		var metadataJSON []byte

		err := rows.Scan(
			&payment.ID,
			&payment.OrderID,
			&payment.UserID,
			&payment.Amount,
			&payment.Currency,
			&payment.Status,
			&payment.PaymentMethod,
			&payment.CreatedAt,
			&payment.UpdatedAt,
			&payment.PaymentURL,
			&payment.ExternalID,
			&metadataJSON,
			&payment.Description,
			&payment.ReturnURL,
		)

		if err != nil {
			return nil, err
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &payment.Metadata); err != nil {
				return nil, err
			}
		}

		payments = append(payments, &payment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return payments, nil
}
