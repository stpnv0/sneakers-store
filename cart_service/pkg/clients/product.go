package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ProductInfo struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	ImageUrl    string  `json:"imageUrl"`
}

type ProductClient struct {
	httpClient *http.Client
	baseURL    string
}

func NewProductClient(baseURL string) *ProductClient {
	return &ProductClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *ProductClient) GetProduct(ctx context.Context, SneakerID int64) (*ProductInfo, error) {
	url := fmt.Sprintf("%s/items/%d", c.baseURL, SneakerID)

	//реквест с контекстом
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("ERROR: Can't create request: %w", err)
	}

	//клиент.Do
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ERROR: Can't do request: %w", err)
	}
	defer resp.Body.Close()

	//проверка статус кода
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected Error: %s", resp.StatusCode)
	}

	//декодирование
	var product ProductInfo
	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		return nil, fmt.Errorf("ERROR: Can't decode data of product: %w", err)
	}

	return &product, nil
}

func (c *ProductClient) GetProductsByIDs(ctx context.Context, IDs []int64) ([]*ProductInfo, error) {
	url := fmt.Sprintf("%s/items/batch", c.baseURL)

	reqBody := struct {
		IDs []int64 `json:"ids"`
	}{
		IDs: IDs,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("ERROR: Can't marshal ids: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("ERROR: Can't create batch request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ERROR: Can't do batch request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected Error on batch: %s", resp.Status)
	}

	var sneakers []struct {
		ID          int64   `json:"id"`
		Title       string  `json:"title"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		ImageUrl    string  `json:"imageUrl"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&sneakers); err != nil {
		return nil, fmt.Errorf("ERROR: can't decode batch response: %w", err)
	}

	products := make([]*ProductInfo, 0, len(sneakers))
	for _, s := range sneakers {
		products = append(products, &ProductInfo{
			ID:          s.ID,
			Title:       s.Title,
			Description: s.Description,
			Price:       s.Price,
			ImageUrl:    s.ImageUrl,
		})
	}

	return products, err
}

func (c *ProductClient) CheckProductAvailability(ctx context.Context, productID int) (bool, error) {
	product, err := c.GetProduct(ctx, int64(productID))
	if err != nil {
		return false, err
	}

	return product != nil, nil
}
