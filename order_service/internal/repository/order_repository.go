package repository

import (
	"context"
	"fmt"
	"order_service/internal/models"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{pool: pool}
}

func (r *OrderRepository) Create(ctx context.Context, order *models.Order, items []models.OrderItem) (*models.OrderWithItems, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var orderID int
	err = tx.QueryRow(ctx,
		`INSERT INTO orders (user_id, status, total_amount, created_at, updated_at) 
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		order.UserID, order.Status, order.TotalAmount, time.Now(), time.Now(),
	).Scan(&orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	for _, item := range items {
		_, err = tx.Exec(ctx,
			`INSERT INTO order_items (order_id, sneaker_id, quantity, price_at_purchase, created_at)
			 VALUES ($1, $2, $3, $4, $5)`,
			orderID, item.SneakerID, item.Quantity, item.PriceAtPurchase, time.Now(),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create order item: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return r.GetByID(ctx, orderID)
}

func (r *OrderRepository) GetByID(ctx context.Context, orderID int) (*models.OrderWithItems, error) {
	var order models.Order
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, status, total_amount, COALESCE(payment_url, '') as payment_url, created_at, updated_at 
		 FROM orders WHERE id = $1`,
		orderID,
	).Scan(&order.ID, &order.UserID, &order.Status, &order.TotalAmount, &order.PaymentURL, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, order_id, sneaker_id, quantity, price_at_purchase, created_at
		 FROM order_items WHERE order_id = $1`,
		orderID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.SneakerID, &item.Quantity, &item.PriceAtPurchase, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		items = append(items, item)
	}

	return &models.OrderWithItems{
		Order: order,
		Items: items,
	}, nil
}

func (r *OrderRepository) GetUserOrders(ctx context.Context, userID int) ([]*models.OrderWithItems, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, status, total_amount, COALESCE(payment_url, '') as payment_url, created_at, updated_at 
		 FROM orders WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user orders: %w", err)
	}
	defer rows.Close()

	var orders []*models.OrderWithItems
	for rows.Next() {
		var order models.Order
		if err := rows.Scan(&order.ID, &order.UserID, &order.Status, &order.TotalAmount, &order.PaymentURL, &order.CreatedAt, &order.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		itemRows, err := r.pool.Query(ctx,
			`SELECT id, order_id, sneaker_id, quantity, price_at_purchase, created_at
			 FROM order_items WHERE order_id = $1`,
			order.ID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to get order items: %w", err)
		}

		var items []models.OrderItem
		for itemRows.Next() {
			var item models.OrderItem
			if err := itemRows.Scan(&item.ID, &item.OrderID, &item.SneakerID, &item.Quantity, &item.PriceAtPurchase, &item.CreatedAt); err != nil {
				itemRows.Close()
				return nil, fmt.Errorf("failed to scan order item: %w", err)
			}
			items = append(items, item)
		}
		itemRows.Close()

		orders = append(orders, &models.OrderWithItems{
			Order: order,
			Items: items,
		})
	}

	return orders, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, orderID int, status string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3`,
		status, time.Now(), orderID,
	)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}
	return nil
}

func (r *OrderRepository) UpdatePaymentURL(ctx context.Context, orderID int, paymentURL string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE orders SET payment_url = $1, updated_at = $2 WHERE id = $3`,
		paymentURL, time.Now(), orderID,
	)
	if err != nil {
		return fmt.Errorf("failed to update payment URL: %w", err)
	}
	return nil
}
