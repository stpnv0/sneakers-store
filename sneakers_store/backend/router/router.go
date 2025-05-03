package router

import (
	"time"

	"sneakers-store/internal/auth"
	"sneakers-store/internal/config"
	"sneakers-store/internal/middleware"
	"sneakers-store/internal/sneakers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var r *gin.Engine

func InitRouter(sneakerHandler *sneakers.Handler, authHandler *auth.Handler, cfg *config.Config) *gin.Engine {
	r = gin.Default()

	// CORS настройки
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Настройка маршрутов
	api := r.Group("/api/v1")
	{
		// Аутентификация
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/login", func(c *gin.Context) {
				authHandler.Login(c.Writer, c.Request)
			})
			authGroup.POST("/register", func(c *gin.Context) {
				authHandler.Register(c.Writer, c.Request)
			})
		}

		// Товары
		itemsGroup := api.Group("/items")
		{
			itemsGroup.POST("", sneakerHandler.AddSneaker)
			itemsGroup.POST("/batch", sneakerHandler.GetSneakersByIDs)
			itemsGroup.GET("", sneakerHandler.GetAllSneakers)
			itemsGroup.DELETE("/:id", sneakerHandler.DeleteSneaker)
			
			// Добавляем новый эндпоинт для получения товаров по списку ID через GET запрос
			itemsGroup.GET("/batch", sneakerHandler.GetSneakersByIDsQuery)
		}

		// Избранное (проксируем через микросервис)
		favGroup := api.Group(
			"/favourites",
			middleware.AuthMiddleware(),
			middleware.FavProxyMiddleware(cfg.Clients.Favourites.Address),
		)
		{
			// Пустые обработчики: прокси перехватит все запросы
			favGroup.GET("", func(c *gin.Context) {})
			favGroup.POST("", func(c *gin.Context) {})
			favGroup.DELETE("/:id", func(c *gin.Context) {})
		}

		// Корзина (проксируем через микросервис)
		cartGroup := api.Group(
			"/cart",
			middleware.AuthMiddleware(),
			middleware.CartProxyMiddleware(cfg.Clients.Cart.Address),
		)
		{
			cartGroup.GET("", func(c *gin.Context) {})
			cartGroup.POST("", func(c *gin.Context) {})
			cartGroup.PUT("/:id", func(c *gin.Context) {})
			cartGroup.DELETE("/:id", func(c *gin.Context) {})
		}
	}

	return r
}

func Start(addr string) error {
	return r.Run(addr)
}