// Создайте файл internal/handlers/fav.go
package handlers

import (
	"context"
	"fav_service/internal/services"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type FavHandler struct {
	service *services.FavService
}

func NewFavHandler(service *services.FavService) *FavHandler {
	return &FavHandler{service: service}
}

func (h *FavHandler) GetAllFavourites(c *gin.Context) {
	userSSOID, exists := c.Get("user_sso_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	log.Printf("DEBUG: Processing GetAllFavourites request for user %d", userSSOID.(int))

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	favourites, err := h.service.GetAllFavourites(ctx, userSSOID.(int))
	if err != nil {
		log.Printf("ERROR: Failed to get favourites for user %d: %v", userSSOID.(int), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get favourites"})
		return
	}

	log.Printf("DEBUG: Returning %d favourites for user %d", len(favourites), userSSOID.(int))
	c.JSON(http.StatusOK, favourites)
}

func (h *FavHandler) AddToFavourite(c *gin.Context) {
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

	log.Printf("DEBUG: Received request to add sneaker_id=%d to favourites for user_id=%d", req.SneakerID, userSSOID.(int))

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := h.service.AddToFavourite(ctx, userSSOID.(int), req.SneakerID); err != nil {
		log.Printf("ERROR: Failed to add sneaker_id=%d to favourites for user_id=%d: %v", req.SneakerID, userSSOID.(int), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add to favourites"})
		return
	}

	log.Printf("INFO: Successfully added sneaker %d to favourites for user %d", req.SneakerID, userSSOID.(int))
	c.JSON(http.StatusOK, gin.H{"message": "Added to favourites"})
}

func (h *FavHandler) RemoveFromFavourite(c *gin.Context) {
	userSSOID, exists := c.Get("user_sso_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	sneakerIDStr := c.Param("id")
	sneakerID, err := strconv.Atoi(sneakerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sneaker ID"})
		return
	}

	log.Printf("DEBUG: Received request to remove sneaker_id=%d from favourites for user_id=%d", sneakerID, userSSOID.(int))

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := h.service.RemoveFromFavourite(ctx, userSSOID.(int), sneakerID); err != nil {
		log.Printf("ERROR: Failed to remove sneaker_id=%d from favourites for user_id=%d: %v", sneakerID, userSSOID.(int), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove from favourites"})
		return
	}

	log.Printf("INFO: Successfully removed sneaker %d from favourites for user %d", sneakerID, userSSOID.(int))
	c.JSON(http.StatusOK, gin.H{"message": "Removed from favourites"})
}

// DeleteFavourite is an alias for RemoveFromFavourite to maintain compatibility with the main backend
func (h *FavHandler) DeleteFavourite(c *gin.Context) {
	h.RemoveFromFavourite(c)
}

func (h *FavHandler) IsFavourite(c *gin.Context) {
	userSSOID, exists := c.Get("user_sso_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	sneakerIDStr := c.Param("id")
	sneakerID, err := strconv.Atoi(sneakerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sneaker ID"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	isFavourite, err := h.service.IsFavourite(ctx, userSSOID.(int), sneakerID)
	if err != nil {
		log.Printf("ERROR: Failed to check if sneaker_id=%d is favourite for user_id=%d: %v", sneakerID, userSSOID.(int), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check favourite status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"is_favourite": isFavourite})
}

func (h *FavHandler) GetFavouritesByIDs(c *gin.Context) {
	idsParam := c.Query("ids")
	if idsParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids parameter is required"})
		return
	}

	ids, err := h.service.ParseIDsString(idsParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ids format: " + err.Error()})
		return
	}

	if len(ids) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "too many ids (max 100)"})
		return
	}

	_, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	c.JSON(http.StatusOK, ids)
}

// DebugFavourites is a debug endpoint to check what's in the database and cache
func (h *FavHandler) DebugFavourites(c *gin.Context) {
	userSSOID, exists := c.Get("user_sso_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Force refresh from database
	favourites, err := h.service.GetAllFavourites(ctx, userSSOID.(int))
	if err != nil {
		log.Printf("ERROR: Failed to get favourites from DB for user %d: %v", userSSOID.(int), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get favourites from DB"})
		return
	}

	log.Printf("DEBUG: Retrieved %d favourites for user %d", len(favourites), userSSOID.(int))
	c.JSON(http.StatusOK, gin.H{
		"favourites": favourites,
		"message":    "Favourites retrieved",
	})
}