package router

import (
	"log/slog"

	cart_handler "api_gateway/internal/handler/cart"
	fav_handler "api_gateway/internal/handler/favourites"
	order_handler "api_gateway/internal/handler/order"
	product_handler "api_gateway/internal/handler/product"
	auth_handler "api_gateway/internal/handler/sso"
	"api_gateway/internal/middleware"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	Product    *product_handler.Handler
	Auth       *auth_handler.Handler
	Cart       *cart_handler.Handler
	Favourites *fav_handler.Handler
	Order      *order_handler.Handler
}

func New(appSecret string, log *slog.Logger, h Handlers, adminChecker middleware.AdminChecker) *gin.Engine {
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.SlogRecovery(log))
	router.Use(middleware.SlogAccessLog(log))
	router.RedirectTrailingSlash = true
	router.RedirectFixedPath = false

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	authMW := middleware.AuthMiddleware(appSecret, log)
	adminMW := middleware.AdminMiddleware(adminChecker, log)

	apiV1 := router.Group("/api/v1")
	{
		productsPublic := apiV1.Group("/products")
		{
			productsPublic.GET("", h.Product.GetAllSneakers)
			productsPublic.GET("/:id", h.Product.GetSneakerByID)
			productsPublic.GET("/batch", h.Product.GetSneakersByIDs)
		}

		authPublic := apiV1.Group("/auth")
		{
			authPublic.POST("/register", h.Auth.Register)
			authPublic.POST("/login", h.Auth.Login)
		}

		auth := apiV1.Group("")
		auth.Use(authMW)
		{
			// Маршруты управления товарами (только для администраторов).
			productsAdmin := auth.Group("/products")
			productsAdmin.Use(adminMW)
			{
				productsAdmin.POST("", h.Product.AddSneaker)
				productsAdmin.POST("/:id/image", h.Product.UpdateProductImage)
			}

			auth.POST("/images/generate-upload-url", h.Product.GenerateUploadURL)

			cartRoutes := auth.Group("/cart")
			{
				cartRoutes.POST("/", h.Cart.AddToCart)
				cartRoutes.GET("/", h.Cart.GetCart)
				cartRoutes.PUT("/:id", h.Cart.UpdateCartItemQuantity)
				cartRoutes.DELETE("/:id", h.Cart.RemoveFromCart)
			}

			favRoutes := auth.Group("/favourites")
			{
				favRoutes.POST("/", h.Favourites.AddToFavourites)
				favRoutes.GET("/", h.Favourites.GetAllFavourites)
				favRoutes.DELETE("/:id", h.Favourites.RemoveFromFavourites)
				favRoutes.GET("/:id", h.Favourites.IsFavourite)
				favRoutes.GET("/batch", h.Favourites.GetFavouritesByIDs)
			}

			orderRoutes := auth.Group("/orders")
			{
				orderRoutes.POST("/", h.Order.CreateOrder)
				orderRoutes.GET("/", h.Order.GetUserOrders)
				orderRoutes.GET("/:id", h.Order.GetOrder)
			}
		}
	}

	return router
}
