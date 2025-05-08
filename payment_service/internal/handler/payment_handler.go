package handlers

import (
	"net/http"
	"payment/internal/domain/models"
	services "payment/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PaymentHandler struct {
	paymentService *services.PaymentService
	logger         *zap.Logger
}

func NewPaymentHandler(paymentService *services.PaymentService, logger *zap.Logger) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
		logger:         logger,
	}
}

func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req models.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request payload",
			zap.Error(err),
			zap.String("path", c.Request.URL.Path))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем ID пользователя из контекста (установлен middleware)
	userSSOID, exists := c.Get("user_sso_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User authentication failed"})
		return
	}

	// Преобразуем ID пользователя в строку
	userID := strconv.Itoa(userSSOID.(int))

	// Получаем ключ идемпотентности из заголовка
	idempotencyKey := c.GetHeader("Idempotency-Key")

	payment, err := h.paymentService.InitiatePayment(
		c.Request.Context(),
		req.OrderID,
		userID,
		req.Amount,
		req.Currency,
		req.Description,
		req.ReturnURL,
		req.Metadata,
		idempotencyKey,
	)

	if err != nil {
		h.logger.Error("Failed to create payment",
			zap.Error(err),
			zap.String("order_id", req.OrderID),
			zap.String("user_id", userID))

		switch err {
		case services.ErrInvalidAmount, services.ErrInvalidCurrency:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment"})
		}
		return
	}

	c.JSON(http.StatusCreated, models.CreatePaymentResponse{
		PaymentID:       payment.ID,
		ConfirmationURL: payment.PaymentURL,
	})
}

func (h *PaymentHandler) GetPayment(c *gin.Context) {
	paymentID := c.Param("id")

	payment, err := h.paymentService.GetPayment(c.Request.Context(), paymentID)
	if err != nil {
		h.logger.Error("ERROR: Failed to get payment",
			zap.Error(err),
			zap.String("payment_id", paymentID))

		if err == services.ErrPaymentNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get payment"})
		}
		return
	}
	c.JSON(http.StatusOK, payment)
}

func (h *PaymentHandler) CapturePayment(c *gin.Context) {
	paymentID := c.Param("id")

	payment, err := h.paymentService.CapturePayment(c.Request.Context(), paymentID)
	if err != nil {
		h.logger.Error("ERROR: Failed to capture payment",
			zap.Error(err),
			zap.String("payment_id", paymentID))

		if err == services.ErrPaymentNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to capture payment"})
		}
		return
	}

	c.JSON(http.StatusOK, payment)
}

func (h *PaymentHandler) CancelPayment(c *gin.Context) {
	paymentID := c.Param("id")

	payment, err := h.paymentService.CancelPayment(c.Request.Context(), paymentID)
	if err != nil {
		h.logger.Error("Failed to cancel payment",
			zap.Error(err),
			zap.String("payment_id", paymentID))

		if err == services.ErrPaymentNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel payment"})
		}
		return
	}

	c.JSON(http.StatusOK, payment)
}
