package repository

import (
	"context"
	"errors"
	"loyalty-ledger/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresOrderRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresOrderRepository(pool *pgxpool.Pool) *PostgresOrderRepository {
	return &PostgresOrderRepository{pool: pool}
}

func (r *PostgresOrderRepository) CreateOrder(ctx context.Context, order *model.Order) error {
	query := `INSERT INTO orders (number, user_id, status, accrual, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	row := r.pool.QueryRow(ctx, query, order.Number, order.UserID, order.Status, order.Accrual, order.CreatedAt)
	err := row.Scan(&order.ID)
	if err != nil {
		var pgErr interface{ SQLState() string }
		if errors.As(err, &pgErr) && pgErr.SQLState() == "23505" {
			return model.ErrOrderExists
		}
		return err
	}
	return nil
}

func (r *PostgresOrderRepository) GetOrderByNumber(ctx context.Context, number string) (*model.Order, error) {
	order := &model.Order{}
	query := `SELECT id, number, user_id, status, accrual, created_at FROM orders WHERE number = $1`
	row := r.pool.QueryRow(ctx, query, number)
	err := row.Scan(&order.ID, &order.Number, &order.UserID, &order.Status, &order.Accrual, &order.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (r *PostgresOrderRepository) GetOrdersByUser(ctx context.Context, userID int64) ([]*model.Order, error) {
	query := `SELECT id, number, user_id, status, accrual, created_at FROM orders WHERE user_id = $1 ORDER BY created_at`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*model.Order
	for rows.Next() {
		order := &model.Order{}
		if err := rows.Scan(&order.ID, &order.Number, &order.UserID, &order.Status, &order.Accrual, &order.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, order)
	}
	return result, nil
}
