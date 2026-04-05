package api

import (
	"context"
	"crypto/subtle"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type WebhookService interface {
	ProcessWebhook(ctx context.Context, yookassaID, status string) error
}

type WebhookHandler struct {
	svc         WebhookService
	log         *slog.Logger
	validate    *validator.Validate
	adminAPIKey string
}

func NewWebhookHandler(svc WebhookService, log *slog.Logger, adminAPIKey string) *WebhookHandler {
	return &WebhookHandler{
		svc:         svc,
		log:         log,
		validate:    validator.New(),
		adminAPIKey: adminAPIKey,
	}
}

func (h *WebhookHandler) RegisterRoutes(router *gin.Engine) {
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	router.POST("/webhook/yookassa", h.HandleWebhook)
	router.POST("/api/manual-status-update", h.ManualStatusUpdate)
}

type yooKassaWebhook struct {
	Type   string `json:"type"`
	Event  string `json:"event"  validate:"required"`
	Object struct {
		ID     string `json:"id"     validate:"required"`
		Status string `json:"status" validate:"required"`
		Amount struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"amount"`
	} `json:"object"`
}

type manualUpdateRequest struct {
	PaymentID string `json:"payment_id" binding:"required" validate:"required"`
	Status    string `json:"status"     binding:"required" validate:"required,oneof=succeeded canceled"`
}

func (h *WebhookHandler) HandleWebhook(c *gin.Context) {
	var webhook yooKassaWebhook
	if err := c.ShouldBindJSON(&webhook); err != nil {
		h.log.Error("failed to bind webhook", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.validate.Struct(webhook); err != nil {
		h.log.Warn("webhook validation failed", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed"})
		return
	}

	h.log.Info("received webhook",
		slog.String("event", webhook.Event),
		slog.String("payment_id", webhook.Object.ID),
		slog.String("status", webhook.Object.Status),
	)

	if webhook.Event == "payment.succeeded" || webhook.Event == "payment.canceled" {
		if err := h.svc.ProcessWebhook(c.Request.Context(), webhook.Object.ID, webhook.Object.Status); err != nil {
			h.log.Error("failed to process webhook", slog.String("error", err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
	}

	c.Status(http.StatusOK)
}

func (h *WebhookHandler) ManualStatusUpdate(c *gin.Context) {
	if h.adminAPIKey == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "endpoint disabled"})
		return
	}

	providedKey := c.GetHeader("X-Admin-API-Key")
	if subtle.ConstantTimeCompare([]byte(providedKey), []byte(h.adminAPIKey)) != 1 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or missing admin API key"})
		return
	}

	var req manualUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.log.Info("manual status update",
		slog.String("payment_id", req.PaymentID),
		slog.String("status", req.Status),
	)

	if err := h.svc.ProcessWebhook(c.Request.Context(), req.PaymentID, req.Status); err != nil {
		h.log.Error("failed to process status update", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "status updated successfully"})
}
