package favourites

import (
	"log/slog"
	"net/http"
	"strconv"

	"api_gateway/internal/client/favourites"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	client *favourites.Client
	log    *slog.Logger
}

func NewHandler(client *favourites.Client, log *slog.Logger) *Handler {
	return &Handler{
		client: client,
		log:    log,
	}
}

func (h *Handler) AddToFavourites(c *gin.Context) {
	const op = "favourites.handler.AddToFavourites"

	userID, exists := c.Get("user_sso_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		SneakerID int `json:"sneaker_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err := h.client.AddToFavourites(c.Request.Context(), int(userID.(int)), req.SneakerID)
	if err != nil {
		h.log.Error("failed to add to favourites", slog.String("op", op), slog.String("error", err.Error()))

		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.Unauthenticated:
				c.JSON(http.StatusUnauthorized, gin.H{"error": st.Message()})
				return
			case codes.InvalidArgument:
				c.JSON(http.StatusBadRequest, gin.H{"error": st.Message()})
				return
			case codes.AlreadyExists:
				c.JSON(http.StatusConflict, gin.H{"error": st.Message()})
				return
			}
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add to favourites"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Added to favourites"})
}

func (h *Handler) RemoveFromFavourites(c *gin.Context) {
	const op = "favourites.handler.RemoveFromFavourites"

	userID, exists := c.Get("user_sso_id")
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

	err = h.client.RemoveFromFavourites(c.Request.Context(), int(userID.(int)), sneakerID)
	if err != nil {
		h.log.Error("failed to remove from favourites", slog.String("op", op), slog.String("error", err.Error()))

		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.Unauthenticated:
				c.JSON(http.StatusUnauthorized, gin.H{"error": st.Message()})
				return
			case codes.NotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": "Item not found in favourites"})
				return
			}
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove from favourites"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Removed from favourites"})
}

func (h *Handler) GetAllFavourites(c *gin.Context) {
	const op = "favourites.handler.GetAllFavourites"

	userID, exists := c.Get("user_sso_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	items, err := h.client.GetFavourites(c.Request.Context(), int(userID.(int)))
	if err != nil {
		h.log.Error("failed to get favourites", slog.String("op", op), slog.String("error", err.Error()))
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.Unauthenticated {
			c.JSON(http.StatusUnauthorized, gin.H{"error": st.Message()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get favourites"})
		return
	}

	// Convert proto items to JSON-friendly structure if needed, or return as is
	// The proto structure is compatible with JSON response
	c.JSON(http.StatusOK, items)
}

func (h *Handler) IsFavourite(c *gin.Context) {
	const op = "favourites.handler.IsFavourite"

	userID, exists := c.Get("user_sso_id")
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

	isFav, err := h.client.IsFavourite(c.Request.Context(), int(userID.(int)), sneakerID)
	if err != nil {
		h.log.Error("failed to check favourite status", slog.String("op", op), slog.String("error", err.Error()))
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.Unauthenticated {
			c.JSON(http.StatusUnauthorized, gin.H{"error": st.Message()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"is_favourite": isFav})
}

// GetFavouritesByIDs is kept for compatibility but returns empty or just echoes IDs
func (h *Handler) GetFavouritesByIDs(c *gin.Context) {
	idsParam := c.Query("ids")
	if idsParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids parameter is required"})
		return
	}

	// Just return the IDs as the original service did (it seems it was a parser helper?)
	// Or return empty list since we don't have this in gRPC yet
	// For now, let's return 501 Not Implemented or just empty list
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented in gRPC version"})
}

// DebugFavourites
func (h *Handler) DebugFavourites(c *gin.Context) {
	// Re-use GetAllFavourites logic for debug
	h.GetAllFavourites(c)
}
