package provider

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"order_service/internal/models"
)

const baseURL = "https://api.yookassa.ru/v3"

type YooKassaProvider struct {
	shopID          string
	secretKey       string
	returnURL       string
	notificationURL string
	client          *http.Client
	log             *slog.Logger
}

func NewYooKassaProvider(shopID, secretKey, returnURL, notificationURL string, httpTimeout time.Duration, log *slog.Logger) *YooKassaProvider {
	return &YooKassaProvider{
		shopID:          shopID,
		secretKey:       secretKey,
		returnURL:       returnURL,
		notificationURL: notificationURL,
		client:          &http.Client{Timeout: httpTimeout},
		log:             log,
	}
}

func (p *YooKassaProvider) authHeader() string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", p.shopID, p.secretKey)))
}

type paymentRequest struct {
	Amount       amount              `json:"amount"`
	Capture      bool                `json:"capture"`
	Confirmation requestConfirmation `json:"confirmation"`
	Description  string              `json:"description"`
	Metadata     map[string]string   `json:"metadata,omitempty"`
}

type amount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

type requestConfirmation struct {
	Type      string `json:"type"`
	ReturnURL string `json:"return_url"`
}

type responseConfirmation struct {
	Type            string `json:"type"`
	ConfirmationURL string `json:"confirmation_url"`
}

type paymentResponse struct {
	ID           string               `json:"id"`
	Status       string               `json:"status"`
	Amount       amount               `json:"amount"`
	Confirmation responseConfirmation `json:"confirmation"`
	Paid         bool                 `json:"paid"`
}

func (p *YooKassaProvider) CreatePayment(ctx context.Context, amountVal int, currency, description string) (*models.PaymentProviderResponse, error) {
	const op = "provider.YooKassaProvider.CreatePayment"

	reqBody := paymentRequest{
		Amount: amount{
			Value:    fmt.Sprintf("%.2f", float64(amountVal)/100.0),
			Currency: currency,
		},
		Capture: true,
		Confirmation: requestConfirmation{
			Type:      "redirect",
			ReturnURL: p.returnURL,
		},
		Description: description,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("%s: marshal request: %w", op, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/payments", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("%s: create request: %w", op, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotence-Key", uuid.New().String())
	req.Header.Set("Authorization", p.authHeader())

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: send request: %w", op, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: read response body: %w", op, err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		p.log.Error("yookassa api error",
			slog.String("op", op),
			slog.Int("status_code", resp.StatusCode),
		)
		return nil, fmt.Errorf("%s: yookassa api error: status %d", op, resp.StatusCode)
	}

	var pr paymentResponse
	if err := json.Unmarshal(bodyBytes, &pr); err != nil {
		return nil, fmt.Errorf("%s: decode response: %w", op, err)
	}

	p.log.Info("yookassa payment created",
		slog.String("op", op),
		slog.String("payment_id", pr.ID),
		slog.String("status", pr.Status),
	)

	return &models.PaymentProviderResponse{
		ID:              pr.ID,
		Status:          pr.Status,
		ConfirmationURL: pr.Confirmation.ConfirmationURL,
	}, nil
}
