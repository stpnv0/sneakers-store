package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"order_service/internal/models"
)

type OrderRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{pool: pool}
}

func (r *OrderRepository) Create(ctx context.Context, order *models.Order, items []models.OrderItem) (*models.OrderWithItems, error) {
	const op = "repository.OrderRepository.Create"

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: begin tx: %w", op, err)
	}
	defer tx.Rollback(ctx)

	now := time.Now()

	var orderID int
	err = tx.QueryRow(ctx,
		`INSERT INTO orders (user_id, status, total_amount, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		order.UserID, order.Status, order.TotalAmount, now, now,
	).Scan(&orderID)
	if err != nil {
		return nil, fmt.Errorf("%s: insert order: %w", op, err)
	}

	for i, item := range items {
		_, err = tx.Exec(ctx,
			`INSERT INTO order_items (order_id, sneaker_id, quantity, price_at_purchase, created_at)
			 VALUES ($1, $2, $3, $4, $5)`,
			orderID, item.SneakerID, item.Quantity, item.PriceAtPurchase, now,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: insert item[%d]: %w", op, i, err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%s: commit tx: %w", op, err)
	}

	return r.GetByID(ctx, orderID)
}

func (r *OrderRepository) GetByID(ctx context.Context, orderID int) (*models.OrderWithItems, error) {
	const op = "repository.OrderRepository.GetByID"

	var o models.Order
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, status, total_amount,
		        COALESCE(payment_url, '') AS payment_url,
		        created_at, updated_at
		 FROM orders WHERE id = $1`, orderID,
	).Scan(&o.ID, &o.UserID, &o.Status, &o.TotalAmount, &o.PaymentURL, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("%s: query order: %w", op, err)
	}

	items, err := r.getItemsByOrderID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &models.OrderWithItems{Order: o, Items: items}, nil
}

func (r *OrderRepository) GetUserOrders(ctx context.Context, userID int) ([]*models.OrderWithItems, error) {
	const op = "repository.OrderRepository.GetUserOrders"

	rows, err := r.pool.Query(ctx,
		`SELECT o.id, o.user_id, o.status, o.total_amount,
		        COALESCE(o.payment_url, '') AS payment_url,
		        o.created_at, o.updated_at,
		        oi.id, oi.order_id, oi.sneaker_id, oi.quantity, oi.price_at_purchase, oi.created_at
		 FROM orders o
		 LEFT JOIN order_items oi ON o.id = oi.order_id
		 WHERE o.user_id = $1
		 ORDER BY o.created_at DESC, oi.id`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: query orders: %w", op, err)
	}
	defer rows.Close()

	ordersMap := make(map[int]*models.OrderWithItems)
	var orderIDs []int

	for rows.Next() {
		var o models.Order
		var itemID, itemOrderID, itemSneakerID, itemQuantity, itemPrice *int
		var itemCreatedAt *time.Time

		if err := rows.Scan(
			&o.ID, &o.UserID, &o.Status, &o.TotalAmount, &o.PaymentURL, &o.CreatedAt, &o.UpdatedAt,
			&itemID, &itemOrderID, &itemSneakerID, &itemQuantity, &itemPrice, &itemCreatedAt,
		); err != nil {
			return nil, fmt.Errorf("%s: scan row: %w", op, err)
		}

		owi, exists := ordersMap[o.ID]
		if !exists {
			owi = &models.OrderWithItems{Order: o}
			ordersMap[o.ID] = owi
			orderIDs = append(orderIDs, o.ID)
		}

		if itemID != nil {
			owi.Items = append(owi.Items, models.OrderItem{
				ID:              *itemID,
				OrderID:         *itemOrderID,
				SneakerID:       *itemSneakerID,
				Quantity:        *itemQuantity,
				PriceAtPurchase: *itemPrice,
				CreatedAt:       *itemCreatedAt,
			})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows iteration: %w", op, err)
	}

	result := make([]*models.OrderWithItems, 0, len(orderIDs))
	for _, id := range orderIDs {
		result = append(result, ordersMap[id])
	}

	return result, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, orderID int, newStatus, expectedCurrentStatus string) error {
	const op = "repository.OrderRepository.UpdateStatus"

	ct, err := r.pool.Exec(ctx,
		`UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3 AND status = $4`,
		newStatus, time.Now(), orderID, expectedCurrentStatus,
	)
	if err != nil {
		return fmt.Errorf("%s: exec: %w", op, err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("%s: order %d not found or status already changed from %q", op, orderID, expectedCurrentStatus)
	}
	return nil
}

func (r *OrderRepository) UpdatePaymentURL(ctx context.Context, orderID int, paymentURL string) error {
	const op = "repository.OrderRepository.UpdatePaymentURL"

	ct, err := r.pool.Exec(ctx,
		`UPDATE orders SET payment_url = $1, updated_at = $2 WHERE id = $3`,
		paymentURL, time.Now(), orderID,
	)
	if err != nil {
		return fmt.Errorf("%s: exec: %w", op, err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("%s: order %d not found", op, orderID)
	}
	return nil
}

func (r *OrderRepository) getItemsByOrderID(ctx context.Context, orderID int) ([]models.OrderItem, error) {
	const op = "repository.OrderRepository.getItemsByOrderID"

	rows, err := r.pool.Query(ctx,
		`SELECT id, order_id, sneaker_id, quantity, price_at_purchase, created_at
		 FROM order_items WHERE order_id = $1`, orderID,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: query: %w", op, err)
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.SneakerID, &item.Quantity, &item.PriceAtPurchase, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("%s: scan: %w", op, err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows iteration: %w", op, err)
	}

	return items, nil
}
