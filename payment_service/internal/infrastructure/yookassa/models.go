package yookassa

import "time"

type Amount struct {
	Value    float64 `json:"value"`
	Currency string  `json:"currency"`
}

type Confirmation struct {
	Type            string `json:"type"`
	ReturnURL       string `json:"return_url,omitempty"`
	ConfirmationURL string `json:"confirmation_url,omitempty"`
}

type CreatePaymentRequest struct {
	Amount       Amount            `json:"amount"`
	Capture      bool              `json:"capture"`
	Confirmation Confirmation      `json:"confirmation"`
	Description  string            `json:"description,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

type CreatePaymentResponse struct {
	ID           string            `json:"id"`
	Status       string            `json:"status"`
	Amount       Amount            `json:"amount"`
	Description  string            `json:"description"`
	Confirmation Confirmation      `json:"confirmation"`
	CreatedAt    time.Time         `json:"created_at"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

type ErrorResponse struct {
	Type        string `json:"type"`
	ID          string `json:"id"`
	Code        string `json:"code"`
	Description string `json:"description"`
}
