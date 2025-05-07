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
	fav := router.Group("")
	fav.Use(middleware.AuthMiddleware())
	{
		fav.POST("", favHandler.AddToFavourite)
		fav.GET("", favHandler.GetAllFavourites)
		fav.DELETE("/:id", favHandler.DeleteFavourite)
		fav.GET("/:id", favHandler.IsFavourite)
		fav.GET("/batch", favHandler.GetFavouritesByIDs)
		fav.GET("/debug", favHandler.DebugFavourites)
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "favourites_service"})
	})

	return router
}
