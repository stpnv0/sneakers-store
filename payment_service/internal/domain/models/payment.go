package models

import "time"

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusSucceeded PaymentStatus = "succeeded"
	PaymentStatusCanceled  PaymentStatus = "canceled"
	PaymentStatusWaiting   PaymentStatus = "waiting_for_capture"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusRefunded  PaymentStatus = "refunded"
)

type Payment struct {
	ID            string            `json:"id" db:"id"`
	OrderID       string            `json:"order_id" db:"order_id"`
	UserID        string            `json:"user_id" db:"user_id"`
	Amount        float64           `json:"amount" db:"amount"`
	Currency      string            `json:"currency" db:"currency"`
	Status        PaymentStatus     `json:"status" db:"status"`
	PaymentMethod string            `json:"payment_method" db:"payment_method"`
	CreatedAt     time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at" db:"updated_at"`
	PaymentURL    string            `json:"payment_url,omitempty" db:"payment_url"`
	ExternalID    string            `json:"external_id" db:"external_id"`
	Metadata      map[string]string `json:"metadata,omitempty" db:"metadata"`
	Description   string            `json:"description" db:"description"`
	ReturnURL     string            `json:"return_url" db:"return_url"`
}

type CreatePaymentRequest struct {
	OrderID     string            `json:"order_id" binding:"required"`
	Amount      float64           `json:"amount" binding:"required,gt=0"`
	Currency    string            `json:"currency" binding:"required,len=3"`
	Description string            `json:"description" binding:"required"`
	ReturnURL   string            `json:"return_url" binding:"required,url"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type CreatePaymentResponse struct {
	PaymentID       string `json:"payment_id"`
	ConfirmationURL string `json:"confirmation_url"`
}

type YooKassaWebhook struct {
	Event  string `json:"event"`
	Object struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	} `json:"object"`
}
