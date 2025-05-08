// internal/api/router/router.go
package router

import (
	handlers "payment/internal/handler"
	"payment/internal/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func SetupRouter(paymentHandler *handlers.PaymentHandler, webhookHandler *handlers.WebhookHandler, logger *zap.Logger) *gin.Engine {
	router := gin.Default()

	// CORS middleware
	router.Use(middleware.CORS())

	api := router.Group("/api/v1")

	protected := api.Group("/payments")
	protected.Use(middleware.AuthMiddleware(logger))
	{
		protected.POST("", paymentHandler.CreatePayment)
		protected.GET("/:id", paymentHandler.GetPayment)
		protected.POST("/:id/capture", paymentHandler.CapturePayment)
		protected.POST("/:id/cancel", paymentHandler.CancelPayment)
	}

	public := api.Group("/webhooks")
	{
		public.POST("/yookassa", webhookHandler.HandleYooKassaWebhook)
	}

	// Эндпоинт для проверки работоспособности
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "payment_service",
		})
	})

	return router
}
