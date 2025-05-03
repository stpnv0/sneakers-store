package sneakers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Service
}

func NewHandler(s Service) *Handler {
	return &Handler{
		Service: s,
	}
}

func (h *Handler) AddSneaker(c *gin.Context) {
	var s CreateSneakerReq
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.Service.AddSneaker(c.Request.Context(), &s)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *Handler) DeleteSneaker(c *gin.Context) {
	idParam := c.Param("id")

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// Вызов метода сервиса
	if err := h.Service.DeleteSneaker(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Успешное удаление
	c.JSON(http.StatusOK, gin.H{"message": "sneaker deleted successfully"})
}

func (h *Handler) GetAllSneakers(c *gin.Context) {
	sneakers, err := h.Service.GetAllSneakers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sneakers)
}

func (h *Handler) GetSneakerByID(c *gin.Context) {
	idParam := c.Param("id")

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	sneaker, err := h.Service.GetSneakerByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sneaker)
}

func (h *Handler) GetSneakersByIDs(c *gin.Context) {
	var req struct {
		IDs []int64 `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.IDs) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "too many ids"})
		return
	}

	sneakers, err := h.Service.GetSneakersByIDs(c.Request.Context(), req.IDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sneakers)
}

// GetSneakersByIDsQuery обрабатывает GET запрос для получения товаров по списку ID
func (h *Handler) GetSneakersByIDsQuery(c *gin.Context) {
	// Получаем параметр ids из запроса
	idsParam := c.Query("ids")
	if idsParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids parameter is required"})
		return
	}

	// Парсим строку с ID в массив
	ids, err := h.Service.ParseIDsString(idsParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ids format: " + err.Error()})
		return
	}

	// Проверяем количество ID
	if len(ids) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "too many ids (max 100)"})
		return
	}

	// Получаем товары по ID
	sneakers, err := h.Service.GetSneakersByIDs(c.Request.Context(), ids)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sneakers)
}