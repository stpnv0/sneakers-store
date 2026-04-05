package order

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	orderv1 "github.com/stpnv0/protos/gen/go/order"
	productv1 "github.com/stpnv0/protos/gen/go/product"

	"api_gateway/internal/middleware"
)

type OrderClient interface {
	CreateOrder(ctx context.Context, userID int64, items []*orderv1.OrderItem) (*orderv1.Order, error)
	GetOrder(ctx context.Context, orderID int64) (*orderv1.Order, error)
	GetUserOrders(ctx context.Context, userID int64) ([]*orderv1.Order, error)
}

type ProductLookup interface {
	GetSneakerByID(ctx context.Context, id int64) (*productv1.Sneaker, error)
	GetSneakersByIDs(ctx context.Context, ids []int64) ([]*productv1.Sneaker, error)
}

type CartClearer interface {
	ClearCart(ctx context.Context, userID int64) error
}

type Handler struct {
	orderClient   OrderClient
	productClient ProductLookup
	cartClient    CartClearer
	log           *slog.Logger
}

func New(orderClient OrderClient, productClient ProductLookup, cartClient CartClearer, log *slog.Logger) *Handler {
	return &Handler{
		orderClient:   orderClient,
		productClient: productClient,
		cartClient:    cartClient,
		log:           log,
	}
}

type CreateOrderRequest struct {
	Items []OrderItemRequest `json:"items" binding:"required"`
}

type OrderItemRequest struct {
	SneakerID int64 `json:"sneaker_id" binding:"required"`
	Quantity  int32 `json:"quantity" binding:"required,min=1"`
}

func (h *Handler) CreateOrder(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Собираем все ID товаров для batch
	sneakerIDs := make([]int64, len(req.Items))
	for i, item := range req.Items {
		sneakerIDs[i] = item.SneakerID
	}

	sneakers, err := h.productClient.GetSneakersByIDs(c.Request.Context(), sneakerIDs)
	if err != nil {
		h.log.Error("failed to get sneakers", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sneaker_id"})
		return
	}

	priceMap := make(map[int64]int64, len(sneakers))
	for _, s := range sneakers {
		priceMap[s.GetId()] = s.GetPriceKopecks()
	}

	var items []*orderv1.OrderItem
	for _, item := range req.Items {
		price, ok := priceMap[item.SneakerID]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sneaker_id"})
			return
		}
		items = append(items, &orderv1.OrderItem{
			SneakerId:              item.SneakerID,
			Quantity:               item.Quantity,
			PriceAtPurchaseKopecks: price,
		})
	}

	order, err := h.orderClient.CreateOrder(c.Request.Context(), userID, items)
	if err != nil {
		h.log.Error("failed to create order", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order"})
		return
	}

	if err := h.cartClient.ClearCart(c.Request.Context(), userID); err != nil {
		h.log.Warn("failed to clear cart after order creation", slog.String("error", err.Error()))
	}

	c.JSON(http.StatusCreated, gin.H{
		"order_id":    order.GetId(),
		"payment_url": order.GetPaymentUrl(),
		"status":      order.GetStatus(),
	})
}

func (h *Handler) GetUserOrders(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	orders, err := h.orderClient.GetUserOrders(c.Request.Context(), userID)
	if err != nil {
		h.log.Error("failed to get user orders", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user orders"})
		return
	}

	c.JSON(http.StatusOK, orders)
}

func (h *Handler) GetOrder(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || orderID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	order, err := h.orderClient.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		h.log.Error("failed to get order", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get order"})
		return
	}

	// Проверяем, что заказ принадлежит аутентифицированному пользователю
	if order.GetUserId() != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, order)
}
