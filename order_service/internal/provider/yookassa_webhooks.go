package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

type webhookRequest struct {
	Event string `json:"event"`
	URL   string `json:"url"`
}

type webhookItem struct {
	ID    string `json:"id"`
	Event string `json:"event"`
	URL   string `json:"url"`
}

type webhookListResponse struct {
	Type  string        `json:"type"`
	Items []webhookItem `json:"items"`
}

// RegisterWebhooks регистрирует URL вебхуков в ЮKassa, Удаляет устаревшие вебхуки,
func (p *YooKassaProvider) RegisterWebhooks(ctx context.Context) {
	const op = "provider.YooKassaProvider.RegisterWebhooks"

	if p.notificationURL == "" {
		p.log.Warn("notification_url is not set — YooKassa webhooks will NOT be registered; payment status updates will not work",
			slog.String("op", op),
		)
		return
	}

	p.log.Info("registering YooKassa webhooks", slog.String("op", op), slog.String("url", p.notificationURL))

	existing, err := p.listWebhooks(ctx)
	if err != nil {
		p.log.Error("failed to list webhooks, skipping registration", slog.String("op", op), slog.String("error", err.Error()))
		return
	}

	desiredEvents := map[string]bool{
		"payment.succeeded": false,
		"payment.canceled":  false,
	}

	for _, wh := range existing {
		if wh.URL == p.notificationURL {
			if _, ok := desiredEvents[wh.Event]; ok {
				desiredEvents[wh.Event] = true
				p.log.Info("webhook already registered", slog.String("event", wh.Event), slog.String("id", wh.ID))
			}
		} else {
			if _, ok := desiredEvents[wh.Event]; ok {
				p.log.Info("deleting stale webhook", slog.String("event", wh.Event), slog.String("old_url", wh.URL), slog.String("id", wh.ID))
				if delErr := p.deleteWebhook(ctx, wh.ID); delErr != nil {
					p.log.Error("failed to delete stale webhook", slog.String("id", wh.ID), slog.String("error", delErr.Error()))
				}
			}
		}
	}

	for event, registered := range desiredEvents {
		if registered {
			continue
		}
		if err := p.createWebhook(ctx, event, p.notificationURL); err != nil {
			p.log.Error("failed to register webhook", slog.String("event", event), slog.String("error", err.Error()))
		} else {
			p.log.Info("webhook registered successfully", slog.String("event", event), slog.String("url", p.notificationURL))
		}
	}
}

func (p *YooKassaProvider) listWebhooks(ctx context.Context) ([]webhookItem, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/webhooks", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", p.authHeader())

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var result webhookListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return result.Items, nil
}

func (p *YooKassaProvider) createWebhook(ctx context.Context, event, url string) error {
	body, err := json.Marshal(webhookRequest{Event: event, URL: url})
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/webhooks", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", p.authHeader())
	req.Header.Set("Idempotence-Key", uuid.New().String())

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("yookassa error: status %d, body: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

func (p *YooKassaProvider) deleteWebhook(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, baseURL+"/webhooks/"+id, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", p.authHeader())

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("yookassa error: status %d", resp.StatusCode)
	}
	return nil
}
