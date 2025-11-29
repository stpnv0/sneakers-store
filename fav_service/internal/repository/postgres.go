package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"fav_service/internal/models"

	_ "github.com/lib/pq"
)

type PostgresRepo struct {
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

func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{db: db}
}

func (p *PostgresRepo) AddToFavourite(ctx context.Context, userSSOID, sneakerID int) error {
	query := `INSERT INTO favourites_items (user_sso_id, sneaker_id) VALUES ($1, $2) ON CONFLICT (user_sso_id, sneaker_id) DO NOTHING`
	_, err := p.db.ExecContext(ctx, query, userSSOID, sneakerID)
	if err != nil {
		return fmt.Errorf("failed to add to favourites: %w", err)
	}
	return nil
}

func (p *PostgresRepo) RemoveFromFavourite(ctx context.Context, userSSOID, sneakerID int) error {
	query := `DELETE FROM favourites_items WHERE user_sso_id = $1 AND sneaker_id = $2`
	result, err := p.db.ExecContext(ctx, query, userSSOID, sneakerID)
	if err != nil {
		return fmt.Errorf("failed to remove from favourites: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("favourite not found")
	}

	return nil
}

func (p *PostgresRepo) GetAllFavourites(ctx context.Context, userSSOID int) ([]models.Favourite, error) {
	query := `SELECT id, user_sso_id, sneaker_id, created_at FROM favourites_items WHERE user_sso_id = $1`
	rows, err := p.db.QueryContext(ctx, query, userSSOID)
	if err != nil {
		return nil, fmt.Errorf("failed to get favourites: %w", err)
	}
	defer rows.Close()

	var favourites []models.Favourite
	for rows.Next() {
		var item models.Favourite
		if err := rows.Scan(&item.ID, &item.UserSSOID, &item.SneakerID, &item.AddedAt); err != nil {
			return nil, fmt.Errorf("failed to scan favourite: %w", err)
		}
		favourites = append(favourites, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating favourites: %w", err)
	}

	return favourites, nil
}

func (p *PostgresRepo) IsFavourite(ctx context.Context, userSSOID, sneakerID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM favourites_items WHERE user_sso_id = $1 AND sneaker_id = $2)`
	var exists bool
	err := p.db.QueryRowContext(ctx, query, userSSOID, sneakerID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if favourite exists: %w", err)
	}
	return exists, nil
}
