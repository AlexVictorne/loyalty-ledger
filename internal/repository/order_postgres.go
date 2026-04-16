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

func (r *PostgresOrderRepository) GetOrdersByStatus(ctx context.Context, statuses ...string) ([]string, error) {
	if len(statuses) == 0 {
		return nil, nil
	}
	// Build query with IN clause
	query := `SELECT number FROM orders WHERE status = ANY($1)`
	rows, err := r.pool.Query(ctx, query, statuses)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []string
	for rows.Next() {
		var number string
		if err := rows.Scan(&number); err != nil {
			return nil, err
		}
		result = append(result, number)
	}
	return result, nil
}

func (r *PostgresOrderRepository) UpdateOrderStatus(ctx context.Context, orderNumber string, status string, accrual *int64) error {
	var accrualVal interface{} = nil
	if accrual != nil {
		accrualVal = *accrual
	}
	query := `UPDATE orders SET status = $1, accrual = $2 WHERE number = $3`
	_, err := r.pool.Exec(ctx, query, status, accrualVal, orderNumber)
	return err
}

// BatchUpdateOrderStatus updates status for multiple orders if current status matches fromStatus.
// Returns list of actually updated order numbers.
func (r *PostgresOrderRepository) BatchUpdateOrderStatus(ctx context.Context, orderNumbers []string, fromStatus, toStatus string) ([]string, error) {
	if len(orderNumbers) == 0 {
		return nil, nil
	}
	// Update only orders with matching fromStatus, return updated numbers
	query := `UPDATE orders SET status = $1 WHERE number = ANY($2) AND status = $3 RETURNING number`
	rows, err := r.pool.Query(ctx, query, toStatus, orderNumbers, fromStatus)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var updated []string
	for rows.Next() {
		var number string
		if err := rows.Scan(&number); err != nil {
			return nil, err
		}
		updated = append(updated, number)
	}
	return updated, nil
}
