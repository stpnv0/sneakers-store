package fav

import (
	"net/http"

	"context"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	client FavClient
}

type FavClient interface {
	AddToFavourite(ctx context.Context, userSSOID, sneakerID int) error
	GetAllFavourites(ctx context.Context, userSSOID int) ([]byte, error)
	DeleteFavourite(ctx context.Context, userSSOID int, itemID string) error
}

func NewHandler(client FavClient) *Handler {
	return &Handler{client: client}
}

func (h *Handler) AddToFavourite(c *gin.Context) {
	userSSOID, exists := c.Get("user_sso_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		SneakerID int `json:"sneaker_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.client.AddToFavourite(c.Request.Context(), userSSOID.(int), req.SneakerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add to favourites"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Added to favourites"})
}

func (h *Handler) DeleteFavourite(c *gin.Context) {
	userSSOID, exists := c.Get("user_sso_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	itemID := c.Param("id")
	if itemID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid favourite item ID"})
		return
	}

	if err := h.client.DeleteFavourite(c.Request.Context(), userSSOID.(int), itemID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete from favourites"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deleted from favourites"})
}

func (h *Handler) GetAllFavourites(c *gin.Context) {
	userSSOID, exists := c.Get("user_sso_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	favData, err := h.client.GetAllFavourites(c.Request.Context(), userSSOID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get favourites"})
		return
	}

	c.Data(http.StatusOK, "application/json", favData)
}
