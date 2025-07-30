package handlers

import (
	"cart_service/internal/services"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CartHandler struct {
	CartService services.CartService
	Log         *slog.Logger
}

func NewCartHandler(cartService services.CartService, log *slog.Logger) *CartHandler {
	return &CartHandler{
		CartService: cartService,
		Log:         log,
	}
}

func (h *CartHandler) AddToCart(c *gin.Context) {
	userSSOID, exists := c.Get("user_sso_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		SneakerID int `json:"sneaker_id" binding:"required"`
		Quantity  int `json:"quantity" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.Log.Info("adding item to cart", slog.Any("user_id", userSSOID), slog.Int("sneaker_id", req.SneakerID))

	if err := h.CartService.AddToCart(c.Request.Context(), userSSOID.(int), req.SneakerID, req.Quantity); err != nil {
		h.Log.Error("failed to add item to cart", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add item to cart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item added to cart successfully"})
}

func (h *CartHandler) GetCart(c *gin.Context) {
	const op = "handler.GetCart"
	log := h.Log.With(slog.String("op", op))

	userSSOID, _ := c.Get("user_sso_id")
	userID := userSSOID.(int)
	log = log.With(slog.Int("user_id", userID))

	cart, err := h.CartService.GetCart(c.Request.Context(), userID)
	if err != nil {
		log.Error("failed to get cart from service", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cart"})
		return
	}

	if cart == nil {
		c.JSON(http.StatusOK, gin.H{"items": []interface{}{}})
		return
	}

	c.JSON(http.StatusOK, cart)
}

func (h *CartHandler) UpdateCartItemQuantity(c *gin.Context) {
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
		Quantity int `json:"quantity" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.Log.Info("updating item quantity", slog.Any("user_id", userSSOID), slog.String("item_id", itemID), slog.Int("quantity", req.Quantity))
	// Обновляем количество товара в корзине
	if err := h.CartService.UpdateCartItemQuantity(c.Request.Context(), userSSOID.(int), itemID, req.Quantity); err != nil {
		h.Log.Error("failed to update item quantity", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update item quantity"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item quantity updated successfully"})
}

func (h *CartHandler) RemoveFromCart(c *gin.Context) {
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

	h.Log.Info("removing item from cart", slog.Any("user_id", userSSOID), slog.String("item_id", itemID))

	if err := h.CartService.RemoveFromCart(c.Request.Context(), userSSOID.(int), itemID); err != nil {
		h.Log.Error("failed to remove item from cart", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove item from cart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item removed from cart successfully"})
}
