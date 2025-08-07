package router

import (
	"cart_service/internal/handlers"
	"cart_service/internal/middleware"

	"github.com/gin-gonic/gin"
)

func InitRouter(cartHandler *handlers.CartHandler) *gin.Engine {
	router := gin.Default()

	cart := router.Group("/api/v1/cart")
	cart.Use(middleware.ExtractUserID())
	{
		cart.POST("/", cartHandler.AddToCart)
		cart.GET("/", cartHandler.GetCart)
		cart.DELETE("/:id", cartHandler.RemoveFromCart)
		cart.PUT("/:id", cartHandler.UpdateCartItemQuantity)
	}
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "cart_service"})
	})

	return router
}
