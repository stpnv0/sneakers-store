package sneakers

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type repository struct {
	db *sqlx.DB
}

// AddSneaker implements Repository.

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetSneakerByID(ctx context.Context, ID int64) (*Sneaker, error) {
	s := Sneaker{}
	query := "SELECT id, title, description, price, imageURL FROM sneakers WHERE id = $1"
	err := r.db.QueryRowContext(ctx, query, ID).Scan(&s.ID, &s.Title, &s.Description, &s.Price, &s.ImageUrl)
	if err != nil {
		return &Sneaker{}, nil
	}

	return &s, nil
}

func (r *repository) AddSneaker(ctx context.Context, sneaker *Sneaker) (*Sneaker, error) {
	query := "INSERT INTO sneakers (title, description, price, imageURL) VALUES ($1, $2, $3, $4) RETURNING id"
	err := r.db.QueryRowContext(ctx, query, sneaker.Title, sneaker.Description, sneaker.Price, sneaker.ImageUrl).Scan(&sneaker.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert sneaker: %w", err)
	}

	return sneaker, nil
}

func (r *repository) DeleteSneaker(ctx context.Context, id int64) error {
	query := "DELETE FROM sneakers WHERE id = $1"
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete sneaker: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to fetch rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("sneaker with id %d not found", id)
	}

	return nil
}

func (r *repository) GetAllSneakers(ctx context.Context) ([]*Sneaker, error) {
	query := "SELECT id, title, description, price, imageURL FROM sneakers"
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sneakers: %w", err)
	}
	defer rows.Close()

	var sneakers []*Sneaker
	for rows.Next() {
		var s Sneaker
		if err := rows.Scan(&s.ID, &s.Title, &s.Description, &s.Price, &s.ImageUrl); err != nil {
			return nil, fmt.Errorf("failed to scan sneaker row: %w", err)
		}
		sneakers = append(sneakers, &s)
	}

	return sneakers, nil
}

func (r *repository) GetSneakersByIDs(ctx context.Context, ids []int64) ([]*Sneaker, error) {
	query, args, err := sqlx.In("SELECT id, title, description, price, imageURL FROM sneakers WHERE id IN (?)", ids)
	if err != nil {
		return nil, fmt.Errorf("failed to create query: %w", err)
	}

	query = r.db.Rebind(query)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sneakers: %w", err)
	}
	defer rows.Close()

	var sneakers []*Sneaker
	for rows.Next() {
		var s Sneaker
		if err := rows.Scan(&s.ID, &s.Title, &s.Description, &s.Price, &s.ImageUrl); err != nil {
			return nil, fmt.Errorf("failed to scan sneaker row: %w", err)
		}
		sneakers = append(sneakers, &s)
	}

	return sneakers, nil
}
