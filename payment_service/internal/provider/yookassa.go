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
	"payment_service/internal/config"

	"github.com/google/uuid"
)

const (
	BaseURL = "https://api.yookassa.ru/v3"
)

type YooKassaProvider struct {
	shopID    string
	secretKey string
	returnURL string
	client    *http.Client
	log       *slog.Logger
}

type PaymentRequest struct {
	Amount       Amount              `json:"amount"`
	Capture      bool                `json:"capture"`
	Confirmation RequestConfirmation `json:"confirmation"`
	Description  string              `json:"description"`
}

type Amount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

type RequestConfirmation struct {
	Type      string `json:"type"`
	ReturnURL string `json:"return_url"`
}

type ResponseConfirmation struct {
	Type            string `json:"type"`
	ConfirmationURL string `json:"confirmation_url"`
}

type PaymentResponse struct {
	ID           string               `json:"id"`
	Status       string               `json:"status"`
	Amount       Amount               `json:"amount"`
	Confirmation ResponseConfirmation `json:"confirmation"`
	Paid         bool                 `json:"paid"`
}

func NewYooKassaProvider(cfg *config.Config, log *slog.Logger) *YooKassaProvider {
	return &YooKassaProvider{
		shopID:    cfg.YooKassa.ShopID,
		secretKey: cfg.YooKassa.SecretKey,
		returnURL: cfg.YooKassa.ReturnURL,
		client: &http.Client{
			Timeout: cfg.HTTP.Timeout,
		},
		log: log,
	}
}

func (p *YooKassaProvider) CreatePayment(ctx context.Context, amount int, currency, description string) (*PaymentResponse, error) {
	idempotenceKey := uuid.New().String()

	reqBody := PaymentRequest{
		Amount: Amount{
			Value:    fmt.Sprintf("%.2f", float64(amount)),
			Currency: currency,
		},
		Capture: true,
		Confirmation: RequestConfirmation{
			Type:      "redirect",
			ReturnURL: p.returnURL,
		},
		Description: description,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", BaseURL+"/payments", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotence-Key", idempotenceKey)

	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", p.shopID, p.secretKey)))
	req.Header.Set("Authorization", "Basic "+auth)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("yookassa api error: status %d", resp.StatusCode)
	}

	// Debug: Log raw response body
	bodyBytes, _ := io.ReadAll(resp.Body)
	p.log.Info("yookassa response", slog.String("body", string(bodyBytes)))

	// Restore body for decoder
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var paymentResp PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&paymentResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &paymentResp, nil
}
