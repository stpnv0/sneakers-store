package product

import (
	"api_gateway/internal/client/product"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	productv1 "github.com/stpnv0/protos/gen/go/product"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	client *product.Client
	log    *slog.Logger
}

func NewHandler(client *product.Client, log *slog.Logger) *Handler {
	return &Handler{
		client: client,
		log:    log,
	}
}

func (h *Handler) GetSneakerByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	sneaker, err := h.client.GetSneakerByID(c.Request.Context(), id)
	if err != nil {
		handleGRPCError(c, h.log, err, "failed to get product by id")
		return
	}
	c.JSON(http.StatusOK, sneaker)
}

func (h *Handler) GetSneakersByIDs(c *gin.Context) {
	idsStr := c.Query("ids")
	if idsStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids query parameter is required"})
		return
	}

	idParts := strings.Split(idsStr, ",")
	ids := make([]int64, 0, len(idParts))
	for _, idStr := range idParts {
		id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
		if err != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID in list"})
			return
		}
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		c.JSON(http.StatusOK, gin.H{"sneakers": make([]interface{}, 0)})
		return
	}

	res, err := h.client.GetSneakersByIDs(c.Request.Context(), ids)
	if err != nil {
		handleGRPCError(c, h.log, err, "failed to get products by ids")
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *Handler) GetAllSneakers(c *gin.Context) {
	limit, _ := strconv.ParseUint(c.DefaultQuery("limit", "20"), 10, 64)
	offset, _ := strconv.ParseUint(c.DefaultQuery("offset", "0"), 10, 64)

	sneakers, err := h.client.GetAllSneakers(c.Request.Context(), limit, offset)
	if err != nil {
		handleGRPCError(c, h.log, err, "failed to get all products")
		return
	}
	if sneakers == nil {
		sneakers = make([]*productv1.Sneaker, 0)
	}

	c.JSON(http.StatusOK, gin.H{"sneakers": sneakers})
}

func (h *Handler) AddSneaker(c *gin.Context) {
	var reqBody struct {
		Title string  `json:"title" binding:"required"`
		Price float32 `json:"price" binding:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sneaker, err := h.client.AddSneaker(c.Request.Context(), reqBody.Title, reqBody.Price)
	if err != nil {
		handleGRPCError(c, h.log, err, "failed to add product")
		return
	}
	c.JSON(http.StatusCreated, sneaker)
}

func (h *Handler) GenerateUploadURL(c *gin.Context) {
	var reqBody struct {
		OriginalFilename string `json:"original_filename" binding:"required"`
		ContentType      string `json:"content_type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.client.GenerateUploadURL(c.Request.Context(), reqBody.OriginalFilename, reqBody.ContentType)
	if err != nil {
		handleGRPCError(c, h.log, err, "failed to generate upload url")
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) UpdateProductImage(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}
	var reqBody struct {
		ImageKey string `json:"image_key" binding:"required"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.client.UpdateProductImage(c.Request.Context(), id, reqBody.ImageKey)
	if err != nil {
		handleGRPCError(c, h.log, err, "failed to update product image")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func handleGRPCError(c *gin.Context, log *slog.Logger, err error, message string) {
	st, ok := status.FromError(err)
	if !ok {
		// Это не gRPC ошибка, а, например, сетевая проблема
		log.Error(message, slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	switch st.Code() {
	case codes.NotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
	case codes.InvalidArgument:
		c.JSON(http.StatusBadRequest, gin.H{"error": st.Message()})
	case codes.Unauthenticated:
		c.JSON(http.StatusUnauthorized, gin.H{"error": st.Message()})
	default:
		log.Error(message, slog.String("grpc_code", st.Code().String()), slog.String("error", st.Message()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "an unexpected error occurred"})
	}
}
