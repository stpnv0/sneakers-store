package cart

import (
	"net/http"

	"context"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	client CartClient
}

type CartClient interface {
	AddToCart(ctx context.Context, userSSOID, sneakerID, quantity int) error
	GetCart(ctx context.Context, userSSOID int) ([]byte, error)
	UpdateCartItemQuantity(ctx context.Context, userSSOID int, itemID string, quantity int) error
	DeleteFromCart(ctx context.Context, userSSOID int, itemID string) error
}

func NewHandler(client CartClient) *Handler {
	return &Handler{client: client}
}

func (h *Handler) AddToCart(c *gin.Context) {
	userSSOID, exists := c.Get("user_sso_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		SneakerID int `json:"sneaker_id" binding:"required"`
		Quantity  int `json:"quantity"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.client.AddToCart(c.Request.Context(), userSSOID.(int), req.SneakerID, req.Quantity); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add to cart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Added to cart"})
}

func (h *Handler) UpdateCartItemQuantity(c *gin.Context) {
	userSSOID, exists := c.Get("user_sso_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	itemID := c.Param("id")
	if itemID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cart item ID"})
		return
	}

	var req struct {
		Quantity int `json:"quantity" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.client.UpdateCartItemQuantity(c.Request.Context(), userSSOID.(int), itemID, req.Quantity); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cart item quantity"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cart item quantity updated"})
}

func (h *Handler) DeleteFromCart(c *gin.Context) {
	userSSOID, exists := c.Get("user_sso_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	itemID := c.Param("id")
	if itemID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cart item ID"})
		return
	}

	if err := h.client.DeleteFromCart(c.Request.Context(), userSSOID.(int), itemID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete from cart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deleted from cart"})
}

func (h *Handler) GetAllCart(c *gin.Context) {
	userSSOID, exists := c.Get("user_sso_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	cartData, err := h.client.GetCart(c.Request.Context(), userSSOID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cart"})
		return
	}

	c.Data(http.StatusOK, "application/json", cartData)
}
