package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"payment/internal/domain/models"
	services "payment/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type WebhookHandler struct {
	paymentService *services.PaymentService
	logger         *zap.Logger
}

func NewWebhookHandler(paymentService *services.PaymentService, logger *zap.Logger) *WebhookHandler {
	return &WebhookHandler{
		paymentService: paymentService,
		logger:         logger,
	}
}

func (h *WebhookHandler) HandleYooKassaWebhook(c *gin.Context) {
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("ERROR: Failed to read request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	var webhook models.YooKassaWebhook
	if err := json.Unmarshal(body, &webhook); err != nil {
		h.logger.Error("ERROR: Failed to parse JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	h.logger.Info("INFO: Received webhook from YooKassa",
		zap.String("event", webhook.Event),
		zap.String("payment_id", webhook.Object.ID),
		zap.String("status", webhook.Object.Status))

	var paymentStatus models.PaymentStatus

	switch webhook.Event {
	case "payment.succeeded":
		paymentStatus = models.PaymentStatusSucceeded
	case "payment.canceled":
		paymentStatus = models.PaymentStatusCanceled
	case "payment.waiting_for_capture":
		paymentStatus = models.PaymentStatusWaiting
	default:
		h.logger.Warn("Unknown event type", zap.String("event", webhook.Event))
		c.JSON(http.StatusOK, gin.H{"status": "ok"}) // Всегда возвращаем 200 OK
		return
	}

	if err := h.paymentService.UpdatePaymentStatusByExternalID(c.Request.Context(), webhook.Object.ID, paymentStatus); err != nil {
		h.logger.Error("Failed to update payment status",
			zap.Error(err),
			zap.String("external_id", webhook.Object.ID))
		// Всё равно возвращаем 200 OK, чтобы ЮKassa не повторяла запрос
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
