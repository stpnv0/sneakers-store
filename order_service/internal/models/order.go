package models

import "time"

const (
	OrderStatusPendingPayment = "PENDING_PAYMENT"
	OrderStatusPaid           = "PAID"
	OrderStatusCancelled      = "CANCELLED"
	OrderStatusShipped        = "SHIPPED"
	OrderStatusPaymentFailed  = "PAYMENT_FAILED"
)

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
