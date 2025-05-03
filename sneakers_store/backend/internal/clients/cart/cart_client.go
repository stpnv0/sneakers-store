package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type CartClient struct {
	httpClient *http.Client
	baseURL    string
}

func NewCartClient(baseURL string, timeout time.Duration) *CartClient {
	return &CartClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// ProxyRequest проксирует запрос к микросервису корзины
func (c *CartClient) ProxyRequest(ctx context.Context, method, path string, userSSOID int, body io.Reader) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, path)

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Добавляем ID пользователя в заголовок
	req.Header.Set("X-User-ID", fmt.Sprintf("%d", userSSOID))
	req.Header.Set("Content-Type", "application/json")

	// Выполняем запрос
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request to cart service: %w", err)
	}

	return resp, nil
}

// AddToCart добавляет товар в корзину
func (c *CartClient) AddToCart(ctx context.Context, userSSOID, sneakerID, quantity int) error {
	reqBody := struct {
		SneakerID int `json:"sneaker_id"`
		Quantity  int `json:"quantity"`
	}{
		SneakerID: sneakerID,
		Quantity:  quantity,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("error marshaling request body: %w", err)
	}

	resp, err := c.ProxyRequest(ctx, "POST", "/api/v1/cart", userSSOID, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error adding to cart, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// GetCart получает содержимое корзины пользователя
func (c *CartClient) GetCart(ctx context.Context, userSSOID int) ([]byte, error) {
	resp, err := c.ProxyRequest(ctx, "GET", "/api/v1/cart", userSSOID, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error getting cart, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	return io.ReadAll(resp.Body)
}

// UpdateCartItemQuantity обновляет количество товара в корзине
func (c *CartClient) UpdateCartItemQuantity(ctx context.Context, userSSOID int, itemID string, quantity int) error {
	reqBody := struct {
		Quantity int `json:"quantity"`
	}{
		Quantity: quantity,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("error marshaling request body: %w", err)
	}

	path := fmt.Sprintf("/api/v1/cart/%s", itemID)
	resp, err := c.ProxyRequest(ctx, "PUT", path, userSSOID, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error updating cart item, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// DeleteFromCart удаляет товар из корзины
func (c *CartClient) DeleteFromCart(ctx context.Context, userSSOID int, itemID string) error {
	path := fmt.Sprintf("/api/v1/cart/%s", itemID)
	resp, err := c.ProxyRequest(ctx, "DELETE", path, userSSOID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error deleting from cart, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
