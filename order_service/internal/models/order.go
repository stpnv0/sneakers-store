package models

import "time"

// Order statuses.
const (
	OrderStatusPendingPayment = "PENDING_PAYMENT"
	OrderStatusPaid           = "PAID"
	OrderStatusCancelled      = "CANCELLED"
	OrderStatusShipped        = "SHIPPED"
	OrderStatusPaymentFailed  = "PAYMENT_FAILED"
)

var validOrderStatuses = map[string]struct{}{
	OrderStatusPendingPayment: {},
	OrderStatusPaid:           {},
	OrderStatusCancelled:      {},
	OrderStatusShipped:        {},
	OrderStatusPaymentFailed:  {},
}

func IsValidStatus(status string) bool {
	_, ok := validOrderStatuses[status]
	return ok
}

var validTransitions = map[string][]string{
	OrderStatusPendingPayment: {OrderStatusPaid, OrderStatusPaymentFailed, OrderStatusCancelled},
	OrderStatusPaid:           {OrderStatusShipped, OrderStatusCancelled},
	OrderStatusPaymentFailed:  {OrderStatusPendingPayment, OrderStatusCancelled},
	OrderStatusShipped:        {},
	OrderStatusCancelled:      {},
}

func ValidTransition(from, to string) bool {
	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

type Order struct {
	ID          int       `db:"id"`
	UserID      int       `db:"user_id"`
	Status      string    `db:"status"`
	TotalAmount int       `db:"total_amount"`
	PaymentURL  string    `db:"payment_url"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type OrderItem struct {
	ID              int       `db:"id"`
	OrderID         int       `db:"order_id"`
	SneakerID       int       `db:"sneaker_id"`
	Quantity        int       `db:"quantity"`
	PriceAtPurchase int       `db:"price_at_purchase"`
	CreatedAt       time.Time `db:"created_at"`
}

type OrderWithItems struct {
	Order
	Items []OrderItem
}

type OrderEvent struct {
	EventType   string `json:"event_type"`
	OrderID     int    `json:"order_id"`
	UserID      int    `json:"user_id"`
	Status      string `json:"status"`
	TotalAmount int    `json:"total_amount"`
	PaymentURL  string `json:"payment_url,omitempty"`
	Timestamp   string `json:"timestamp"`
}
