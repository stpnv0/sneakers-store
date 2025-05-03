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

type FavClient struct {
	httpClient *http.Client
	baseURL    string
}

func NewFavClient(baseURL string, timeout time.Duration) *FavClient {
	return &FavClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *FavClient) ProxyRequest(ctx context.Context, method, path string, userSSOID int, body io.Reader) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, path)

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Получаем токен из контекста запроса
	authToken, ok := ctx.Value("auth_token").(string)
	if !ok || authToken == "" {
		// Если токен не найден, используем заголовок с ID пользователя как запасной вариант
		req.Header.Set("X-User-ID", fmt.Sprintf("%d", userSSOID))
	} else {
		// Если токен найден, используем его
		req.Header.Set("Authorization", authToken)
	}
	
	req.Header.Set("Content-Type", "application/json")

	// Выполняем запрос
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request to favourites service: %w", err)
	}

	return resp, nil
}

func (c *FavClient) AddToFavourite(ctx context.Context, userSSOID, sneakerID int) error {
	reqBody := struct {
		SneakerID int `json:"sneaker_id"`
	}{
		SneakerID: sneakerID,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("error marshaling request body: %w", err)
	}

	resp, err := c.ProxyRequest(ctx, "POST", "/api/v1/favourites", userSSOID, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error adding to favourites, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (c *FavClient) GetAllFavourites(ctx context.Context, userSSOID int) ([]byte, error) {
	resp, err := c.ProxyRequest(ctx, "GET", "/api/v1/favourites", userSSOID, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error getting favourites, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	return io.ReadAll(resp.Body)
}

func (c *FavClient) DeleteFavourite(ctx context.Context, userSSOID int, itemID string) error {
	path := fmt.Sprintf("/api/v1/favourites/%s", itemID)
	resp, err := c.ProxyRequest(ctx, "DELETE", path, userSSOID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error deleting from favourites, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}