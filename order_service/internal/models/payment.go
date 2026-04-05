package models

import "time"

const (
	PaymentStatusPending   = "pending"
	PaymentStatusSucceeded = "succeeded"
	PaymentStatusCanceled  = "canceled"
)

var validPaymentTransitions = map[string][]string{
	PaymentStatusPending:   {PaymentStatusSucceeded, PaymentStatusCanceled},
	PaymentStatusSucceeded: {},
	PaymentStatusCanceled:  {},
}

func ValidPaymentTransition(from, to string) bool {
	allowed, ok := validPaymentTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

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

type PaymentProviderResponse struct {
	ID              string
	Status          string
	ConfirmationURL string
}
