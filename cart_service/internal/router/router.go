package router

import (
	"cart_service/internal/handlers"
	"cart_service/internal/middleware"

	"github.com/gin-gonic/gin"
)

func InitRouter(cartHandler *handlers.CartHandler) *gin.Engine {
	router := gin.Default()

	// CORS middleware
	router.Use(middleware.CORS())

	// API группа
	api := router.Group("/api/v1")
	{
		// Маршруты для корзины с аутентификацией
		cart := api.Group("/cart")
		cart.Use(middleware.AuthMiddleware())
		{
			cart.POST("", cartHandler.AddToCart)
			cart.GET("", cartHandler.GetCart)
			cart.DELETE("/:id", cartHandler.RemoveFromCart)
			cart.PUT("/:id", cartHandler.UpdateCartItemQuantity)
		}

		// Проверка здоровья сервиса
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})
	}

	return router
}
