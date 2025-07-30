package router

import (
	"api_gateway/internal/client/product"
	"api_gateway/internal/client/sso"
	"api_gateway/internal/config"
	product_handler "api_gateway/internal/handler/product"
	auth_handler "api_gateway/internal/handler/sso"
	"api_gateway/internal/middleware"
	"api_gateway/internal/proxy"
	"log/slog"

	"github.com/gin-gonic/gin"
)

func New(cfg *config.Config, log *slog.Logger, productClient *product.Client, ssoClient *sso.Client) *gin.Engine {
	router := gin.Default()

	// Отключаем автоматические редиректы, чтобы избежать 307
	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false

	// Подключаем ваши middleware (CORS и т.д.)
	// router.Use(middleware.CORSMiddleware())
	authMiddleware := middleware.AuthMiddleware(cfg.AppSecret)

	productHandler := product_handler.NewHandler(productClient, log)
	authHandler := auth_handler.NewHandler(ssoClient, log)

	apiV1 := router.Group("/api/v1")
	{
		// --- ПУБЛИЧНЫЕ РОУТЫ ---

		productsPublic := apiV1.Group("/products")
		{
			productsPublic.GET("", productHandler.GetAllSneakers)
			productsPublic.GET("/:id", productHandler.GetSneakerByID)
			// ИСПРАВЛЕНО: Роут /batch теперь в публичной группе
			productsPublic.GET("/batch", productHandler.GetSneakersByIDs)
		}

		authPublic := apiV1.Group("/auth")
		{
			authPublic.POST("/register", authHandler.Register)
			authPublic.POST("/login", authHandler.Login)
		}

		// --- ЗАЩИЩЕННЫЕ РОУТЫ ---
		auth := apiV1.Group("")
		auth.Use(authMiddleware)
		{
			// Защищенные роуты для товаров
			productsProtected := auth.Group("/products")
			{
				productsProtected.POST("", productHandler.AddSneaker)
				productsProtected.POST("/:id/image", productHandler.UpdateProductImage)
				// Пример роута для удаления, если он понадобится
				// productsProtected.DELETE("/:id", productHandler.DeleteSneaker)
			}

			// Защищенный роут для генерации URL
			auth.POST("/images/generate-upload-url", productHandler.GenerateUploadURL)

			// --- ИСПРАВЛЕННЫЕ РОУТЫ ДЛЯ ПРОКСИ ---
			// Используем один универсальный роут для каждого сервиса
			cart := auth.Group("/cart")
			{
				cart.Any("/*path", proxy.ProxyHandler(cfg.Downstream.Cart))
			}

			favourites := auth.Group("/favourites")
			{
				favourites.Any("/*path", proxy.ProxyHandler(cfg.Downstream.Favourites))
			}
		}
	}

	return router
}
