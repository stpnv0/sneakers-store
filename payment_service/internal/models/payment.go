package models

import "time"

const (
	PaymentStatusPending   = "pending"
	PaymentStatusSucceeded = "succeeded"
	PaymentStatusCanceled  = "canceled"
)

type Payment struct {
	ID                int       `db:"id"`
	OrderID           int       `db:"order_id"`
	YooKassaPaymentID string    `db:"yookassa_payment_id"`
	Amount            int       `db:"amount"`
	Currency          string    `db:"currency"`
	Status            string    `db:"status"`
	ConfirmationURL   string    `db:"confirmation_url"`
	CreatedAt         time.Time `db:"created_at"`
	UpdatedAt         time.Time `db:"updated_at"`
}
