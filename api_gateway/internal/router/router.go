package router

import (
	"api_gateway/internal/client/cart"
	"api_gateway/internal/client/favourites"
	"api_gateway/internal/client/order"
	"api_gateway/internal/client/product"
	"api_gateway/internal/client/sso"
	"api_gateway/internal/config"
	cart_handler "api_gateway/internal/handler/cart"
	fav_handler "api_gateway/internal/handler/favourites"
	order_handler "api_gateway/internal/handler/order"
	product_handler "api_gateway/internal/handler/product"
	auth_handler "api_gateway/internal/handler/sso"
	"api_gateway/internal/middleware"
	"log/slog"

	"github.com/gin-gonic/gin"
)

func New(cfg *config.Config, log *slog.Logger, productClient *product.Client, ssoClient *sso.Client, cartClient *cart.Client, favClient *favourites.Client, orderClient *order.Client) *gin.Engine {
	router := gin.Default()

	router.RedirectTrailingSlash = true
	router.RedirectFixedPath = false

	authMiddleware := middleware.AuthMiddleware(cfg.AppSecret)

	productHandler := product_handler.NewHandler(productClient, log)
	authHandler := auth_handler.NewHandler(ssoClient, log)
	cartHandler := cart_handler.NewHandler(cartClient, log)
	favHandler := fav_handler.NewHandler(favClient, log)
	orderHandler := order_handler.New(orderClient, productClient, cartClient, log)

	apiV1 := router.Group("/api/v1")
	{

		productsPublic := apiV1.Group("/products")
		{
			productsPublic.GET("", productHandler.GetAllSneakers)
			productsPublic.GET("/:id", productHandler.GetSneakerByID)
			productsPublic.GET("/batch", productHandler.GetSneakersByIDs)
		}

		authPublic := apiV1.Group("/auth")
		{
			authPublic.POST("/register", authHandler.Register)
			authPublic.POST("/login", authHandler.Login)
		}

		auth := apiV1.Group("")
		auth.Use(authMiddleware)
		{
			productsProtected := auth.Group("/products")
			{
				productsProtected.POST("", productHandler.AddSneaker)
				productsProtected.POST("/:id/image", productHandler.UpdateProductImage)
			}

			auth.POST("/images/generate-upload-url", productHandler.GenerateUploadURL)

			cartRoutes := auth.Group("/cart")
			{
				cartRoutes.POST("/", cartHandler.AddToCart)
				cartRoutes.GET("/", cartHandler.GetCart)
				cartRoutes.PUT("/:id", cartHandler.UpdateCartItemQuantity)
				cartRoutes.DELETE("/:id", cartHandler.RemoveFromCart)
			}

			favRoutes := auth.Group("/favourites")
			{
				favRoutes.POST("/", favHandler.AddToFavourites)
				favRoutes.GET("/", favHandler.GetAllFavourites)
				favRoutes.DELETE("/:id", favHandler.RemoveFromFavourites)
				favRoutes.GET("/:id", favHandler.IsFavourite)
				favRoutes.GET("/batch", favHandler.GetFavouritesByIDs)
				favRoutes.GET("/debug", favHandler.DebugFavourites)
			}

			orderRoutes := auth.Group("/orders")
			{
				orderRoutes.POST("/", orderHandler.CreateOrder)
				orderRoutes.GET("/", orderHandler.GetUserOrders)
				orderRoutes.GET("/:id", orderHandler.GetOrder)
			}
		}
	}

	return router
}
