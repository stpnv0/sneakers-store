package cart

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	cartv1 "github.com/stpnv0/protos/gen/go/cart"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"api_gateway/internal/middleware"
)

type CartClient interface {
	AddToCart(ctx context.Context, userID int64, sneakerID int64, quantity int32) error
	GetCart(ctx context.Context, userID int64) (*cartv1.Cart, error)
	UpdateCartItemQuantity(ctx context.Context, userID int64, itemID string, quantity int32) error
	RemoveFromCart(ctx context.Context, userID int64, itemID string) error
	ClearCart(ctx context.Context, userID int64) error
}

type Handler struct {
	cartClient CartClient
	log        *slog.Logger
}

func NewHandler(cartClient CartClient, log *slog.Logger) *Handler {
	return &Handler{cartClient: cartClient, log: log}
}

// AddToCart - POST /api/v1/cart/
func (h *Handler) AddToCart(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req struct {
		SneakerID int64 `json:"sneaker_id" binding:"required"`
		Quantity  int32 `json:"quantity" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.cartClient.AddToCart(c.Request.Context(), userID, req.SneakerID, req.Quantity); err != nil {
		h.log.Error("failed to add to cart", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add item to cart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "item added to cart successfully"})
}

// GetCart - GET /api/v1/cart/
func (h *Handler) GetCart(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	cart, err := h.cartClient.GetCart(c.Request.Context(), userID)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "cart not found"})
			return
		}
		h.log.Error("failed to get cart", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get cart"})
		return
	}

	c.JSON(http.StatusOK, convertCartToJSON(cart))
}

// UpdateCartItemQuantity - PUT /api/v1/cart/:id
func (h *Handler) UpdateCartItemQuantity(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	itemID := c.Param("id")
	if itemID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "item ID is required"})
		return
	}

	var req struct {
		Quantity int32 `json:"quantity" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.cartClient.UpdateCartItemQuantity(c.Request.Context(), userID, itemID, req.Quantity); err != nil {
		h.log.Error("failed to update cart item", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update item quantity"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "item quantity updated successfully"})
}

// RemoveFromCart - DELETE /api/v1/cart/:id
func (h *Handler) RemoveFromCart(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	itemID := c.Param("id")
	if itemID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "item ID is required"})
		return
	}

	if err := h.cartClient.RemoveFromCart(c.Request.Context(), userID, itemID); err != nil {
		h.log.Error("failed to remove from cart", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove item from cart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "item removed from cart successfully"})
}

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
