package favourites

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	favv1 "github.com/stpnv0/protos/gen/go/favourites"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"api_gateway/internal/middleware"
)

type FavouritesClient interface {
	AddToFavourites(ctx context.Context, userID, sneakerID int64) error
	RemoveFromFavourites(ctx context.Context, userID, sneakerID int64) error
	GetFavourites(ctx context.Context, userID int64) ([]*favv1.FavouriteItem, error)
	IsFavourite(ctx context.Context, userID, sneakerID int64) (bool, error)
}

type Handler struct {
	client FavouritesClient
	log    *slog.Logger
}

func NewHandler(client FavouritesClient, log *slog.Logger) *Handler {
	return &Handler{client: client, log: log}
}

func (h *Handler) AddToFavourites(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req struct {
		SneakerID int64 `json:"sneaker_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.client.AddToFavourites(c.Request.Context(), userID, req.SneakerID); err != nil {
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
		h.log.Error("failed to add to favourites", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add to favourites"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "added to favourites"})
}

func (h *Handler) RemoveFromFavourites(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	sneakerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sneaker ID"})
		return
	}

	if err := h.client.RemoveFromFavourites(c.Request.Context(), userID, sneakerID); err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.Unauthenticated:
				c.JSON(http.StatusUnauthorized, gin.H{"error": st.Message()})
				return
			case codes.NotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": "item not found in favourites"})
				return
			}
		}
		h.log.Error("failed to remove from favourites", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove from favourites"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "removed from favourites"})
}

func (h *Handler) GetAllFavourites(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	items, err := h.client.GetFavourites(c.Request.Context(), userID)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.Unauthenticated {
			c.JSON(http.StatusUnauthorized, gin.H{"error": st.Message()})
			return
		}
		h.log.Error("failed to get favourites", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get favourites"})
		return
	}

	c.JSON(http.StatusOK, items)
}

func (h *Handler) IsFavourite(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	sneakerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sneaker ID"})
		return
	}

	isFav, err := h.client.IsFavourite(c.Request.Context(), userID, sneakerID)
	if err != nil {
		h.log.Error("failed to check favourite status", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"is_favourite": isFav})
}

func (h *Handler) GetFavouritesByIDs(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented in gRPC version"})
}

func (h *Handler) DebugFavourites(c *gin.Context) {
	h.GetAllFavourites(c)
}
