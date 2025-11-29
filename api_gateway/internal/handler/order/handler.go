package order

import (
	"api_gateway/internal/client/cart"
	"api_gateway/internal/client/order"
	"api_gateway/internal/client/product"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	orderv1 "github.com/stpnv0/protos/gen/go/order"
)

type Handler struct {
	orderClient   *order.Client
	productClient *product.Client
	cartClient    *cart.Client
	log           *slog.Logger
}

func New(orderClient *order.Client, productClient *product.Client, cartClient *cart.Client, log *slog.Logger) *Handler {
	return &Handler{
		orderClient:   orderClient,
		productClient: productClient,
		cartClient:    cartClient,
		log:           log,
	}
}

type CreateOrderRequest struct {
	Items []OrderItemRequest `json:"items" binding:"required"`
}

type OrderItemRequest struct {
	SneakerID int64 `json:"sneaker_id" binding:"required"`
	Quantity  int32 `json:"quantity" binding:"required"`
}

func (h *Handler) CreateOrder(c *gin.Context) {
	userID, exists := c.Get("user_sso_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Fetch product details to get current prices
	var items []*orderv1.OrderItem
	for _, item := range req.Items {
		// Get product details
		sneaker, err := h.productClient.GetSneakerByID(c.Request.Context(), item.SneakerID)
		if err != nil {
			h.log.Error("failed to get sneaker", slog.Int64("sneaker_id", item.SneakerID), slog.String("error", err.Error()))
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sneaker_id"})
			return
		}

		items = append(items, &orderv1.OrderItem{
			SneakerId:       int32(item.SneakerID),
			Quantity:        item.Quantity,
			PriceAtPurchase: int32(sneaker.Price),
		})
	}

	orderID, err := h.orderClient.CreateOrder(c.Request.Context(), int64(userID.(int)), items)
	if err != nil {
		h.log.Error("failed to create order", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order"})
		return
	}

	// Clear cart after successful order creation
	if err := h.cartClient.ClearCart(c.Request.Context(), userID.(int)); err != nil {
		h.log.Warn("failed to clear cart after order creation", slog.String("error", err.Error()))
		// Don't fail the request, just log the warning
	}

	c.JSON(http.StatusCreated, gin.H{"order_id": orderID})
}

func (h *Handler) GetUserOrders(c *gin.Context) {
	userID, exists := c.Get("user_sso_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	orders, err := h.orderClient.GetUserOrders(c.Request.Context(), int64(userID.(int)))
	if err != nil {
		h.log.Error("failed to get user orders", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user orders"})
		return
	}

	c.JSON(http.StatusOK, orders)
}

func (h *Handler) GetOrder(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	order, err := h.orderClient.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		h.log.Error("failed to get order", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get order"})
		return
	}

	c.JSON(http.StatusOK, order)
}
