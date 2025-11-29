package api

import (
	"log/slog"
	"net/http"
	"payment_service/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *service.PaymentService
	log     *slog.Logger
}

func NewHandler(service *service.PaymentService, log *slog.Logger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

type YooKassaWebhook struct {
	Type   string `json:"type"`
	Event  string `json:"event"`
	Object struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Amount struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"amount"`
	} `json:"object"`
}

func (h *Handler) HandleWebhook(c *gin.Context) {
	var webhook YooKassaWebhook
	if err := c.ShouldBindJSON(&webhook); err != nil {
		h.log.Error("failed to bind webhook", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	h.log.Info("received webhook",
		slog.String("event", webhook.Event),
		slog.String("payment_id", webhook.Object.ID),
		slog.String("status", webhook.Object.Status))

	if webhook.Event == "payment.succeeded" || webhook.Event == "payment.canceled" {
		err := h.service.ProcessWebhook(c.Request.Context(), webhook.Object.ID, webhook.Object.Status)
		if err != nil {
			h.log.Error("failed to process webhook", slog.String("error", err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
	}

	c.Status(http.StatusOK)
}

// ManualStatusUpdate - temporary endpoint for testing (should be removed in production)
func (h *Handler) ManualStatusUpdate(c *gin.Context) {
	var req struct {
		PaymentID string `json:"payment_id" binding:"required"`
		Status    string `json:"status" binding:"required"` // "succeeded" or "canceled"
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	h.log.Info("manual status update",
		slog.String("payment_id", req.PaymentID),
		slog.String("status", req.Status))

	err := h.service.ProcessWebhook(c.Request.Context(), req.PaymentID, req.Status)
	if err != nil {
		h.log.Error("failed to process status update", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "status updated successfully"})
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	router.POST("/webhook/yookassa", h.HandleWebhook)
	router.POST("/api/manual-status-update", h.ManualStatusUpdate) // Temporary for testing
}
