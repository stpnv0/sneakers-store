package repository

import (
	"cart_service/internal/models"
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening database connection: %w", err)
	}

	// Проверяем соединение
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	// Устанавливаем параметры пула соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

// SaveCart сохраняет корзину в PostgreSQL
func (r *PostgresRepository) SaveCart(ctx context.Context, cart *models.Cart) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}

	// В случае ошибки откатываем транзакцию
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Вставляем или обновляем запись в таблице carts
	_, err = tx.ExecContext(ctx, `
		INSERT INTO carts (user_sso_id, updated_at)
		VALUES ($1, $2)
		ON CONFLICT (user_sso_id)
		DO UPDATE SET updated_at = $2
	`, cart.UserSSOID, cart.UpdatedAt)
	if err != nil {
		return fmt.Errorf("error upserting cart: %w", err)
	}

	// Удаляем все текущие элементы корзины
	_, err = tx.ExecContext(ctx, `
		DELETE FROM cart_items
		WHERE cart_id = $1
	`, cart.UserSSOID)
	if err != nil {
		return fmt.Errorf("error deleting cart items: %w", err)
	}

	// Вставляем элементы корзины
	for _, item := range cart.Items {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO cart_items (cart_id, user_sso_id, sneaker_id, quantity, added_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, cart.UserSSOID, item.UserSSOID, item.SneakerID, item.Quantity, item.AddedAt, time.Now())
		if err != nil {
			return fmt.Errorf("error inserting cart item: %w", err)
		}
	}

	// Завершаем транзакцию
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

// GetCart получает корзину из PostgreSQL
func (r *PostgresRepository) GetCart(ctx context.Context, userSSOID int) (*models.Cart, error) {
	// Проверяем существование корзины
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM carts WHERE user_sso_id = $1)
	`, userSSOID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("error checking cart existence: %w", err)
	}

	if !exists {
		// Возвращаем пустую корзину, если в БД её нет
		return &models.Cart{
			UserSSOID: userSSOID,
			Items:     []models.CartItem{},
			UpdatedAt: time.Now(),
		}, nil
	}

	// Получаем элементы корзины
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, sneaker_id, quantity, added_at, updated_at
		FROM cart_items
		WHERE cart_id = $1
	`, userSSOID)
	if err != nil {
		return nil, fmt.Errorf("error querying cart items: %w", err)
	}
	defer rows.Close()

	cart := &models.Cart{
		UserSSOID: userSSOID,
		Items:     []models.CartItem{},
	}

	for rows.Next() {
		var item models.CartItem
		var id int
		if err := rows.Scan(&id, &item.SneakerID, &item.Quantity, &item.AddedAt, &cart.UpdatedAt); err != nil {
			return nil, fmt.Errorf("error scanning cart item: %w", err)
		}

		item.ID = fmt.Sprintf("%d", id)
		item.UserSSOID = userSSOID
		item.Synchronized = true

		cart.Items = append(cart.Items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating cart items: %w", err)
	}

	return cart, nil
}

// AddCartItem добавляет элемент в корзину
func (r *PostgresRepository) AddCartItem(ctx context.Context, item *models.CartItem) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}

	// В случае ошибки откатываем транзакцию
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Проверяем существование корзины и создаем если нет
	var exists bool
	err = tx.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM carts WHERE user_sso_id = $1)
	`, item.UserSSOID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("error checking cart existence: %w", err)
	}

	if !exists {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO carts (user_sso_id, updated_at)
			VALUES ($1, $2)
		`, item.UserSSOID, time.Now())
		if err != nil {
			return fmt.Errorf("error creating cart: %w", err)
		}
	}

	// Добавляем элемент в корзину
	var itemID int
	err = tx.QueryRowContext(ctx, `
		INSERT INTO cart_items (cart_id, user_sso_id, sneaker_id, quantity, added_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, item.UserSSOID, item.UserSSOID, item.SneakerID, item.Quantity, item.AddedAt, time.Now()).Scan(&itemID)
	if err != nil {
		return fmt.Errorf("error inserting cart item: %w", err)
	}

	// Обновляем время последнего изменения корзины
	_, err = tx.ExecContext(ctx, `
		UPDATE carts
		SET updated_at = $1
		WHERE user_sso_id = $2
	`, time.Now(), item.UserSSOID)
	if err != nil {
		return fmt.Errorf("error updating cart timestamp: %w", err)
	}

	// Завершаем транзакцию
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	// Обновляем ID элемента
	item.ID = fmt.Sprintf("%d", itemID)

	return nil
}

// UpdateCartItemQuantity обновляет количество элемента в корзине
func (r *PostgresRepository) UpdateCartItemQuantity(ctx context.Context, userSSOID int, itemID string, quantity int) error {
	// Обновляем количество элемента
	result, err := r.db.ExecContext(ctx, `
		UPDATE cart_items
		SET quantity = $1, updated_at = $2
		WHERE id = $3 AND cart_id = $4
	`, quantity, time.Now(), itemID, userSSOID)
	if err != nil {
		return fmt.Errorf("error updating cart item quantity: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("cart item not found")
	}

	// Обновляем время последнего изменения корзины
	_, err = r.db.ExecContext(ctx, `
		UPDATE carts
		SET updated_at = $1
		WHERE user_sso_id = $2
	`, time.Now(), userSSOID)
	if err != nil {
		return fmt.Errorf("error updating cart timestamp: %w", err)
	}

	return nil
}

// RemoveCartItem удаляет элемент из корзины
func (r *PostgresRepository) RemoveCartItem(ctx context.Context, userSSOID int, itemID string) error {
	// Удаляем элемент из корзины
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM cart_items
		WHERE id = $1 AND cart_id = $2
	`, itemID, userSSOID)
	if err != nil {
		return fmt.Errorf("error removing cart item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("cart item not found")
	}

	// Обновляем время последнего изменения корзины
	_, err = r.db.ExecContext(ctx, `
		UPDATE carts
		SET updated_at = $1
		WHERE user_sso_id = $2
	`, time.Now(), userSSOID)
	if err != nil {
		return fmt.Errorf("error updating cart timestamp: %w", err)
	}

	return nil
}

// ClearCart очищает корзину пользователя
func (r *PostgresRepository) ClearCart(ctx context.Context, userSSOID int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}

	// В случае ошибки откатываем транзакцию
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Удаляем все элементы корзины
	_, err = tx.ExecContext(ctx, `
		DELETE FROM cart_items
		WHERE cart_id = $1
	`, userSSOID)
	if err != nil {
		return fmt.Errorf("error clearing cart items: %w", err)
	}

	// Обновляем время последнего изменения корзины
	_, err = tx.ExecContext(ctx, `
		UPDATE carts
		SET updated_at = $1
		WHERE user_sso_id = $2
	`, time.Now(), userSSOID)
	if err != nil {
		return fmt.Errorf("error updating cart timestamp: %w", err)
	}

	// Завершаем транзакцию
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}
