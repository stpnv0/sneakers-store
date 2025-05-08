package yookassa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type GetPaymentResponse struct {
	ID            string            `json:"id"`
	Status        string            `json:"status"`
	Amount        Amount            `json:"amount"`
	Description   string            `json:"description"`
	CreatedAt     time.Time         `json:"created_at"`
	CapturedAt    *time.Time        `json:"captured_at,omitempty"`
	Confirmation  Confirmation      `json:"confirmation,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	PaymentMethod struct {
		Type string `json:"type"`
	} `json:"payment_method"`
}

type Client struct {
	shopID      string
	secretKey   string
	httpClient  *http.Client
	apiEndpoint string
	logger      *zap.Logger
}

func NewClient(shopID, secretKey string, logger *zap.Logger) *Client {
	return &Client{
		shopID:      shopID,
		secretKey:   secretKey,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
		apiEndpoint: "https://api.yookassa.ru/v3",
		logger:      logger,
	}
}

func (c *Client) CreatePayment(ctx context.Context, req *CreatePaymentRequest, idempotencyKey string) (*CreatePaymentResponse, error) {
	url := fmt.Sprintf("%s/payments", c.apiEndpoint)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Idempotence-Key", idempotencyKey)
	httpReq.SetBasicAuth(c.shopID, c.secretKey)

	c.logger.Debug("Sending request to YooKassa",
		zap.String("url", url),
		zap.String("idempotency_key", idempotencyKey))

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			return nil, fmt.Errorf("YooKassa API error: %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("YooKassa API error: %s (code: %s)", errorResp.Description, errorResp.Code)
	}

	var paymentResp CreatePaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&paymentResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.Debug("Received response from YooKassa",
		zap.String("payment_id", paymentResp.ID),
		zap.String("status", paymentResp.Status))

	return &paymentResp, nil
}

// GetPayment получает информацию о платеже по его ID в ЮКассе
func (c *Client) GetPayment(ctx context.Context, paymentID string) (*GetPaymentResponse, error) {
	url := fmt.Sprintf("%s/payments/%s", c.apiEndpoint, paymentID)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.SetBasicAuth(c.shopID, c.secretKey)

	c.logger.Debug("Getting payment info from YooKassa",
		zap.String("payment_id", paymentID))

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			return nil, fmt.Errorf("YooKassa API error: %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("YooKassa API error: %s (code: %s)", errorResp.Description, errorResp.Code)
	}

	var paymentResp GetPaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&paymentResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.Debug("Received payment info from YooKassa",
		zap.String("payment_id", paymentResp.ID),
		zap.String("status", paymentResp.Status))

	return &paymentResp, nil
}

// CapturePayment подтверждает платеж (для двухстадийных платежей)
func (c *Client) CapturePayment(ctx context.Context, paymentID string, amount *Amount, idempotencyKey string) (*GetPaymentResponse, error) {
	url := fmt.Sprintf("%s/payments/%s/capture", c.apiEndpoint, paymentID)

	var reqBody []byte
	var err error

	if amount != nil {
		reqData := map[string]interface{}{
			"amount": amount,
		}
		reqBody, err = json.Marshal(reqData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
	} else {
		reqBody = []byte("{}")
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Idempotence-Key", idempotencyKey)
	httpReq.SetBasicAuth(c.shopID, c.secretKey)

	c.logger.Debug("Capturing payment in YooKassa",
		zap.String("payment_id", paymentID),
		zap.String("idempotency_key", idempotencyKey))

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			return nil, fmt.Errorf("YooKassa API error: %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("YooKassa API error: %s (code: %s)", errorResp.Description, errorResp.Code)
	}

	var paymentResp GetPaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&paymentResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.Debug("Payment captured successfully",
		zap.String("payment_id", paymentResp.ID),
		zap.String("status", paymentResp.Status))

	return &paymentResp, nil
}

// CancelPayment отменяет платеж
func (c *Client) CancelPayment(ctx context.Context, paymentID string, idempotencyKey string) (*GetPaymentResponse, error) {
	url := fmt.Sprintf("%s/payments/%s/cancel", c.apiEndpoint, paymentID)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer([]byte("{}")))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Idempotence-Key", idempotencyKey)
	httpReq.SetBasicAuth(c.shopID, c.secretKey)

	c.logger.Debug("Cancelling payment in YooKassa",
		zap.String("payment_id", paymentID),
		zap.String("idempotency_key", idempotencyKey))

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			return nil, fmt.Errorf("YooKassa API error: %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("YooKassa API error: %s (code: %s)", errorResp.Description, errorResp.Code)
	}

	var paymentResp GetPaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&paymentResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.Debug("Payment cancelled successfully",
		zap.String("payment_id", paymentResp.ID),
		zap.String("status", paymentResp.Status))

	return &paymentResp, nil
}
