package router

import (
	"api_gateway/internal/config"
	"api_gateway/internal/middleware"
	"api_gateway/internal/proxy"

	"github.com/gin-gonic/gin"
)

func New(cfg *config.Config) *gin.Engine {
	router := gin.Default()

	router.Use(middleware.CORSMiddleware())

	authMiddleware := middleware.AuthMiddleware(cfg.AppSecret)

	apiV1 := router.Group("/api/v1")
	{
		// публичные роуты
		apiV1.Any("/items", proxy.ProxyHandler(cfg.Downstream.Backend, "/api/v1"))
		apiV1.Any("/items/*path", proxy.ProxyHandler(cfg.Downstream.Backend, "/api/v1"))
		apiV1.POST("/auth/login", proxy.ProxyHandler(cfg.Downstream.Backend, "/api/v1"))
		apiV1.POST("/auth/register", proxy.ProxyHandler(cfg.Downstream.Backend, "/api/v1"))

		// защищенные роуты
		auth := apiV1.Group("")
		auth.Use(authMiddleware)
		{
			// Маршруты для корзины
			cart := auth.Group("/cart")
			{
				cart.Any("", proxy.ProxyHandler(cfg.Downstream.Cart, "/api/v1/cart"))
				cart.Any("/*path", proxy.ProxyHandler(cfg.Downstream.Cart, "/api/v1/cart"))
			}

			// Маршруты для избранного
			favourites := auth.Group("/favourites")
			{
				favourites.Any("", proxy.ProxyHandler(cfg.Downstream.Favourites, "/api/v1/favourites"))
				favourites.Any("/*path", proxy.ProxyHandler(cfg.Downstream.Favourites, "/api/v1/favourites"))
			}
		}
	}

	return router
}
