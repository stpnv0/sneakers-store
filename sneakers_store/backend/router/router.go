package router

import (
	"sneakers-store/internal/auth"
	"sneakers-store/internal/config"
	"sneakers-store/internal/sneakers"

	"github.com/gin-gonic/gin"
)

var r *gin.Engine

func InitRouter(sneakerHandler *sneakers.Handler, authHandler *auth.Handler, cfg *config.Config) *gin.Engine {
	r = gin.Default()

	authGroup := r.Group("/auth")
	{
		authGroup.POST("/login", func(c *gin.Context) { authHandler.Login(c.Writer, c.Request) })
		authGroup.POST("/register", func(c *gin.Context) { authHandler.Register(c.Writer, c.Request) })
	}

	itemsGroup := r.Group("/items")
	{
		itemsGroup.POST("", sneakerHandler.AddSneaker)
		itemsGroup.POST("/batch", sneakerHandler.GetSneakersByIDs)
		itemsGroup.GET("", sneakerHandler.GetAllSneakers) // Этот роут обработает /items
		itemsGroup.DELETE("/:id", sneakerHandler.DeleteSneaker)
		itemsGroup.GET("/batch", sneakerHandler.GetSneakersByIDsQuery)
	}
	return r
}

func Start(addr string) error {
	return r.Run(addr)
}
