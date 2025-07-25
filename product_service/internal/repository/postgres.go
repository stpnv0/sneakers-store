package repository

import (
	"context"
	"errors"
	"fmt"
	domain "product_service/internal/domain/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("entity not found")

type PostgresRepo struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *PostgresRepo {
	return &PostgresRepo{
		db: db,
	}
}

func (r *PostgresRepo) GetAllSneakers(ctx context.Context, limit, offset uint64) ([]*domain.Sneaker, error) {
	query := "SELECT id, title, price, image_key FROM sneakers ORDER BY id LIMIT $1 OFFSET $2"

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query sneakers: %w", err)
	}
	defer rows.Close()

	sneakers, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.Sneaker])
	if err != nil {
		return nil, fmt.Errorf("failed to collect sneaker rows: %w", err)
	}

	return sneakers, nil
}

func (r *PostgresRepo) AddSneaker(ctx context.Context, sneaker *domain.Sneaker) (int64, error) {
	query := "INSERT INTO sneakers (title, price, image_key) VALUES ($1, $2, $3) RETURNING id"

	var id int64
	err := r.db.QueryRow(ctx, query, sneaker.Title, sneaker.Price, sneaker.ImageKey).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to add sneaker: %w", err)
	}

	return id, nil
}

func (r *PostgresRepo) GetSneakerByID(ctx context.Context, id int64) (*domain.Sneaker, error) {
	query := "SELECT id, title, price, image_key from sneakers WHERE id = $1"

	row := r.db.QueryRow(ctx, query, id)

	var s domain.Sneaker
	err := row.Scan(&s.Id, &s.Title, &s.Price, &s.ImageKey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to scan sneaker row: %w", err)
	}

	return &s, nil
}

func (r *PostgresRepo) GetSneakersByIDs(ctx context.Context, ids []int64) ([]*domain.Sneaker, error) {
	if len(ids) == 0 {
		return []*domain.Sneaker{}, nil
	}
	query := "SELECT id, title, price, image_key from sneakers WHERE id = ANY($1)"

	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to query sneakers by ids: %w", err)
	}

	sneakers, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.Sneaker])
	if err != nil {
		return nil, fmt.Errorf("failed to collect sneaker rows: %w", err)
	}

	return sneakers, nil
}

func (r *PostgresRepo) DeleteSneaker(ctx context.Context, id int64) error {
	query := "DELETE FROM sneakers WHERE id = $1"

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete sneaker: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *PostgresRepo) UpdateImageKey(ctx context.Context, id int64, imageKey string) error {
	query := "UPDATE sneakers SET image_key = $1 WHERE id = $2"
	result, err := r.db.Exec(ctx, query, imageKey, id)
	if err != nil {
		return fmt.Errorf("failed to update sneaker image key: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
