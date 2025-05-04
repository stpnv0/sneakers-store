package router

import (
	"fav_service/internal/handlers"
	"fav_service/internal/middleware"

	"github.com/gin-gonic/gin"
)

func InitRouter(favHandler *handlers.FavHandler) *gin.Engine {
	router := gin.Default()

	// CORS middleware
	router.Use(middleware.CORS())

	// API группа
	api := router.Group("/api/v1")
	{
		// Маршруты для избранного с аутентификацией
		favourites := api.Group("/favourites")
		favourites.Use(middleware.AuthMiddleware())
		{
			favourites.POST("", favHandler.AddToFavourite)
			favourites.GET("", favHandler.GetAllFavourites)
			favourites.DELETE("/:id", favHandler.DeleteFavourite)
			favourites.GET("/:id", favHandler.IsFavourite)
			favourites.GET("/batch", favHandler.GetFavouritesByIDs)

			// отладочный эндпоинт
			favourites.GET("/debug", favHandler.DebugFavourites)
		}

		// Проверка здоровья сервиса
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})
	}

	return router
}
