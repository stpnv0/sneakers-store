package cart

import (
	"api_gateway/internal/client/cart"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	cartv1 "github.com/stpnv0/protos/gen/go/cart"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	cartClient *cart.Client
	log        *slog.Logger
}

func NewHandler(cartClient *cart.Client, log *slog.Logger) *Handler {
	return &Handler{
		cartClient: cartClient,
		log:        log,
	}
}

// AddToCart handles POST /api/v1/cart
func (h *Handler) AddToCart(c *gin.Context) {
	const op = "handler.cart.AddToCart"

	// Extract user ID from context (set by auth middleware)
	userSSOID, exists := c.Get("user_sso_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		SneakerID int32 `json:"sneaker_id" binding:"required"`
		Quantity  int32 `json:"quantity" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.cartClient.AddToCart(c.Request.Context(), userSSOID.(int), req.SneakerID, req.Quantity)
	if err != nil {
		h.log.Error("failed to add to cart", slog.String("op", op), slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add item to cart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item added to cart successfully"})
}

// GetCart handles GET /api/v1/cart
func (h *Handler) GetCart(c *gin.Context) {
	const op = "handler.cart.GetCart"

	// Extract user ID from context
	userSSOID, exists := c.Get("user_sso_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	cart, err := h.cartClient.GetCart(c.Request.Context(), userSSOID.(int))
	if err != nil {
		h.log.Error("failed to get cart", slog.String("op", op), slog.String("error", err.Error()))

		// Handle specific gRPC errors
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Cart not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cart"})
		return
	}

	// Convert proto cart to JSON response
	response := convertCartToJSON(cart)
	c.JSON(http.StatusOK, response)
}

// UpdateCartItemQuantity handles PUT /api/v1/cart/:id
func (h *Handler) UpdateCartItemQuantity(c *gin.Context) {
	const op = "handler.cart.UpdateCartItemQuantity"

	// Extract user ID from context
	userSSOID, exists := c.Get("user_sso_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	itemID := c.Param("id")
	if itemID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Item ID is required"})
		return
	}

	var req struct {
		Quantity int32 `json:"quantity" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.cartClient.UpdateCartItemQuantity(c.Request.Context(), userSSOID.(int), itemID, req.Quantity)
	if err != nil {
		h.log.Error("failed to update cart item", slog.String("op", op), slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update item quantity"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item quantity updated successfully"})
}

// RemoveFromCart handles DELETE /api/v1/cart/:id
func (h *Handler) RemoveFromCart(c *gin.Context) {
	const op = "handler.cart.RemoveFromCart"

	// Extract user ID from context
	userSSOID, exists := c.Get("user_sso_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	itemID := c.Param("id")
	if itemID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Item ID is required"})
		return
	}

	err := h.cartClient.RemoveFromCart(c.Request.Context(), userSSOID.(int), itemID)
	if err != nil {
		h.log.Error("failed to remove from cart", slog.String("op", op), slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove item from cart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item removed from cart successfully"})
}

// Helper to convert proto Cart to JSON
func convertCartToJSON(cart *cartv1.Cart) map[string]interface{} {
	items := make([]map[string]interface{}, 0, len(cart.GetItems()))

	for _, item := range cart.GetItems() {
		items = append(items, map[string]interface{}{
			"id":         item.GetId(),
			"sneaker_id": item.GetSneakerId(),
			"quantity":   item.GetQuantity(),
			"added_at":   item.GetAddedAt(),
		})
	}

	return map[string]interface{}{
		"user_sso_id": cart.GetUserId(),
		"items":       items,
		"updated_at":  cart.GetUpdatedAt(),
	}
}
